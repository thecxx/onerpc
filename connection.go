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
	"io"
	"net"
	"sync"
	"sync/atomic"
)

var (
	ErrConnectionHangUp = errors.New("hang up")
)

type Connection struct {
	Conn net.Conn
	// Protocol
	Proto Protocol
	// Hang up
	Hang <-chan struct{}
	// Others
	running uint32
	locker  sync.Mutex
}

// Run
func (c *Connection) Run(ctx context.Context,
	fn func(c *Connection, message Message)) (err error) {

	if fn == nil {
		return errors.New("invalid message handler")
	}

	if c.Proto == nil {
		return errors.New("invalid message protocol")
	}

	atomic.StoreUint32(&c.running, 1)
	defer atomic.StoreUint32(&c.running, 0)

	var r = c.Conn
	var m Message

	for {
		// Scan message
		if m, err = c.scan(r); err != nil {
			break
		}

		// Handle message
		fn(c, m)

		select {
		// Cancel
		case <-ctx.Done():
			err = ctx.Err()
		// Hang up
		case <-c.Hang:
			err = ErrConnectionHangUp
		}

		if err != nil {
			break
		}
	}

	return
}

// Send
func (c *Connection) Send(ctx context.Context, message Message) (err error) {
	if atomic.LoadUint32(&c.running) == 0 {
		return errors.New("connection is not running")
	}
	c.locker.Lock()
	defer c.locker.Unlock()
	// Send
	_, err = message.WriteTo(c.Conn)
	return
}

// scan
func (c *Connection) scan(r io.Reader) (message Message, err error) {
	message = c.Proto.NewMessage()
	// Read message
	_, err = message.ReadFrom(r)
	return
}
