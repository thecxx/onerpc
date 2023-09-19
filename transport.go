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

package onerpc

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"time"
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
	Next() (c *Connection)
	Add(c *Connection, weight int)
	Remove(c *Connection)
}

type Transport struct {
	// Options
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
	MaxLifeTime      time.Duration
	ReaderBufferSize int
	WriterBufferSize int
	//
	Dialer   Dialer
	Proto    Protocol
	Handler  Handler
	Balancer Balancer
	Ctx      context.Context
	//
	cgroup CGroup
	// awaits  sync.Map
	pendings pendings
	seqincr  uint64
}

// Broadcast
func (t *Transport) Broadcast(ctx context.Context, message []byte) (err error) { return }

// Send
func (t *Transport) Send(ctx context.Context, message []byte) (reply []byte, err error) {

	// Local message
	m := t.wrapm(message)

	// Push message
	go t.pushm(m)

	if m.IsOneway() {
		return nil, nil
	}

	r, err := t.pendings.Wait(ctx, m.Seq())

	// Load bytes
	if r != nil {
		reply = r.Bytes()
	}

	return
}

// pushm
func (t *Transport) pushm(m Message) {

	seq := m.Seq()

	c := t.Balancer.Next()
	if c == nil {
		t.pendings.Trigger(seq, nil, errors.New("no worker found"))
		return
	}

	if _, err := c.Send(m); err != nil {
		t.pendings.Trigger(seq, nil, err)
	} else if m.IsOneway() {
		t.pendings.Trigger(seq, nil, nil)
	}
}

// seq
func (t *Transport) seq() uint64 {
	return atomic.AddUint64(&t.seqincr, 1)
}

// wrapm
func (t *Transport) wrapm(b []byte) (m Message) {
	m = t.Proto.NewMessage()
	// Fill message
	m.Store(b)
	m.SetSeq(t.seq())
	return
}

// Join
func (t *Transport) Join(conn net.Conn, weight int, hang <-chan struct{}) {

	c := new(Connection)

	// Connection
	c.raw = conn
	c.hang = hang

	// Protocol
	c.proto = t.Proto

	// Options
	c.rt = t.ReadTimeout
	c.wt = t.WriteTimeout
	c.it = t.IdleTimeout
	c.lt = t.MaxLifeTime

	// Add worker
	t.cgroup.Add(c)
	// Add balancer
	if t.Balancer != nil {
		t.Balancer.Add(c, weight)
	}

	go func() {
		// Bootstrap
		err := c.Run(t.Ctx, func(m Message) { t.handleMessage(c, m) })
		if err != nil {

		}
		// Remove worker
		t.cgroup.Remove(c)
		// Remove balancer
		if t.Balancer != nil {
			t.Balancer.Remove(c)
		}
	}()
}

// handleMessage
func (t *Transport) handleMessage(c *Connection, m Message) {

	seq := m.Seq()

	if v, ok := t.awaits.LoadAndDelete(seq); ok {
		v.(chan interface{}) <- m
		return
	}

	p := &Packet{
		ctx:     t.Ctx,
		conn:    c,
		message: m,
		proto:   t.Proto,
		replied: false,
	}

	w := messageWriter{
		conn:   c,
		packet: p,
		proto:  t.Proto,
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
		t.Join(conn, weight, hang)
	}
}
