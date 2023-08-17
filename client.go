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
	ReaderBufferSize int
	WriterBufferSize int
	transport        transport.Transport
	dial             transport.DialFunc
	packet           transport.PacketFunc
}

// NewClient
func NewClient(dial transport.DialFunc, packet transport.PacketFunc) (c *Client) {
	return &Client{
		ReaderBufferSize: 1024,
		WriterBufferSize: 1024,
		dial:             dial,
		packet:           packet,
	}
}

// Connect
func (c *Client) Connect() (err error) {
	conn, err := c.dial()
	if err != nil {
		return
	}

	c.transport.IdleTimeout = 10 * time.Second
	c.transport.ReadTimeout = 3 * time.Second
	c.transport.WriteTimeout = 3 * time.Second
	// c.transport.LifeTime = 600 * time.Second
	c.transport.ReaderBufferSize = c.ReaderBufferSize
	c.transport.WriterBufferSize = c.WriterBufferSize
	c.transport.Dial = c.dial
	c.transport.Packet = c.packet
	c.transport.Event = c

	// Transport
	c.transport.Start(context.TODO())

	// First connection
	c.transport.Join(conn)

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
