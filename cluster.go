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
)

type ClusterDialer struct {
	srvname string
	timeout time.Duration
	cluster *Cluster
}

// Dial implements transport.Dialer.
func (d *ClusterDialer) Dial() (conn net.Conn, weight int, hang <-chan struct{}, err error) {
	return
}

// Hang implements transport.Dialer.
func (d *ClusterDialer) Hang(conn net.Conn) {

}

type ClusterListener struct {
	srvname string
	network string
	addr    string
	weight  int
	cluster *Cluster
}

// Listen implements transport.Listener.
func (l *ClusterListener) Listen() (ln net.Listener, err error) {
	return
}

type Cluster struct {
	discovery Discovery
}

// NewCluster
func NewCluster() (c *Cluster) {
	return &Cluster{
		discovery: nil,
	}
}

// NewServiceDialer
func (c *Cluster) NewServiceDialer(srvname string, timeout time.Duration) (d *ClusterDialer) {
	return &ClusterDialer{srvname: srvname, timeout: timeout}
}

// NewServiceListener
func (c *Cluster) NewServiceListener(srvname, network, addr string, weight int) (l *ClusterListener) {
	return &ClusterListener{srvname: srvname, network: network, addr: addr, weight: weight}
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
