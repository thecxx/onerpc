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
	"net"
	"time"

	"github.com/govoltron/onerpc/transport"
)

// ServiceDial
func (c *Cluster) ServiceDial(srvname string, timeout time.Duration) transport.DialFunc {
	return func(someone uint64) (conn net.Conn, uniqid uint64, err error) { return }
}

type Cluster struct {
	discovery Discovery
}

func NewCluster() (c *Cluster) {
	return &Cluster{}
}

// RegisterName 注册服务，并为服务名设置接收服务。
func (c *Cluster) RegisterName(srvname string, receiver Service) {
	c.recv(srvname, receiver)
}

// Send 向指定服务发送请求，并返回响应结果。
func (c *Cluster) Send(srvname string, in []byte) (out []byte, err error) {

	// // send to underlay
	// pk, err := c.send(srvname, in)
	// if err != nil {
	// 	return
	// }

	// for {
	// 	select {
	// 	// OK
	// 	case <-pk.Done:

	// 	}
	// }

	return
}

// Async 向指定服务发送请求，并返回响应结果。
func (c *Cluster) Async(srvname string, in []byte, handler func(out []byte, err error)) (err error) {
	return
}

// func (c *Cluster) call(srvname string, in []byte) {
// 	seq := int64(0)
// 	sf := new(stackframe)
// 	sf.push(in)
// 	sf.push("xxx")
// 	sf.push(seq)
// 	sf.push((int)(0))
// 	sf.push((error)(nil))

// }

// func (c *Cluster) send(srvname string, buff []byte) (pk *packet, err error) {

// 	pk = &packet{
// 		Buff: buff,
// 		Done: make(chan struct{}),
// 	}

// 	return
// }

// Recv 注册服务，并为服务名设置接收服务。
func (c *Cluster) recv(srvname string, receiver Service) {}
