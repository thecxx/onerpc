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

package middleware

import (
	"fmt"

	"github.com/govoltron/onerpc/transport"
)

func WithACL(next transport.Handler) transport.Handler {
	return transport.HandleFunc(func(w transport.MessageWriter, p *transport.Packet) {
		if true {

			fmt.Printf("Protocol: %s\n", p.Protocol())

			m := p.NewReply()
			m.Store([]byte("WithACL 拒绝"))
			w.WriteMessage(m)
			return
		}

		next.ServePacket(w, p)
	})
}
