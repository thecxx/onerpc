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
	"time"

	"github.com/govoltron/onerpc/transport"
)

type CanOption interface {

	// SetReadTimeout
	SetReadTimeout(timeout time.Duration)

	// SetWriteTimeout
	SetWriteTimeout(timeout time.Duration)

	// SetIdleTimeout
	SetIdleTimeout(timeout time.Duration)

	// SetMaxLifeTime
	SetMaxLifeTime(timeout time.Duration)

	// SetReaderBufferSize
	SetReaderBufferSize(size int)

	// SetWriterBufferSize
	SetWriterBufferSize(size int)

	// SetBalancer
	SetBalancer(b transport.Balancer)
}

type Option func(setter CanOption)

// WithReadTimeout
func WithReadTimeout(timeout time.Duration) Option {
	return func(setter CanOption) { setter.SetReadTimeout(timeout) }
}

// WithWriteTimeout
func WithWriteTimeout(timeout time.Duration) Option {
	return func(setter CanOption) { setter.SetWriteTimeout(timeout) }
}

// WithIdleTimeout
func WithIdleTimeout(timeout time.Duration) Option {
	return func(setter CanOption) { setter.SetIdleTimeout(timeout) }
}

// WithMaxLifeTime
func WithMaxLifeTime(timeout time.Duration) Option {
	return func(setter CanOption) { setter.SetMaxLifeTime(timeout) }
}

// WithReaderBufferSize
func WithReaderBufferSize(size int) Option {
	return func(setter CanOption) { setter.SetReaderBufferSize(size) }
}

// WithWriterBufferSize
func WithWriterBufferSize(size int) Option {
	return func(setter CanOption) { setter.SetWriterBufferSize(size) }
}

// WithBalancer
func WithBalancer(b transport.Balancer) Option {
	return func(setter CanOption) { setter.SetBalancer(b) }
}
