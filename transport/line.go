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

type LineEventHandler interface {
	OnTraffic(l *Line, m Message)
	OnDisconnect(l *Line)
}

type Line struct {
	// Raw connection
	conn net.Conn
	// Hang connection
	hang <-chan struct{}
	// Handler
	handler LineEventHandler
	// Pool
	mpool *sync.Pool
	// Options
	rt time.Duration
	wt time.Duration
	it time.Duration
	lt time.Duration
	// error
	err    error
	ctx    context.Context
	cancel context.CancelFunc
	locker sync.Mutex
}

// Start
func (l *Line) Start(ctx context.Context) {
	if l.lt <= 0 {
		l.ctx, l.cancel = context.WithCancel(ctx)
	} else {
		l.ctx, l.cancel = context.WithTimeout(ctx, l.lt)
	}
	// Recv from remote
	go l.recv()
}

// Stop
func (l *Line) Stop() {
	l.cancel()
}

// Write
func (l *Line) Write(m Message) (n int64, err error) {
	l.locker.Lock()
	defer l.locker.Unlock()
	// Write to remote
	return m.WriteTo(l.conn)
}

// recv
func (l *Line) recv() {

	r := l.conn

	for {
		if l.err != nil {
			break
		}

		m := l.mpool.Get().(Message)
		m.Reset()

		if l.it > 0 {
			l.conn.SetReadDeadline(time.Now().Add(l.it))
		}

		// Recv from connection
		if _, l.err = m.ReadFrom(r); l.err != nil {
			break
		}

		// Handle message
		l.handler.OnTraffic(l, m)

		select {
		// Cancel
		case <-l.ctx.Done():
			l.err = l.ctx.Err()
			break
		// Hang
		case <-l.hang:
			l.err = errors.New("hang up")
			break
		default:
			// Nothing to do
		}
	}

	// Disconnect
	l.conn.Close()

	// TODO 未处理完的数据包
	l.handler.OnDisconnect(l)
}
