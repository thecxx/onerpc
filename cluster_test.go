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

package onerpc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/thecxx/onerpc"
	"github.com/thecxx/onerpc/link"
	"github.com/thecxx/onerpc/middleware"
	"github.com/thecxx/onerpc/protocol"
)

// func TestCluster_Send(t *testing.T) {
// 	cluster := onerpc.NewCluster()
// 	// d, err := onerpc.NewEtcdDiscovery()
// 	client := onerpc.NewClient(
// 		// client.DirectDial("tcp", "127.0.0.1:80"),
// 		cluster.NewServiceDialer("user-core-service", 1*time.Second),
// 	)
// 	err := client.Connect()
// 	if err != nil {
// 		return
// 	}
// 	defer client.Close()

// 	// client.Send(context.TODO(), protocol.NewPacket())
// }

func TestServer_Start(t *testing.T) {
	// cluster := onerpc.NewCluster()

	server := onerpc.NewServer(
		link.NewDirectListener("tcp", "127.0.0.1:8080"),
		onerpc.WithProtocol(protocol.NewProtocol()),
	)
	err := server.Listen()
	if err != nil {
		t.Errorf("Start failed, error is %s", err.Error())
		return
	}
	defer server.Close()

	server.Use(middleware.WithACL)

	server.HandleFunc(func(w onerpc.MessageWriter, p *onerpc.Packet) {
		fmt.Printf("OnPacket OK: %+v\n", string(p.Bytes()))

		go func() {
			server.Broadcast(context.TODO(), []byte("OK, Is't broadcasting!"))
		}()

		w.Reply([]byte("OK, I got!"))
	})

	// server.Broadcast(context.TODO(), protocol.NewPacket())

	time.Sleep(20 * time.Second)
}

func TestClient_Send(t *testing.T) {
	client := onerpc.NewClient(
		link.NewDirectDialer("tcp", "127.0.0.1:8080"),
		onerpc.WithProtocol(protocol.NewProtocol()),
	)
	err := client.Connect()
	if err != nil {
		t.Errorf("Connect failed, error is %s", err.Error())
		return
	}
	defer client.Close()

	client.HandleFunc(func(w onerpc.MessageWriter, p *onerpc.Packet) {
		if p.IsOneway() {
			fmt.Printf("OnPacket oneway\n")
		}
		fmt.Printf("OnPacket OK: %s\n", string(p.Bytes()))
		return
	})

	for i := 0; i < 10; i++ {

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

		r, err := client.Send(ctx, []byte("Hello world!"))
		cancel()
		if err != nil {
			t.Errorf("Async failed, error is %s", err.Error())
			continue
		}
		t.Logf("ack: %s", string(r))

	}

	time.Sleep(5 * time.Second)

	// r, err := client.Send(context.TODO(), sp)
	// if err != nil {
	// 	t.Errorf("Send failed, error is %s", err.Error())
	// 	return
	// }

	// rp := r.(*protocol.Packet)

	// t.Logf("sh: %+v rh: %+v", *sp.Header, *rp.Header)
	// t.Logf("ack: %s", rp.String())
}
