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
	"sync"
	"testing"
	"time"

	"github.com/govoltron/onerpc"
	"github.com/govoltron/onerpc/libnet"
	"github.com/govoltron/onerpc/proto"
	"github.com/govoltron/onerpc/transport"
)

func TestCluster_Send(t *testing.T) {
	cluster := onerpc.NewCluster()
	// d, err := onerpc.NewEtcdDiscovery()
	client := onerpc.NewClient(
		// client.DirectDial("tcp", "127.0.0.1:80"),
		cluster.NewServiceDialer("user-core-service", 1*time.Second),
		func() transport.Packet { return proto.NewPacket() },
	)
	err := client.Connect()
	if err != nil {
		return
	}
	defer client.Close()

	client.Send(context.TODO(), proto.NewPacket())
}

func TestServer_Start(t *testing.T) {
	server0 := onerpc.NewServer(
		libnet.NewDirectListener("tcp", ":8080"),
		func() transport.Packet { return proto.NewPacket() },
	)
	err := server0.Listen()
	if err != nil {
		t.Errorf("Start failed, error is %s", err.Error())
		return
	}
	defer server0.Close()

	server1 := onerpc.NewServer(
		libnet.NewDirectListener("tcp", ":8081"),
		func() transport.Packet { return proto.NewPacket() },
	)
	err = server1.Listen()
	if err != nil {
		t.Errorf("Start failed, error is %s", err.Error())
		return
	}
	defer server1.Close()

	// server.Broadcast(context.TODO(), protocol.NewPacket())

	time.Sleep(20 * time.Second)
}

func TestClient_Send(t *testing.T) {
	client := onerpc.NewClient(
		libnet.NewDirectDialer("tcp", "127.0.0.1:8080", "127.0.0.1:8081"),
		func() transport.Packet { return proto.NewPacket() },
		onerpc.WithBalancer(transport.NewWeightRoundRobinBalancer()),
	)
	err := client.Connect()
	if err != nil {
		t.Errorf("Connect failed, error is %s", err.Error())
		return
	}
	defer client.Close()

	time.Sleep(2 * time.Second)

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {

		sp := proto.NewPacket()
		sp.Store([]byte("Hello world!"))

		wg.Add(1)

		client.Async(context.TODO(), sp, func(rp transport.Packet, err error) {
			defer wg.Done()
			if err != nil {
				t.Errorf("Async failed, error is %s", err.Error())
				return
			}
			t.Logf("ack: %s", rp.String())
		})

	}

	wg.Wait()

	time.Sleep(5 * time.Second)

	// r, err := client.Send(context.TODO(), sp)
	// if err != nil {
	// 	t.Errorf("Send failed, error is %s", err.Error())
	// 	return
	// }

	// rp := r.(*proto.Packet)

	// t.Logf("sh: %+v rh: %+v", *sp.Header, *rp.Header)
	// t.Logf("ack: %s", rp.String())
}
