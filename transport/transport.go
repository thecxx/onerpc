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

type await struct {
	// Return message
	ret Message
	err error
	// Done signal
	done chan struct{}
}

// Done
func (a *await) Done(ret Message, err error) {
	a.ret = ret
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
	// Bundler
	bundler Bundler
	// Pools
	ap, pp, mp sync.Pool
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
	// Bundler
	t.bundler.lines = make(map[*Line]time.Time)
	// await pool
	t.ap.New = func() interface{} { return new(await) }
	// Packet pool
	t.pp.New = func() interface{} { return new(Packet) }
	// Message pool
	t.mp.New = func() interface{} { return t.Proto.NewMessage() }
	// Background goroutine
	if t.Dialer != nil {
		go t.background()
	}
}

// Stop
func (t *Transport) Stop() {
	t.bundler.Walk(func(l *Line) {
		l.Stop()
	})
	t.cancel()
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

	// Add new line
	t.bundler.Add(l)
	if t.Balancer != nil {
		t.Balancer.Add(l, weight)
	}
}

// dial
func (t *Transport) dial() (err error) {
	if t.Dialer == nil {
		return errors.New("no dialer found")
	}
	conn, weight, hang, err := t.Dialer.Dial()
	if err != nil {
		if err == ErrNoIdleServerFound {
			err = nil
		}
		return
	}
	// Join
	t.Join(conn, weight, hang)
	return
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
			_ = t.dial()
		}
	}
}
