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
	"sync"
	"sync/atomic"
)

type RoundRobinBalancer struct {
	rss []*Connection
	rsm map[*Connection]int
	cu  uint32
	mu  sync.RWMutex
}

// NewRoundRobinBalancer
func NewRoundRobinBalancer() (b *RoundRobinBalancer) {
	return &RoundRobinBalancer{
		rss: make([]*Connection, 0),
		rsm: make(map[*Connection]int),
	}
}

// Next implements Balancer.
func (b *RoundRobinBalancer) Next() (c *Connection) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.rss[int(atomic.AddUint32(&b.cu, 1))%len(b.rss)]
}

// Add implements Balancer.
func (b *RoundRobinBalancer) Add(c *Connection, _ int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Update
	b.update(append([]*Connection{c}, b.rss...))
}

// Remove implements Balancer.
func (b *RoundRobinBalancer) Remove(c *Connection) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Remove
	if i, ok := b.rsm[c]; !ok {
		return
	} else if i < len(b.rss) {
		if i+1 == len(b.rss) {
			b.rss = b.rss[0:i]
		} else {
			b.rss = append(b.rss[0:i], b.rss[i+1:]...)
			for j := i; j < len(b.rss); i++ {
				b.rsm[b.rss[j]] = j
			}
		}
		delete(b.rsm, c)
	}
}

// update
func (b *RoundRobinBalancer) update(cs []*Connection) {
	var (
		rsm = make(map[*Connection]int)
		rss = make([]*Connection, len(cs))
	)
	for i, c := range cs {
		rss[i], rsm[c] = c, i
	}
	b.rss, b.rsm = rss, rsm
}
