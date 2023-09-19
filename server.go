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
	"net"
	"time"
)

type Server struct {
	transport   Transport
	listener    Listener
	handler     Handler
	middlewares []func(next Handler) Handler
	// Context
	ctx    context.Context
	cancel context.CancelFunc
}

// NewServer
func NewServer(listener Listener, opts ...Option) (s *Server) {
	s = &Server{
		listener: listener,
	}
	// Set options
	for _, setOpt := range opts {
		setOpt(s)
	}
	// Option: Protocol
	if s.transport.Proto == nil {
		panic("invalid protocol for server")
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

	s.transport.Handler = s

	return
}

// SetProtocol implements CanOption.
func (s *Server) SetProtocol(p Protocol) {
	s.transport.Proto = p
}

// SetBalancer implements CanOption.
func (s *Server) SetBalancer(b Balancer) {
	panic("option 'Balancer' not supported")
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

// ServePacket
func (s *Server) ServePacket(w MessageWriter, p *Packet) {
	if s.handler != nil {
		s.handler.ServePacket(w, p)
	}
}

// Listen
func (s *Server) Listen() (err error) {
	ln, err := s.listener.Listen()
	if err != nil {
		return
	}

	// Start Transport
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.transport.Ctx = s.ctx

	// Handle remote connect
	go s.handleRemoteConnect(ln)

	return
}

// Handle
func (s *Server) Handle(handler Handler) {
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		handler = s.middlewares[i](handler)
	}
	s.handler = handler
}

// HandleFunc
func (s *Server) HandleFunc(handler func(w MessageWriter, p *Packet)) {
	s.Handle(HandleFunc(handler))
}

// Use
func (s *Server) Use(middleware func(next Handler) Handler) {
	s.middlewares = append(s.middlewares, middleware)
}

// Broadcast
func (s *Server) Broadcast(ctx context.Context, message []byte) (err error) {
	return s.transport.Broadcast(ctx, message)
}

// Close
func (s *Server) Close() {
	s.cancel()
}

// handleRemoteConnect
func (s *Server) handleRemoteConnect(ln net.Listener) {
	for {
		if conn, err := ln.Accept(); err != nil {
			break
		} else {
			s.transport.Join(conn, 100, nil)
		}
	}
}
