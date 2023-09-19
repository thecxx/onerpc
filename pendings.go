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
	"context"
	"sync"
)

type pendings struct {
	signals sync.Map
}

// Wait
func (p *pendings) Wait(ctx context.Context, seq uint64) (r Message, err error) {
	var (
		signal = make(chan interface{})
	)

	// Save signal for sequence
	p.signals.Store(seq, signal)

	select {
	// Cancel
	case <-ctx.Done():
		err = ctx.Err()
	// Signal
	case x := <-signal:
		switch v := x.(type) {
		// Message
		case Message:
			r = v
		// Error
		case error:
			err = v
		}
	}

	return
}

// Trigger
func (p *pendings) Trigger(seq uint64, r Message, err error) (handled bool) {

	v, ok := p.signals.LoadAndDelete(seq)
	if !ok {
		return false
	}

	// Reply signal
	signal := v.(chan interface{})

	if err != nil {
		signal <- err
	} else {
		signal <- r
	}

	return true
}
