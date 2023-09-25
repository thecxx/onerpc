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

package line

import (
	"net"
	"runtime"
	"sync/atomic"
)

type DirectDialer struct {
	idle uint32
}

// NewDirectDialer
func NewDirectDialer() *DirectDialer {
	return &DirectDialer{}
}

// Dial implements onerpc.Dialer.
func (dialer *DirectDialer) Dial(hold bool) (conn net.Conn, weight int, hang <-chan struct{}, err error) {
	for {
		if atomic.LoadUint32(&dialer.idle) > 0 {
			break
		}
		if hold {
			runtime.Gosched()
		} else {
			return
		}
	}

	panic("unimplemented")
}

// Hang implements onerpc.Dialer.
func (dialer *DirectDialer) Hang(conn net.Conn) {
	panic("unimplemented")
}

// Close implements onerpc.Dialer.
func (dialer *DirectDialer) Close() (err error) {
	panic("unimplemented")
}
