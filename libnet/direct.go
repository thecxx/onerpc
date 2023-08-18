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

package libnet

import (
	"net"
	"time"

	"github.com/govoltron/onerpc/transport"
)

// DirectDial
func DirectDial(network, addr string, timeout time.Duration) transport.DialFunc {
	return func(someone uint64) (conn net.Conn, uniqid uint64, err error) {
		conn, err = net.DialTimeout(network, addr, timeout)
		if err == nil {
			uniqid = 1
		}
		return
	}
}

// DirectListen
func DirectListen(network, addr string) transport.ListenFunc {
	return func() (listener net.Listener, err error) {
		return net.Listen(network, addr)
	}
}
