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
	"net"
	"sync"
	"time"
)

// Connection
type Connection struct {

	// Uniq ID
	UniqID uint64

	// Standard connection
	Conn net.Conn

	// New packet
	Packet PacketFunc

	// Timeout
	IdleTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// Connection lifetime
	MaxLifeTime time.Duration

	// Reader/Writer buffer
	ReaderBufferSize int
	WriterBufferSize int

	// Recv queue
	RQueue chan *Receiver

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
	// Handle foreign packet
	go c.handleForeignPacket()
}

// Stop
func (c *Connection) Stop() {
	c.cancel()
}

// Send
func (c *Connection) Send(sp Packet) (n int64, err error) {
	return c.send(sp)
}

// Error
func (c *Connection) Error() (err error) {
	return c.err
}

// ack
func (c *Connection) ack(sp Packet) (n int64, err error) {
	return c.send(sp)
}

// send
func (c *Connection) send(sp Packet) (n int64, err error) {
	c.locker.Lock()
	defer c.locker.Unlock()
	// Write to connection
	if c.WriteTimeout > 0 {
		c.Conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
	}
	return sp.WriteTo(c.Conn)
}

// handleForeignPacket
func (c *Connection) handleForeignPacket() {
	// Buffer
	r := bufio.NewReaderSize(c.Conn, c.ReaderBufferSize)

	for {
		if c.err != nil {
			break
		}

		if c.IdleTimeout > 0 {
			c.Conn.SetReadDeadline(time.Now().Add(c.IdleTimeout))
		}

		rp := c.Packet()

		// Read from connection
		if _, c.err = rp.ReadFrom(r); c.err != nil {
			break
		}

		select {
		// Cancel
		case <-c.ctx.Done():
			c.err = c.ctx.Err()
			break
		// Submit to transport
		case c.RQueue <- &Receiver{RP: rp, OnAck: c.ack}:
			// In the pipe, five by five!
		}
	}

}
