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
)

type MessageWriter interface {

	// Send
	Send(b []byte) (err error)
}

type Handler interface {

	// ServeMessage
	ServeMessage(w MessageWriter, m SimpleMessage)
}

type HandleFunc func(w MessageWriter, m SimpleMessage)

// ServeMessage
func (f HandleFunc) ServeMessage(w MessageWriter, m SimpleMessage) {
	f(w, m)
}

type Dialer interface {

	// Dial
	Dial(hold bool) (conn net.Conn, weight int, hang <-chan struct{}, err error)

	// Hang
	Hang(conn net.Conn) (err error)

	// Close
	Close() (err error)
}

type Listener interface {

	// Listen
	Listen() (ln net.Listener, err error)
}
