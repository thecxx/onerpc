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
)

type messageWriter struct {
	send func([]byte) error
}

// Send
func (w messageWriter) Send(b []byte) (err error) {
	return w.send(b)
}

// newMessage
func newMessage(proto Protocol, b []byte) (message Message) {
	message = proto.NewMessage()
	message.Store(b)
	return
}

// sendMessage
func sendMessage(ctx context.Context,
	cc *Connection, message Message) (err error) {
	return cc.Send(ctx, message)
}
