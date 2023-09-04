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
	"bufio"
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

// Connection
type Connection struct {

	// Standard connection
	Conn net.Conn

	// Weight
	Weight int

	// Hang
	Hang <-chan struct{}

	// Protocol
	Protocol Protocol

	// Packet pool
	PacketPool *sync.Pool

	// Timeout
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	// Connection lifetime
	MaxLifeTime time.Duration

	// Reader/Writer buffer
	ReaderBufferSize int
	WriterBufferSize int

	// Disconnect queue
	DQueue chan *Connection

	// Recv queue
	RQueue chan *Packet

	err    error
	ctx    context.Context
	cancel context.CancelFunc
	locker sync.Mutex
}

// Start
func (c *Connection) Start(ctx context.Context) {
	if c.MaxLifeTime <= 0 {
		c.ctx, c.cancel = context.WithCancel(ctx)
	} else {
		c.ctx, c.cancel = context.WithTimeout(ctx, c.MaxLifeTime)
	}
	// Handle remote packet
	go c.handleRemotePacket()
}

// Stop
func (c *Connection) Stop() {
	c.cancel()
}

// Send
func (c *Connection) Send(m Message) (n int64, err error) {
	return c.send(m)
}

// Error
func (c *Connection) Error() (err error) {
	return c.err
}

// send
func (c *Connection) send(m Message) (n int64, err error) {
	c.locker.Lock()
	defer c.locker.Unlock()
	// Write to connection
	if c.WriteTimeout > 0 {
		c.Conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
	}
	return m.WriteTo(c.Conn)
}

// handleRemotePacket
func (c *Connection) handleRemotePacket() {
	// Buffer
	r := bufio.NewReaderSize(c.Conn, c.ReaderBufferSize)

	for {
		if c.err != nil {
			break
		}

		if c.IdleTimeout > 0 {
			c.Conn.SetReadDeadline(time.Now().Add(c.IdleTimeout))
		}

		m := c.Protocol.NewMessage()

		// Read from connection
		if _, c.err = m.ReadFrom(r); c.err != nil {
			break
		}

		p := c.PacketPool.Get().(*Packet)
		p.reset()
		p.ctx = c.ctx
		p.conn = c
		p.message = m
		p.protocol = c.Protocol

		select {
		// Cancel
		case <-c.ctx.Done():
			c.err = c.ctx.Err()
			break
		// Hang up
		case <-c.Hang:
			c.err = errors.New("hang up")
			break
		// Submit to transport
		case c.RQueue <- p:
			// In the pipe, five by five!
		}
	}

	// Disconnect
	c.Conn.Close()

	// TODO 未处理完的数据包

	c.DQueue <- c
}
