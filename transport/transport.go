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
	Dial() (conn net.Conn, weight int, hang chan struct{}, err error)
	Hang(conn net.Conn)
}

// New packet
type PacketFunc func() Packet

// Event handler
type EventHandler interface {
	OnPacket(ctx context.Context, rp Packet) (sp Packet, err error)
}

type Balancer interface {
	Add(c *Connection, weight int)
	Next() (c *Connection)
	Remove(c *Connection)
}

type Transport struct {
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
	MaxLifeTime      time.Duration
	ReaderBufferSize int
	WriterBufferSize int
	Dialer           Dialer
	Packet           PacketFunc
	Event            EventHandler
	Balancer         Balancer

	workers map[*Connection]time.Time

	// Queues
	rqueue chan *Receiver
	dqueue chan *Connection

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
	t.workers = make(map[*Connection]time.Time)
	t.rqueue = make(chan *Receiver, 1024)
	// Handle foreign packet
	go t.handleForeignPacket()
}

// Stop
func (t *Transport) Stop() {
	for c, _ := range t.workers {
		c.Stop()
	}
	t.cancel()
}

// Broadcast
func (t *Transport) Broadcast(ctx context.Context, sp Packet) (err error) {
	sp.SetOneway()
	return t.broadcast(ctx, sp)
}

// Send
func (t *Transport) Send(ctx context.Context, sp Packet) (rp Packet, err error) {
	return t.send(ctx, sp, nil)
}

// Async
func (t *Transport) Async(ctx context.Context, sp Packet, fn func(rp Packet, err error)) {
	t.send(ctx, sp, fn)
}

// Join
func (t *Transport) Join(conn net.Conn, weight int, hang chan struct{}) {
	t.join(conn, weight, hang)
}

// join
func (t *Transport) join(conn net.Conn, weight int, hang chan struct{}) {
	c := new(Connection)

	c.Weight = weight

	c.Hang = hang
	// Connection and Packet
	c.Conn = conn
	c.Packet = t.Packet

	// Options
	c.IdleTimeout = t.IdleTimeout
	c.ReadTimeout = t.ReadTimeout
	c.WriteTimeout = t.WriteTimeout
	c.MaxLifeTime = t.MaxLifeTime
	c.ReaderBufferSize = t.ReaderBufferSize
	c.WriterBufferSize = t.WriterBufferSize

	// Queue
	c.DQueue = t.dqueue
	c.RQueue = t.rqueue

	t.locker.Lock()
	t.workers[c] = time.Now()
	if t.Balancer != nil {
		t.Balancer.Add(c, weight)
	}
	t.locker.Unlock()

	c.Start(t.ctx)
}

// discard
func (t *Transport) discard(c *Connection) {
	t.locker.Lock()
	defer t.locker.Unlock()
	// Delete
	delete(t.workers, c)
	if t.Balancer != nil {
		t.Balancer.Remove(c)
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
		t.join(conn, weight, hang)
	}
}

// hang
func (t *Transport) hang(c *Connection) {
	if t.Dialer == nil {
		return
	}
	t.Dialer.Hang(c.Conn)
}

// seq
func (t *Transport) seq() (seq uint64) {
	if seq = atomic.AddUint64(&t.sequence, 1); seq == 0 {
		seq = atomic.AddUint64(&t.sequence, 1)
	}
	return
}

// broadcast
func (t *Transport) broadcast(ctx context.Context, sp Packet) (err error) {
	t.locker.RLock()
	defer t.locker.RUnlock()
	// Broadcast
	for c, _ := range t.workers {
		c.Send(sp)
	}
	return
}

// send
func (t *Transport) send(ctx context.Context, sp Packet, fn func(rp Packet, err error)) (rp Packet, err error) {

	seq := t.seq()

	// Sequence number
	sp.SetSeq(seq)

	sender := new(Sender)
	sender.SP = sp

	// Done signal
	if fn != nil {
		sender.OnRecv = fn
	} else {
		sender.Done = make(chan *Sender, 10)
	}

	// Twoway
	if !sp.IsOneway() {
		t.pendings.Store(seq, sender)
	}

	// Send packet
	go t.gogogo(sender)

	// Async
	if sender.OnRecv != nil {
		return
	}

	select {
	// Cancel
	case <-ctx.Done():
		return
	// OK
	case <-sender.Done:
		rp = sender.RP
	}

	if sender.Err != nil {
		rp, err = nil, sender.Err
	}

	return
}

// gogogo
func (t *Transport) gogogo(sender *Sender) {
	sp := sender.SP

	if c := t.selectWorker(); c == nil {
		sender.Ack(nil, errors.New("no worker found"))
	} else {
		if _, err := c.Send(sp); err != nil {
			sender.Ack(nil, err)
		}
		// Oneway
		if sp.IsOneway() {
			sender.Ack(nil, nil)
		}
	}
}

// onRecv
func (t *Transport) onRecv(receiver *Receiver) {

	seq := receiver.RP.Seq()

	// Find sender
	sender, ok := t.findSender(seq)
	if !ok {
		t.noSender(receiver)
	} else {
		sender.Ack(receiver.RP, nil)
	}
}

// noSender
func (t *Transport) noSender(receiver *Receiver) {
	if t.Event == nil {
		return
	}

	rp := receiver.RP

	sp, err := t.Event.OnPacket(t.ctx, rp)
	if err != nil {
		return
	}

	// Oneway
	if rp.IsOneway() || sp == nil {
		return
	}

	// Sequence number
	sp.SetSeq(rp.Seq())
	// Ack
	receiver.OnAck(sp)
}

// selectWorker
func (t *Transport) selectWorker() (c *Connection) {
	t.locker.RLock()
	defer t.locker.RUnlock()
	// Select a worker connection
	if t.Balancer != nil {
		c = t.Balancer.Next()
	}
	return
}

// findSender
func (t *Transport) findSender(seq uint64) (sender *Sender, ok bool) {
	value, ok := t.pendings.LoadAndDelete(seq)
	if ok {
		sender = value.(*Sender)
	}
	return
}

// handleForeignPacket
func (t *Transport) handleForeignPacket() {
	tk := time.NewTicker(time.Second)
	defer tk.Stop()

	for {
		select {
		// Cancel
		case <-t.ctx.Done():
			return
		// Recv queue
		case r := <-t.rqueue:
			t.onRecv(r)
		// Disconnect queue
		case c := <-t.dqueue:
			t.discard(c)
			t.hang(c)
		// Join new connection
		case <-tk.C:
			t.tryDial()
		}
	}
}
