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
	"io"
)

type Message interface {

	// Sequence number
	Seq() uint64

	// Set sequence number
	SetSeq(number uint64)

	// NeedReply
	NeedReply() (need bool)

	// DoNotReply
	DoNotReply()

	// Bytes
	Bytes() (b []byte)

	// Store
	Store(buff []byte)

	// ReadFrom
	ReadFrom(r io.Reader) (n int64, err error)

	// WriteTo
	WriteTo(w io.Writer) (n int64, err error)
}

type SimpleMessage interface {

	// Sequence number
	Seq() uint64

	// NeedReply
	NeedReply() (need bool)

	// Bytes
	Bytes() (b []byte)
}

type Protocol interface {

	// NewMessage
	NewMessage() (message Message)
}
