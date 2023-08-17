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

type stackframe struct {
	top   int
	units []interface{}
}

func (s *stackframe) encode() (raw []byte, err error) { return }
func (s *stackframe) decode(raw []byte) (err error)   { return }

// push
func (s *stackframe) push(value interface{}) {
	if s.units == nil {
		s.top = -1
		s.units = make([]interface{}, 0)
	}
	s.units = append(s.units, value)
	s.top++
}

// pop
func (s *stackframe) pop() (value interface{}, ok bool) {
	if s.top < 0 {
		return nil, false
	}
	value = s.units[s.top]
	s.top--
	return
}
