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
	// Options
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	MaxLifeTime  time.Duration
	// Buffer sizes
	ReaderBufferSize int
	WriterBufferSize int
	//
	Dialer   Dialer
	Proto    Protocol
	Handler  Handler
	Balancer Balancer
	// Pool
	spool sync.Pool
	ppool sync.Pool
	mpool sync.Pool
	// Workers
	workers map[*Line]time.Time
	// Pending
	pendings sync.Map
	sequence uint64
	ctx      context.Context
	cancel   context.CancelFunc
	locker   sync.RWMutex
}

// Start
func (t *Transport) Start(ctx context.Context) {
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.workers = make(map[*Line]time.Time)
	t.spool.New = func() interface{} { return new(Sender) }
	t.ppool.New = func() interface{} { return new(Packet) }
	t.mpool.New = func() interface{} { return t.Proto.NewMessage() }
	// Handle remote packet
	go t.handleRemotePacket()
}

// Stop
func (t *Transport) Stop() {
	for c := range t.workers {
		c.Stop()
	}
	t.cancel()
}

// Send
func (t *Transport) Send(ctx context.Context, message []byte) (reply []byte, err error) {
	return t.send(ctx, message, nil)
}

// Async
func (t *Transport) Async(ctx context.Context, message []byte, fn func(reply []byte, err error)) {
	t.send(ctx, message, fn)
}

// Broadcast
func (t *Transport) Broadcast(ctx context.Context, message []byte) (err error) {
	t.locker.RLock()
	defer t.locker.RUnlock()
	//
	m := t.mpool.Get().(Message)
	m.Reset()
	m.SetOneway()
	m.Store(message)

	defer func() {
		t.mpool.Put(m)
	}()

	// Broadcast
	for l := range t.workers {
		l.Write(m)
	}
	return
}

// Join
func (t *Transport) Join(conn net.Conn, weight int, hang <-chan struct{}) {

	l := new(Line)

	// Connection
	l.conn = conn
	l.hang = hang
	l.handler = t

	// Pool
	l.mpool = &t.mpool

	// Options
	l.rt = t.ReadTimeout
	l.wt = t.WriteTimeout
	l.it = t.IdleTimeout
	l.lt = t.MaxLifeTime

	// Insert new worker
	t.insertWorker(l, weight)

	l.Start(t.ctx)
}

// OnTraffic
func (t *Transport) OnTraffic(l *Line, m Message) {

	seq := m.Seq()

	// Find sender
	sender, ok := t.findSender(seq)
	if ok {
		sender.Ack(m, nil)
		return
	}

	p := t.ppool.Get().(*Packet)
	p.reset()
	p.ctx = t.ctx
	p.line = l
	p.message = m
	p.proto = t.Proto
	p.replied = false

	defer func() {
		t.ppool.Put(p)
	}()

	w := messageWriter{
		packet: p,
		line:   l,
		mpool:  &t.mpool,
	}

	t.Handler.ServePacket(w, p)
}

// Disconnect
func (t *Transport) OnDisconnect(l *Line) {
	t.removeWorker(l)
	// Hang
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

// send
func (t *Transport) send(ctx context.Context, message []byte, fn func(reply []byte, err error)) (reply []byte, err error) {

	seq := t.seq()

	m := t.mpool.Get().(Message)
	m.Reset()

	// Sequence number
	m.SetSeq(seq)

	sender := t.spool.Get().(*Sender)
	sender.reset()
	sender.message = m

	// Done signal
	if fn != nil {
		sender.handler = fn
	} else {
		sender.done = make(chan struct{}, 0)
	}

	// TODO 删除Pendings

	// Twoway
	if !m.IsOneway() {
		t.insertSender(seq, sender)
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
		return nil, ctx.Err()
	// Done
	case <-sender.done:
		if sender.err != nil {
			reply, err = nil, sender.err
		} else if sender.reply != nil {
			reply = sender.reply.Bytes()
		}
	}

	return
}

// gogo
func (t *Transport) gogo(sender *Sender) {
	if l := t.selectWorker(); l == nil {
		sender.Ack(nil, errors.New("no worker found"))
	} else {
		if _, err := l.Write(sender.message); err != nil {
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

// tryDial
func (t *Transport) tryDial() {
	if t.Dialer == nil {
		return
	}
	conn, weight, hang, err := t.Dialer.Dial()
	if err != nil {
		// TODO
	} else {
		t.Join(conn, weight, hang)
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
		}
	}
}
