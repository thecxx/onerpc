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
	"sync"
	"time"
)

type CGroup struct {
	locker  sync.RWMutex
	workers map[*Connection]time.Time
}

// Add
func (b *CGroup) Add(c *Connection) {
	b.locker.Lock()
	defer b.locker.Unlock()
	// If emtry
	if b.workers == nil {
		b.workers = make(map[*Connection]time.Time)
	}
	// Add connection
	b.workers[c] = time.Now()
}

// Remove
func (b *CGroup) Remove(c *Connection) {
	b.locker.Lock()
	defer b.locker.Unlock()
	// Remove connection
	delete(b.workers, c)
}

// Walk
func (b *CGroup) Range(fn func(c *Connection)) {
	b.locker.RLock()
	defer b.locker.RUnlock()
	// for-range
	for l := range b.workers {
		fn(l)
	}
}
