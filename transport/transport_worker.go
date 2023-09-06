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
	"time"
)

// seq
func (t *Transport) seq() (seq uint64) {
	return atomic.AddUint64(&t.seqincr, 1)
}

// write
func (t *Transport) write(ctx context.Context, in []byte, opts ...MessageOption) (out []byte, err error) {

	seq := t.seq()

	m := t.mp.Get().(Message)

	// Set options
	for _, setOpt := range opts {
		setOpt(m)
	}

	m.Store(in)

	// Sequence number
	m.SetSeq(seq)

	a := t.ap.Get().(*Await)

	// Done signal
	a.done = make(chan struct{}, 0)

	// TODO 删除Pendings

	// Twoway
	if !m.IsOneway() {
		t.registerAwait(seq, a)
	}

	// Send message
	go t.gogogo(a, m)

	select {
	// Cancel
	case <-ctx.Done():
		return nil, ctx.Err()
	// Done
	case <-a.done:
		if a.err != nil {
			out, err = nil, a.err
		} else if a.out != nil {
			out = a.out.Bytes()
		}
	}

	return
}

// gogogo
func (t *Transport) gogogo(a *Await, m Message) {
	if l := t.selectWorker(); l == nil {
		a.Done(nil, errors.New("no worker found"))
	} else {
		if _, err := l.Write(m); err != nil {
			a.Done(nil, err)
		} else if m.IsOneway() {
			a.Done(nil, nil)
		}
	}
}

// searchAwait
func (t *Transport) searchAwait(seq uint64) (a *Await, ok bool) {
	value, ok := t.pendings.LoadAndDelete(seq)
	if ok {
		a = value.(*Await)
	}
	return
}

// registerAwait
func (t *Transport) registerAwait(seq uint64, a *Await) {
	t.pendings.Store(seq, a)
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

// OnTraffic
func (t *Transport) OnTraffic(l *Line, m Message) {

	seq := m.Seq()

	// Find await
	a, ok := t.searchAwait(seq)
	if ok {
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
	t.removeWorker(l)
	// Hang
	if t.Dialer != nil {
		t.Dialer.Hang(l.conn)
	}
}
