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
)

type Client struct {
	dialer   Dialer
	protocol Protocol
	balancer Balancer
	handler  Handler
	sequence sequence
	cancel   func()
}

// NewClient
func NewClient(dialer Dialer) (c *Client) {
	return &Client{}
}

// Connect
func (c *Client) Connect() (err error) {
	// Dial
	conn, weight, hang, err := c.dialer.Dial(false)
	if err != nil {
		return
	} else if conn == nil {
		return errors.New("not answered")
	}

	// Context
	ctx, cancel := context.WithCancel(context.Background())
	// Cancel
	c.cancel = func() {
		c.dialer.Close()
		cancel()
	}

	// First join
	c.join(ctx, conn, weight, hang)
	// Dial
	go func() {
		for {
			conn, weight, hang, err := c.dialer.Dial(true)
			if err != nil {
				// TODO
			} else if conn != nil {
				c.join(ctx, conn, weight, hang)
			}
		}
	}()

	return
}

// Send
func (c *Client) Send(ctx context.Context, m []byte) (r []byte, err error) {

	message := newMessage(c.protocol, m)
	message.SetSeq(c.sequence.Next())

	sendMessage(ctx, nil, message)

	return
}

// Close
func (c *Client) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

// join
func (c *Client) join(ctx context.Context, conn net.Conn, weight int, hang <-chan struct{}) {
	cc := new(Connection)
	cc.Conn = conn
	cc.Hang = hang
	cc.Proto = c.protocol

	c.balancer.Add(cc, weight)
	// Bootstrap connection
	go func() {
		if err := cc.Run(ctx, c.handleMessage); err != nil {
			// TODO
		}
		c.balancer.Remove(cc)
		// Hang up
		c.dialer.Hang(conn)
	}()
}

// handleMessage
func (c *Client) handleMessage(cc *Connection, message Message) {

	// Handle reply
	if c.handleReply(cc, message) {
		return
	}

	var sent = false
	var writer = new(messageWriter)

	// Send message
	writer.send = func(b []byte) error {
		if sent {
			return errors.New("already sent")
		}
		defer func() {
			sent = true
		}()

		if !message.NeedReply() {
			return errors.New("no reply needed")
		}

		message := newMessage(c.protocol, b)
		// Set same sequence number
		message.SetSeq(message.Seq())

		return sendMessage(context.TODO(), cc, message)
	}

	c.handler.ServeMessage(writer, message)
}

// handleReply
func (c *Client) handleReply(cc *Connection, message Message) (handled bool) {

	// seq := message.Seq()
	return
}
