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

package link

import (
	"errors"
	"net"
	"time"

	"github.com/govoltron/onerpc/transport"
)

type DirectListener struct {
	ln      net.Listener
	network string
	addr    string
}

// NewDirectListener
func NewDirectListener(network, addr string) *DirectListener {
	return &DirectListener{network: network, addr: addr}
}

// Listen implements transport.Listener.
func (l *DirectListener) Listen() (ln net.Listener, err error) {
	if l.ln != nil {
		return nil, errors.New("already listened")
	}
	if ln, err = net.Listen(l.network, l.addr); err == nil {
		l.ln = ln
	}
	return
}

type DirectContact struct {
	addr string
	conn net.Conn
	hang chan struct{}
}

type DirectDialer struct {
	network  string
	timeout  time.Duration
	contacts map[string]*DirectContact
}

// NewDirectDialer
func NewDirectDialer(network string, addrs ...string) (d *DirectDialer) {
	return NewDirectDialerWithTimeout(3*time.Second, network, addrs...)
}

// NewDirectDialerWithTimeout
func NewDirectDialerWithTimeout(timeout time.Duration, network string, addrs ...string) (d *DirectDialer) {
	d = &DirectDialer{
		timeout:  timeout,
		network:  network,
		contacts: make(map[string]*DirectContact),
	}
	for _, addr := range addrs {
		d.contacts[addr] = &DirectContact{addr: addr}
	}
	return
}

// Dial implements transport.Dialer.
func (d *DirectDialer) Dial() (conn net.Conn, weight int, hang <-chan struct{}, err error) {
	for _, contact := range d.contacts {
		if contact.conn == nil {
			return d.dial(contact)
		}
	}
	return nil, 0, nil, errors.New("no idle contact found")
}

// dial
func (d *DirectDialer) dial(contact *DirectContact) (conn net.Conn, weight int, hang <-chan struct{}, err error) {
	contact.conn, err = net.DialTimeout(d.network, contact.addr, d.timeout)
	if err == nil {
		contact.hang = make(chan struct{})
		return contact.conn, transport.WeightNormal, contact.hang, nil
	}
	return
}

// Hang implements transport.Dialer.
func (d *DirectDialer) Hang(conn net.Conn) {
	for _, contact := range d.contacts {
		if contact.conn != nil && contact.conn == conn {
			d.hang(contact)
			break
		}
	}
}

// hang
func (d *DirectDialer) hang(contact *DirectContact) {
	contact.conn = nil
	contact.hang = nil
}
