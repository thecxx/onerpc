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
)

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

	// Line
	line *Line

	// Transport
	transport *Transport

	// Message
	message Message

	// Protocol
	proto Protocol

	// Already replied
	replied bool
}

// Bytes
func (p *Packet) Bytes() []byte {
	return p.message.Bytes()
}

// IsOneway
func (p *Packet) IsOneway() bool {
	return p.message.IsOneway()
}

// SetReplied
func (p *Packet) SetReplied() {
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

// NewMessage
func (p *Packet) NewMessage() (m Message) {
	return p.transport.newm()
}

// NewReply
func (p *Packet) NewReply() (r Message) {
	r = p.transport.newm()
	// Same sequence number
	r.SetSeq(p.message.Seq())
	return
}

// Done
func (p *Packet) Done() <-chan struct{} {
	return p.ctx.Done()
}

// reset
func (p *Packet) reset() {
	p.ctx = nil
	p.line = nil
	p.message = nil
	p.proto = nil
	p.replied = false
}
