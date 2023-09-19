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
	"sync"
	"time"
)

type Connection struct {
	// Raw connection
	raw net.Conn
	// Hang connection
	hang <-chan struct{}
	// Protocol
	proto Protocol
	// Options
	rt time.Duration
	wt time.Duration
	it time.Duration
	lt time.Duration
	//
	locker sync.Mutex
}

// Run
func (c *Connection) Run(ctx context.Context, fn func(Message)) (err error) {
	if fn == nil {
		return errors.New("invalid message handler")
	}
	return c.recv(ctx, fn)
}

// Send
func (c *Connection) Send(m Message) (n int64, err error) {
	c.locker.Lock()
	defer c.locker.Unlock()
	// Write to remote
	return m.WriteTo(c.raw)
}

// recv
func (c *Connection) recv(ctx context.Context, fn func(Message)) (err error) {
	var (
		r = c.raw
		m = c.proto.NewMessage()
	)

	for {
		if err != nil {
			break
		}
		// TODO Reset message

		if c.it > 0 {
			c.raw.SetReadDeadline(time.Now().Add(c.it))
		}

		// Recv from line
		if _, err = m.ReadFrom(r); err != nil {
			break
		}

		// Handle message
		fn(m)

		select {
		// Cancel
		case <-ctx.Done():
			err = ctx.Err()
			break
		// Hang
		case <-c.hang:
			err = errors.New("hang up")
			break
		default:
			// Nothing to do
		}
	}

	// Disconnect
	c.raw.Close()

	return
}
