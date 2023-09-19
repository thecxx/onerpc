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
)

type MessageWriter interface {

	// Reply
	Reply(b []byte) (n int64, err error)
}

type Handler interface {
	ServePacket(w MessageWriter, p *Packet)
}

type HandleFunc func(w MessageWriter, p *Packet)

func (f HandleFunc) ServePacket(w MessageWriter, p *Packet) {
	f(w, p)
}

type Packet struct {
	// Context
	ctx context.Context

	// Connection
	conn *Connection

	// Message
	message Message

	// Protocol
	proto Protocol

	// Already replied
	replied bool
}

// Seq
func (p *Packet) Seq() uint64 {
	return p.message.Seq()
}

// Bytes
func (p *Packet) Bytes() []byte {
	return p.message.Bytes()
}

// IsOneway
func (p *Packet) IsOneway() bool {
	return p.message.IsOneway()
}

// setReplied
func (p *Packet) setReplied() {
	p.replied = true
}

// IsReplied
func (p *Packet) IsReplied() bool {
	return p.replied
}

// Protocol returns the protocol version.
func (p *Packet) Protocol() string {
	return p.proto.Version()
}

// Done
func (p *Packet) Done() <-chan struct{} {
	return p.ctx.Done()
}

type messageWriter struct {
	conn   *Connection
	packet *Packet
	proto  Protocol
}

// Reply implements MessageWriter.
func (w messageWriter) Reply(b []byte) (n int64, err error) {
	if w.packet.IsOneway() {
		return 0, errors.New("oneway message")
	}
	if w.packet.IsReplied() {
		return 0, errors.New("already replied")
	}

	m := w.proto.NewMessage()
	m.Store(b)
	m.SetSeq(w.packet.Seq())

	defer func() {
		w.packet.setReplied()
	}()

	// Send message
	return w.conn.Send(m)
}
