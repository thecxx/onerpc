// Copyright 2023 Kami
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transport

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	WeightNormal int = 100
)

var (
	ErrNoIdleServerFound = errors.New("no idle server found")
)

// Listener
type Listener interface {
	Listen() (ln net.Listener, err error)
}

// Dialer
type Dialer interface {
	Dial() (conn net.Conn, weight int, hang <-chan struct{}, err error)
	Hang(conn net.Conn)
}

type Balancer interface {
	Add(l *Line, weight int)
	Next() (l *Line)
	Remove(l *Line)
}

type Transport struct {
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
	MaxLifeTime      time.Duration
	ReaderBufferSize int
	WriterBufferSize int
	Dialer           Dialer
	Proto            Protocol
	Balancer         Balancer
	Handler          Handler

	// Packet pool
	ppool   sync.Pool
	workers map[*Line]time.Time

	// Queues
	rqueue chan *Packet
	dqueue chan *Line

	// Pending
	pendings sync.Map
	sequence uint64

	ctx    context.Context
	cancel context.CancelFunc
	locker sync.RWMutex
}

// Start
func (t *Transport) Start(ctx context.Context) {
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.workers = make(map[*Line]time.Time)
	t.dqueue = make(chan *Line, 1024)
	t.ppool.New = func() interface{} { return new(Packet) }
	// Handle remote packet
	go t.handleRemotePacket()
}

// Stop
func (t *Transport) Stop() {
	for c, _ := range t.workers {
		c.Stop()
	}
	t.cancel()
}

// Broadcast
func (t *Transport) Broadcast(ctx context.Context, m Message) (err error) {
	m.SetOneway()
	return t.broadcast(ctx, m)
}

// Send
func (t *Transport) Send(ctx context.Context, m Message) (r Message, err error) {
	return t.send(ctx, m, nil)
}

// // Async
// func (t *Transport) Async(ctx context.Context, sp Packet, fn func(rp Packet, err error)) {
// 	t.send(ctx, sp, fn)
// }

// Join
func (t *Transport) Join(conn net.Conn, weight int, hang <-chan struct{}) {
	t.join(conn, weight, hang)
}

// join
func (t *Transport) join(conn net.Conn, weight int, hang <-chan struct{}) {

	l := new(Line)

	// Protocol
	l.proto = t.Proto

	// Connection
	l.conn = conn
	l.hang = hang
	l.handler = t

	// Queues
	l.dqueue = t.dqueue

	// Options
	l.rt = t.ReadTimeout
	l.wt = t.WriteTimeout
	l.it = t.IdleTimeout
	l.lt = t.MaxLifeTime

	// Insert new worker
	t.insertWorker(l, weight)

	l.Start(t.ctx)
}

// disconnect
func (t *Transport) disconnect(l *Line) {
	t.removeWorker(l)
	t.hang(l)
}

// ServeMessage
func (t *Transport) ServeMessage(l *Line, m Message) {

	seq := m.Seq()

	// Find sender
	sender, ok := t.findSender(seq)
	if ok {
		sender.Ack(m, nil)
		return
	}

	p := t.ppool.New().(*Packet)
	p.ctx = t.ctx
	p.proto = l.proto
	p.line = l
	p.message = m

	defer func() {
		t.ppool.Put(p)
	}()

	w := messageWriter{
		m: m,
		l: l,
		t: t,
	}

	t.Handler.ServePacket(w, p)
}

// tryDial
func (t *Transport) tryDial() {
	if t.Dialer == nil {
		return
	}
	conn, weight, hang, err := t.Dialer.Dial()
	if err != nil {
		// TODO
	} else {
		t.join(conn, weight, hang)
	}
}

// hang
func (t *Transport) hang(l *Line) {
	if t.Dialer != nil {
		t.Dialer.Hang(l.conn)
	}
}

// seq
func (t *Transport) seq() (seq uint64) {
	if seq = atomic.AddUint64(&t.sequence, 1); seq == 0 {
		seq = atomic.AddUint64(&t.sequence, 1)
	}
	return
}

// broadcast
func (t *Transport) broadcast(ctx context.Context, m Message) (err error) {
	t.locker.RLock()
	defer t.locker.RUnlock()
	// Broadcast
	for c, _ := range t.workers {
		c.Send(m)
	}
	return
}

// send
func (t *Transport) send(ctx context.Context, m Message, fn func(r Message, err error)) (r Message, err error) {

	seq := t.seq()

	// Sequence number
	m.SetSeq(seq)

	sender := new(Sender)
	sender.message = m

	// Done signal
	if fn != nil {
		sender.handler = fn
	} else {
		sender.done = make(chan struct{}, 0)
	}

	// autoRemovePendings := false

	// Twoway
	if !m.IsOneway() {
		t.insertSender(seq, sender)
		// if autoRemovePendings {
		// 	t.pendings.Delete(seq)
		// }
	}

	// Send packet
	go t.gogo(sender)

	// Async
	if sender.handler != nil {
		return
	}

	select {
	// Cancel
	case <-ctx.Done():
		// autoRemovePendings = true
		err = ctx.Err()
	// Done
	case <-sender.done:
		r = sender.reply
	}

	if sender.err != nil {
		r, err = nil, sender.err
	}

	return
}

// gogo
func (t *Transport) gogo(sender *Sender) {
	if l := t.selectWorker(); l == nil {
		sender.Ack(nil, errors.New("no worker found"))
	} else {
		if _, err := l.Send(sender.message); err != nil {
			sender.Ack(nil, err)
		} else if sender.message.IsOneway() {
			sender.Ack(nil, nil)
		}
	}
}

// findSender
func (t *Transport) findSender(seq uint64) (sender *Sender, ok bool) {
	value, ok := t.pendings.LoadAndDelete(seq)
	if ok {
		sender = value.(*Sender)
	}
	return
}

// insertSender
func (t *Transport) insertSender(seq uint64, sender *Sender) {
	t.pendings.Store(seq, sender)
}

// selectWorker
func (t *Transport) selectWorker() (l *Line) {
	t.locker.RLock()
	defer t.locker.RUnlock()
	// Select a worker
	if t.Balancer != nil {
		l = t.Balancer.Next()
	}
	return
}

// insertWorker
func (t *Transport) insertWorker(l *Line, weight int) {
	t.locker.Lock()
	defer t.locker.Unlock()
	// Insert new worker
	t.workers[l] = time.Now()
	if t.Balancer != nil {
		t.Balancer.Add(l, weight)
	}
}

// removeWorker
func (t *Transport) removeWorker(l *Line) {
	t.locker.Lock()
	defer t.locker.Unlock()
	// Delete a worker
	delete(t.workers, l)
	if t.Balancer != nil {
		t.Balancer.Remove(l)
	}
}

// handleRemotePacket
func (t *Transport) handleRemotePacket() {
	tk := time.NewTicker(time.Second)
	defer tk.Stop()

	for {
		select {
		// Cancel
		case <-t.ctx.Done():
			return
		// Join new connection
		case <-tk.C:
			t.tryDial()
		// Disconnect queue
		case c := <-t.dqueue:
			t.disconnect(c)
		}
	}
}

type messageWriter struct {
	m    Message
	l    *Line
	t    *Transport
	sent bool
}

// WriteMessage implements MessageWriter.
func (w messageWriter) WriteMessage(m Message) (n int64, err error) {
	if m.Seq() == 0 {
		m.SetSeq(w.t.seq())
	}
	// Reply
	if w.m.Seq() == m.Seq() {
		if w.m.IsOneway() || w.sent {
			return
		}
		defer func() {
			w.sent = true
		}()
	}
	return w.l.Send(m)
}
