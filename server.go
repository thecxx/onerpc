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
	"net"
	"time"

	"github.com/govoltron/onerpc/proto"
	"github.com/govoltron/onerpc/transport"
)

type Server struct {
	transport transport.Transport
	listen    transport.ListenFunc
	packet    transport.PacketFunc
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewServer
func NewServer(listen transport.ListenFunc, packet transport.PacketFunc, opts ...Option) (s *Server) {
	s = &Server{
		listen: listen,
		packet: packet,
	}
	// Set options
	for _, setOpt := range opts {
		setOpt(s)
	}
	// Option: ReadTimeout
	if s.transport.ReadTimeout < 0 {
		s.transport.ReadTimeout = 3 * time.Second
	}
	// Option: WriteTimeout
	if s.transport.WriteTimeout < 0 {
		s.transport.WriteTimeout = 3 * time.Second
	}
	// Option: IdleTimeout
	if s.transport.IdleTimeout < 0 {
		s.transport.IdleTimeout = 10 * time.Second
	}
	// Option: MaxLifeTime
	if s.transport.MaxLifeTime < 0 {
		s.transport.MaxLifeTime = 600 * time.Second
	}
	// Option: ReaderBufferSize
	if s.transport.ReaderBufferSize < 0 {
		s.transport.ReaderBufferSize = 10 * 1024
	}
	// Option: WriterBufferSize
	if s.transport.WriterBufferSize < 0 {
		s.transport.WriterBufferSize = 10 * 1024
	}

	s.transport.Event = s
	s.transport.Packet = s.packet

	return
}

// SetReadTimeout implements CanOption.
func (s *Server) SetReadTimeout(timeout time.Duration) {
	s.transport.ReadTimeout = timeout
}

// SetWriteTimeout implements CanOption.
func (s *Server) SetWriteTimeout(timeout time.Duration) {
	s.transport.WriteTimeout = timeout
}

// SetIdleTimeout implements CanOption.
func (s *Server) SetIdleTimeout(timeout time.Duration) {
	s.transport.IdleTimeout = timeout
}

// SetMaxLifeTime implements CanOption.
func (s *Server) SetMaxLifeTime(timeout time.Duration) {
	s.transport.MaxLifeTime = timeout
}

// SetReaderBufferSize implements CanOption.
func (s *Server) SetReaderBufferSize(size int) {
	s.transport.ReaderBufferSize = size
}

// SetWriterBufferSize implements CanOption.
func (s *Server) SetWriterBufferSize(size int) {
	s.transport.WriterBufferSize = size
}

// Listen
func (s *Server) Listen() (err error) {
	l, err := s.listen()
	if err != nil {
		return
	}

	// Start Transport
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.transport.Start(s.ctx)

	// Handle foreign connect
	go s.handleForeignConnect(l)

	return
}

// Broadcast
func (s *Server) Broadcast(ctx context.Context, sp transport.Packet) (err error) {
	sp.SetOneway()
	_, err = s.transport.Send(ctx, sp)
	return
}

// OnPacket
func (s *Server) OnPacket(ctx context.Context, rp transport.Packet) (sp transport.Packet, err error) {
	fmt.Printf("OnPacket OK: %+v\n", rp)

	sp = proto.NewPacket()
	sp.Store([]byte("OK, I got!"))

	go func() {
		sp := proto.NewPacket()
		sp.Store([]byte("OK, Is't broadcasting!"))
		s.Broadcast(ctx, sp)
	}()

	return
}

// handleForeignConnect
func (s *Server) handleForeignConnect(l net.Listener) {
	for {
		if conn, err := l.Accept(); err != nil {
			break
		} else {
			s.transport.Join(conn, 0)
		}
	}
}

// Close
func (s *Server) Close() {
	s.transport.Stop()
	s.cancel()
}
