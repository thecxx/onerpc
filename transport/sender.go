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

// Sender
type Sender struct {
	// Sended packet
	SP Packet

	// Recved packet
	RP Packet

	Err error

	// OnRecv
	OnRecv func(rp Packet, err error)

	// Done signal
	Done chan *Sender
}

// Ack
func (s *Sender) Ack(rp Packet, err error) {
	s.RP, s.Err = rp, err

	if s.OnRecv != nil {
		s.OnRecv(rp, err)
		return
	}

	select {
	// Ack
	case s.Done <- s:
		// Nothing to do
	// Discard
	default:
		// Nothing to do
	}
}
