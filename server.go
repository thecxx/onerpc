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
	ReaderBufferSize int
	WriterBufferSize int
	transport        transport.Transport
	listen           transport.ListenFunc
	packet           transport.PacketFunc
}

// NewServer
func NewServer(listen transport.ListenFunc, packet transport.PacketFunc) (s *Server) {
	return &Server{
		ReaderBufferSize: 1024,
		WriterBufferSize: 1024,
		listen:           listen,
		packet:           packet,
	}
}

// Listen
func (s *Server) Listen() (err error) {
	l, err := s.listen()
	if err != nil {
		return
	}

	s.transport.IdleTimeout = 10 * time.Second
	s.transport.ReadTimeout = 3 * time.Second
	s.transport.WriteTimeout = 3 * time.Second
	// s.transport.LifeTime = 600 * time.Second
	s.transport.ReaderBufferSize = s.ReaderBufferSize
	s.transport.WriterBufferSize = s.WriterBufferSize
	s.transport.Packet = s.packet
	s.transport.Event = s

	// Transport
	s.transport.Start(context.TODO())

	// Handle foreign connect
	go s.handleForeignConnect(l)

	return
}

// Stop
func (s *Server) Close() {

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

	// go func() {
	// 	sp := proto.NewPacket()
	// 	// sp.Payload = []byte("OK, Is't broadcasting!")
	// 	sp.Store([]byte("OK, Is't broadcasting!"))
	// 	s.Broadcast(ctx, sp)
	// }()

	return
}

// handleForeignConnect
func (s *Server) handleForeignConnect(l net.Listener) {
	for {
		if conn, err := l.Accept(); err != nil {
			break
		} else {
			s.transport.Join(conn)
		}
	}
}
