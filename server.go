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
)

type Server struct {
	listener Listener
	protocol Protocol
	handler  Handler
	group    Group
	cancel   func()
}

// NewServer
func NewServer(listener Listener) (s *Server) {
	return &Server{}
}

// Listen
func (s *Server) Listen() (err error) {
	// Listen
	ln, err := s.listener.Listen()
	if err != nil {
		return
	}

	// Context
	ctx, cancel := context.WithCancel(context.Background())
	// Cancel
	s.cancel = func() {
		ln.Close()
		cancel()
	}

	// Accept
	go func() {
		for {
			if conn, err := ln.Accept(); err != nil {
				// TODO
			} else {
				s.join(ctx, conn)
			}
		}
	}()

	return
}

// Close
func (s *Server) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}

// join
func (s *Server) join(ctx context.Context, conn net.Conn) {
	cc := new(Connection)
	cc.Conn = conn
	cc.Hang = nil
	cc.Proto = s.protocol

	s.group.Add(cc)
	// Bootstrap connection
	go func() {
		if err := cc.Run(ctx, s.handleMessage); err != nil {
			// TODO
		}
		s.group.Remove(cc)
	}()
}

// handleMessage
func (s *Server) handleMessage(cc *Connection, message Message) {
	s.handler.ServeMessage(nil, message)
}
