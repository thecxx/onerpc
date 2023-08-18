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
	"fmt"
	"time"

	"github.com/govoltron/onerpc/transport"
)

type Client struct {
	transport transport.Transport
	dial      transport.DialFunc
	packet    transport.PacketFunc
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewClient
func NewClient(dial transport.DialFunc, packet transport.PacketFunc, opts ...Option) (c *Client) {
	c = &Client{
		dial:   dial,
		packet: packet,
	}
	// Set options
	for _, setOpt := range opts {
		setOpt(c)
	}
	// Option: ReadTimeout
	if c.transport.ReadTimeout < 0 {
		c.transport.ReadTimeout = 3 * time.Second
	}
	// Option: WriteTimeout
	if c.transport.WriteTimeout < 0 {
		c.transport.WriteTimeout = 3 * time.Second
	}
	// Option: IdleTimeout
	if c.transport.IdleTimeout < 0 {
		c.transport.IdleTimeout = 10 * time.Second
	}
	// Option: MaxLifeTime
	if c.transport.MaxLifeTime < 0 {
		c.transport.MaxLifeTime = 0
	}
	// Option: ReaderBufferSize
	if c.transport.ReaderBufferSize < 0 {
		c.transport.ReaderBufferSize = 10 * 1024
	}
	// Option: WriterBufferSize
	if c.transport.WriterBufferSize < 0 {
		c.transport.WriterBufferSize = 10 * 1024
	}

	c.transport.Event = c
	c.transport.Dial = c.dial
	c.transport.Packet = c.packet

	return
}

// SetReadTimeout implements CanOption.
func (c *Client) SetReadTimeout(timeout time.Duration) {
	c.transport.ReadTimeout = timeout
}

// SetWriteTimeout implements CanOption.
func (c *Client) SetWriteTimeout(timeout time.Duration) {
	c.transport.WriteTimeout = timeout
}

// SetIdleTimeout implements CanOption.
func (c *Client) SetIdleTimeout(timeout time.Duration) {
	c.transport.IdleTimeout = timeout
}

// SetMaxLifeTime implements CanOption.
func (c *Client) SetMaxLifeTime(timeout time.Duration) {
	c.transport.MaxLifeTime = timeout
}

// SetReaderBufferSize implements CanOption.
func (c *Client) SetReaderBufferSize(size int) {
	c.transport.ReaderBufferSize = size
}

// SetWriterBufferSize implements CanOption.
func (c *Client) SetWriterBufferSize(size int) {
	c.transport.WriterBufferSize = size
}

// Connect
func (c *Client) Connect() (err error) {
	conn, uniqid, err := c.dial(transport.Anyone)
	if err != nil {
		return
	}

	// Start Transport
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.transport.Start(c.ctx)

	// First connection
	c.transport.Join(conn, uniqid)

	return
}

// OnPacket
func (c *Client) OnPacket(ctx context.Context, rp transport.Packet) (sp transport.Packet, err error) {
	if rp.IsOneway() {
		fmt.Printf("OnPacket oneway\n")
	}
	fmt.Printf("OnPacket OK: %s\n", rp.String())
	return
}

// Send
func (c *Client) Send(ctx context.Context, sp transport.Packet) (rp transport.Packet, err error) {
	return c.transport.Send(ctx, sp)
}

// Async
func (c *Client) Async(ctx context.Context, sp transport.Packet, fn func(rp transport.Packet, err error)) {
	c.transport.Async(ctx, sp, fn)
}

// Close
func (c *Client) Close() {
	c.transport.Stop()
}
