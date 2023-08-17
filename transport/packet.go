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

import (
	"io"
)

// Packet
type Packet interface {
	// Seq gets the sequence number
	Seq() uint64

	// SetSeq sets the sequence number
	SetSeq(seq uint64)

	// IsOneway
	IsOneway() bool

	// SetOneway
	SetOneway()

	// Bytes
	Bytes() []byte

	// String
	String() string

	// Store
	Store(data []byte)

	// Read reads the packet from r
	ReadFrom(r io.Reader) (n int64, err error)

	// Write writes the packet into w
	WriteTo(w io.Writer) (n int64, err error)
}
