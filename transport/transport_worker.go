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
	"sync/atomic"
)

// Send
func (t *Transport) Send(ctx context.Context, message []byte, opts ...MessageOption) (reply []byte, err error) {

	seq := t.seq()

	m := t.mp.Get().(Message)

	// Set options
	for _, setOpt := range opts {
		setOpt(m)
	}

	m.Store(message)
	m.SetSeq(seq)

	a := t.ap.Get().(*await)
	a.ret = nil
	a.err = nil

	// Done signal
	a.done = make(chan struct{}, 0)

	// TODO 删除Pendings

	// Twoway
	if !m.IsOneway() {
		t.regAwait(seq, a)
	}

	// Send message
	go t.gogo(a, m)

	select {
	// Cancel
	case <-ctx.Done():
		return nil, ctx.Err()
	// Done
	case <-a.done:
		if a.err != nil {
			reply, err = nil, a.err
		} else if a.ret != nil {
			reply = a.ret.Bytes()
		}
	}

	return
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
	t.bundler.Walk(func(l *Line) {
		l.Write(m)
	})
	return
}

// seq
func (t *Transport) seq() (seq uint64) {
	return atomic.AddUint64(&t.seqincr, 1)
}

// gogo
func (t *Transport) gogo(a *await, m Message) {
	if l := t.Balancer.Next(); l == nil {
		a.Done(nil, errors.New("no worker found"))
	} else {
		if _, err := l.Write(m); err != nil {
			a.Done(nil, err)
		} else if m.IsOneway() {
			a.Done(nil, nil)
		}
	}
}

// regAwait
func (t *Transport) regAwait(seq uint64, a *await) {
	t.pendings.Store(seq, a)
}

// findAwait
func (t *Transport) findAwait(seq uint64) (a *await) {
	if value, ok := t.pendings.LoadAndDelete(seq); ok {
		a = value.(*await)
	}
	return
}

// OnTraffic
func (t *Transport) OnTraffic(l *Line, m Message) {

	seq := m.Seq()

	// Find await
	if a := t.findAwait(seq); a != nil {
		a.Done(m, nil)
		return
	}

	p := t.pp.Get().(*Packet)
	p.reset()
	p.ctx = t.ctx
	p.line = l
	p.message = m
	p.proto = t.Proto
	p.replied = false

	defer func() {
		t.pp.Put(p)
	}()

	w := messageWriter{
		packet: p,
		line:   l,
		mp:     &t.mp,
	}

	t.Handler.ServePacket(w, p)
}

// Disconnect
func (t *Transport) OnDisconnect(l *Line) {
	t.bundler.Remove(l)
	if t.Balancer != nil {
		t.Balancer.Remove(l)
	}
	if t.Dialer != nil {
		t.Dialer.Hang(l.conn)
	}
}
