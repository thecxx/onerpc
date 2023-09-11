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

	"github.com/thecxx/onerpc/transport"
)

func WithACL(next transport.Handler) transport.Handler {
	return transport.HandleFunc(func(w transport.MessageWriter, p *transport.Packet) {
		if true {

			fmt.Printf("Protocol: %s => %s\n", p.Protocol(), string(p.Bytes()))

			_, err := w.Reply([]byte("WithACL 拒绝"))
			if err != nil {
				fmt.Printf("WithACL: %s\n", err.Error())
			}
			return
		}

		next.ServePacket(w, p)
	})
}
