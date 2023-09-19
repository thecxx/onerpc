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
	"time"
)

type Client struct {
	transport   Transport
	dialer      Dialer
	handler     Handler
	middlewares []func(next Handler) Handler
	// Context
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClient
func NewClient(dialer Dialer, opts ...Option) (c *Client) {
	c = &Client{
		dialer: dialer,
	}
	// Set options
	for _, setOpt := range opts {
		setOpt(c)
	}
	// Option: Protocol
	if c.transport.Proto == nil {
		panic("invalid protocol for client")
	}
	// Option: Balancer
	if c.transport.Balancer == nil {
		c.transport.Balancer = NewRoundRobinBalancer()
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

	c.transport.Handler = c

	return
}

// SetProtocol implements CanOption.
func (c *Client) SetProtocol(p Protocol) {
	c.transport.Proto = p
}

// SetBalancer implements CanOption.
func (c *Client) SetBalancer(b Balancer) {
	c.transport.Balancer = b
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

// ServePacket
func (c *Client) ServePacket(w MessageWriter, p *Packet) {
	if c.handler != nil {
		c.handler.ServePacket(w, p)
	}
}

// Connect
func (c *Client) Connect() (err error) {
	conn, weight, hang, err := c.dialer.Dial()
	if err != nil {
		return
	}

	c.transport.Dialer = c.dialer

	// Start Transport
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.transport.Ctx = c.ctx

	// First connection
	c.transport.Join(conn, weight, hang)

	return
}

// Handle
func (c *Client) Handle(handler Handler) {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}
	c.handler = handler
}

// HandleFunc
func (c *Client) HandleFunc(handler func(w MessageWriter, p *Packet)) {
	c.handler = HandleFunc(handler)
}

// Use
func (c *Client) Use(middleware func(next Handler) Handler) {
	c.middlewares = append(c.middlewares, middleware)
}

// Send
func (c *Client) Send(ctx context.Context, message []byte) (reply []byte, err error) {
	return c.transport.Send(ctx, message)
}

// Close
func (c *Client) Close() {
	c.cancel()
}
