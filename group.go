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

// Connection group
type Group struct {
	members map[*Connection]time.Time
	locker  sync.RWMutex
}

// Add
func (g *Group) Add(c *Connection) {
	g.locker.Lock()
	defer g.locker.Unlock()
	// Add
	g.members[c] = time.Now()
}

// Remove
func (g *Group) Remove(c *Connection) {
	g.locker.Lock()
	defer g.locker.Unlock()
	// Delete
	delete(g.members, c)
}
