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
	"time"
)

type Bundler struct {
	lines  map[*Line]time.Time
	locker sync.RWMutex
}

// Add
func (b *Bundler) Add(l *Line) {
	b.locker.Lock()
	defer b.locker.Unlock()
	// Add line
	b.lines[l] = time.Now()
}

// Remove
func (b *Bundler) Remove(l *Line) {
	b.locker.Lock()
	defer b.locker.Unlock()
	// Remove line
	delete(b.lines, l)
}

// Walk
func (b *Bundler) Walk(fn func(l *Line)) {
	b.locker.RLock()
	defer b.locker.RUnlock()
	// for-range
	for l := range b.lines {
		fn(l)
	}
}
