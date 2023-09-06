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

package transport

// // Sender
// type Sender struct {
// 	// Message
// 	message Message

// 	// Reply
// 	reply Message

// 	// handler
// 	handler func(reply []byte, err error)

// 	err error

// 	// Done signal
// 	done chan struct{}
// }

// // Reply
// func (s *Sender) Reply(r Message, err error) {
// 	s.reply, s.err = r, err

// 	if s.handler != nil {
// 		s.handler(s.reply.Bytes(), err)
// 	} else {
// 		close(s.done)
// 	}
// }

// // reset
// func (s *Sender) reset() {
// 	s.message = nil
// 	s.reply = nil
// 	s.handler = nil
// 	s.err = nil
// 	s.done = nil
// }
