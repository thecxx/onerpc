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

package protocol

import (
	"encoding/binary"
	"errors"
	"io"
	"math"

	"github.com/govoltron/onerpc/transport"
)

const (
	MagicNumber uint16 = 'H'<<8 | 'Y'
)

const (
	Version10 byte = 1 << 4
)

type Option byte

const (
	_ Option = 1 << iota
	_
	_
	_
	_
	_
	OptOneway
	OptLargeSize
)

// Packet header
// Format:
// 0 ~ 1 : 	Magic Number
// 2	 : 	Version
// 3 ~ 10: 	Sequence number
// 11    :  Options
type Header [12]byte

// MagicNumber
func (h *Header) MagicNumber() uint16 {
	return uint16(h[0])<<8 | uint16(h[1])
}

// Version
func (h *Header) Version() byte {
	return h[2]
}

// Seq
func (h *Header) Seq() (seq uint64) {
	return binary.BigEndian.Uint64(h[3:11])
}

// SetSeq
func (h *Header) SetSeq(seq uint64) {
	binary.BigEndian.PutUint64(h[3:11], seq)
}

// IsOneway
func (h *Header) IsOneway() bool {
	return h.GetOption(OptOneway)
}

// SetOneway
func (h *Header) SetOneway() {
	h.SetOption(OptOneway, true)
}

// GetOption
func (h *Header) GetOption(opt Option) (open bool) {
	return h[11]&byte(opt) == byte(opt)
}

// SetOption
func (h *Header) SetOption(opt Option, open bool) {
	if open {
		h[11] |= byte(opt)
	} else {
		h[11] &= ^byte(opt)
	}
}

// Message
type Message struct {
	*Header
	buff []byte
}

// NewMessage
func NewMessage() (m *Message) {
	m = &Message{
		Header: new(Header),
	}
	m.Header[0] = byte(MagicNumber & 0xFF00 >> 8)
	m.Header[1] = byte(MagicNumber & 0x00FF)
	m.Header[2] = Version10
	return
}

// Bytes
func (m *Message) Bytes() []byte {
	return m.buff
}

// Store
func (m *Message) Store(data []byte) {
	m.buff = data
}

// Reset
func (m *Message) Reset() {
	m.Header[11] = 0
	m.buff = m.buff[:0]
}

// ReadFrom
func (m *Message) ReadFrom(r io.Reader) (n int64, err error) {

	// Magic number
	_, err = io.ReadFull(r, m.Header[0:2])
	if err != nil {
		return
	}

	if m.MagicNumber() != MagicNumber {
		return 0, errors.New("invalid magic number")
	}

	// Version
	_, err = io.ReadFull(r, m.Header[2:3])
	if err != nil {
		return
	}

	if m.Version() != Version10 {
		return 0, errors.New("invalid version")
	}

	// Sequence number
	_, err = io.ReadFull(r, m.Header[3:11])
	if err != nil {
		return
	}

	// Options
	_, err = io.ReadFull(r, m.Header[11:12])
	if err != nil {
		return
	}

	var (
		size int
		buff []byte
	)
	if m.GetOption(OptLargeSize) {
		buff = make([]byte, 4)
	} else {
		buff = make([]byte, 2)
	}

	_, err = io.ReadFull(r, buff)
	if err != nil {
		return
	}

	if len(buff) == 4 {
		size = int(binary.BigEndian.Uint32(buff))
	} else {
		size = int(binary.BigEndian.Uint16(buff))
	}

	m.buff = make([]byte, size)

	_, err = io.ReadFull(r, m.buff)
	if err != nil {
		return
	}

	return int64(size), nil
}

// WriteTo
func (m *Message) WriteTo(w io.Writer) (n int64, err error) {
	var (
		size = len(m.buff)
		buff []byte
	)
	if size <= math.MaxUint16 {
		buff = make([]byte, 2)
	} else {
		buff = make([]byte, 4)
		m.Header.SetOption(OptLargeSize, true)
	}

	if len(buff) == 2 {
		binary.BigEndian.PutUint16(buff, uint16(size))
	} else {
		binary.BigEndian.PutUint32(buff, uint32(size))
	}

	_, err = w.Write(m.Header[:])
	if err != nil {
		return
	}
	_, err = w.Write(buff)
	if err != nil {
		return
	}
	n1, err := w.Write(m.buff)
	if err != nil {
		return
	}

	return int64(n1), nil
}

type Protocol struct {
}

// NewProtocol
func NewProtocol() (p *Protocol) {
	return new(Protocol)
}

// Version
func (p *Protocol) Version() string {
	return "onerpc message 1.0"
}

// NewMessage
func (p *Protocol) NewMessage() (m transport.Message) {
	return NewMessage()
}
