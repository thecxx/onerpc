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
)

type Endpoint struct {
	l      *Line
	weight int
}

type node struct {
	current   int
	effective int
	endpoint  Endpoint
}

type WeightRoundRobinBalancer struct {
	rss []*node
	rsm map[*Line]int
	mu  sync.RWMutex
}

// NewWeightRoundRobinBalancer
func NewWeightRoundRobinBalancer() (b *WeightRoundRobinBalancer) {
	return &WeightRoundRobinBalancer{
		rss: make([]*node, 0),
		rsm: make(map[*Line]int),
	}
}

// Next
func (b *WeightRoundRobinBalancer) Next() (l *Line) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Next
	return b.next()
}

// Add implements Balancer.
func (b *WeightRoundRobinBalancer) Add(l *Line, weight int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Add
	var endpoints []Endpoint
	for _, node := range b.rss {
		endpoints = append(endpoints, node.endpoint)
	}
	endpoints = append(endpoints, Endpoint{l: l, weight: weight})
	// Update
	b.update(endpoints)
}

// Remove implements Balancer.
func (b *WeightRoundRobinBalancer) Remove(l *Line) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Remove
	if i, ok := b.rsm[l]; !ok {
		return
	} else if i < len(b.rss) {
		if i+1 == len(b.rss) {
			b.rss = b.rss[0:i]
		} else {
			b.rss = append(b.rss[0:i], b.rss[i+1:]...)
			for j := i; j < len(b.rss); i++ {
				b.rsm[b.rss[j].endpoint.l] = j
			}
		}
		delete(b.rsm, l)
	}
}

// update
func (b *WeightRoundRobinBalancer) update(endpoints []Endpoint) {
	// Get endpoints and update
	var (
		rss = make([]*node, 0)
		rsm = make(map[*Line]int)
	)
	for _, endpoint := range endpoints {
		node := &node{
			endpoint:  endpoint,
			effective: endpoint.weight,
		}
		if i, ok := b.rsm[endpoint.l]; ok {
			node.current = b.rss[i].current
			node.effective = b.rss[i].effective
		}
		rss = append(rss, node)
	}
	for i, node := range rss {
		rsm[node.endpoint.l] = i
	}
	// Replace
	b.rss, b.rsm = rss, rsm
}

// next
func (b *WeightRoundRobinBalancer) next() (l *Line) {
	var (
		total int
		best  *node
	)
	for i := 0; i < len(b.rss); i++ {
		node := b.rss[i]
		total += node.effective
		node.current += node.effective
		// -1 when the connection is abnormal,
		// +1 when the communication is successful
		if node.effective < node.endpoint.weight {
			node.effective++
		}
		// Maximum temporary weight node
		if best == nil || node.current > best.current {
			best = node
		}
	}
	if best != nil {
		best.current -= total
		l = best.endpoint.l
	}
	return
}

// node
func (b *WeightRoundRobinBalancer) node(l *Line) (node *node, ok bool) {
	if i, ok1 := b.rsm[l]; !ok1 {
		return nil, false
	} else if i < len(b.rss) {
		node, ok = b.rss[i], true
	}
	return
}
