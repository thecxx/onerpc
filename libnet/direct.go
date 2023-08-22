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

package libnet

import (
	"errors"
	"net"
	"time"

	"github.com/govoltron/onerpc/transport"
)

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
	return NewDirectDialerWithTimeout(1*time.Second, network, addrs...)
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
func (d *DirectDialer) Dial() (conn net.Conn, weight int, hang chan struct{}, err error) {
	var (
		contact *DirectContact
	)
	for _, value := range d.contacts {
		if value.conn == nil {
			contact = value
			break
		}
	}
	if contact == nil {
		return nil, 0, nil, errors.New("no idle contact found")
	}
	// Dial
	if conn, err = net.DialTimeout(d.network, contact.addr, d.timeout); err == nil {
		weight = transport.WeightNormal
		hang = make(chan struct{}, 1)
		contact.conn = conn
		contact.hang = hang
	}
	return
}

// Hang implements transport.Dialer.
func (d *DirectDialer) Hang(conn net.Conn) {
	var (
		contact *DirectContact
	)
	for _, value := range d.contacts {
		if value.conn != nil && value.conn == conn {
			contact = value
			break
		}
	}
	if contact != nil {
		close(contact.hang)
		contact.conn = nil
		contact.hang = nil
		return
	}
}

type DirectListener struct {
	network string
	addr    string
	ln      net.Listener
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
