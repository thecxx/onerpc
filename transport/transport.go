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
	"time"
)

const (
	WeightNormal int = 100
)

var (
	ErrNoIdleServerFound = errors.New("no idle server found")
)

type Await struct {
	out  Message
	err  error
	done chan struct{}
}

// Done
func (a *Await) Done(out Message, err error) {
	a.out = out
	a.err = err
	// Done signal
	close(a.done)
}

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
	ap, pp, mp sync.Pool
	// Workers
	workers map[*Line]time.Time
	// Pending
	pendings sync.Map
	seqincr  uint64
	ctx      context.Context
	cancel   context.CancelFunc
	locker   sync.RWMutex
}

// Start
func (t *Transport) Start(ctx context.Context) {
	t.ctx, t.cancel = context.WithCancel(ctx)
	t.workers = make(map[*Line]time.Time)
	// Await pool
	t.ap.New = func() interface{} { return new(Await) }
	// Packet pool
	t.pp.New = func() interface{} { return new(Packet) }
	// Message pool
	t.mp.New = func() interface{} { return t.Proto.NewMessage() }

	// Background goroutine
	go t.background()
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
	return t.write(ctx, message)
}

// Broadcast
func (t *Transport) Broadcast(ctx context.Context, message []byte) (err error) {
	t.locker.RLock()
	defer t.locker.RUnlock()
	//
	m := t.mp.Get().(Message)
	m.SetOneway()
	m.Store(message)
	m.SetSeq(t.seq())

	defer func() {
		t.mp.Put(m)
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
	l.mp = &t.mp

	// Options
	l.rt = t.ReadTimeout
	l.wt = t.WriteTimeout
	l.it = t.IdleTimeout
	l.lt = t.MaxLifeTime

	l.Start(t.ctx)

	// Insert new worker
	t.insertWorker(l, weight)
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

// background
func (t *Transport) background() {
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
