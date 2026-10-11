// Copyright 2026 PolitePixels Limited
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

// This project stands against fascism, authoritarianism, and all forms of
// oppression. We built this to empower people, not to enable those who would
// strip others of their rights and dignity.

package goastutil

import (
	"strings"
	"sync"
)

const (
	dynamicCacheMaxEntries  = 4096
	dynamicCacheMaxBytes    = 1 << 20
	dynamicCacheMaxKeyBytes = 4 << 10
)

// boundedCache retains one generation of small immutable values. A miss may evict
// the entire generation; callers must not depend on cached pointer identity.
// The insertion lock covers both limits and eviction, including explicit resets.
type boundedCache[T any] struct {
	mu      sync.RWMutex
	entries map[string]T
	bytes   int
}

func (cache *boundedCache[T]) Load(key string) (T, bool) {
	if len(key) > dynamicCacheMaxKeyBytes {
		var zero T
		return zero, false
	}
	cache.mu.RLock()
	value, ok := cache.entries[key]
	cache.mu.RUnlock()
	return value, ok
}

func (cache *boundedCache[T]) Store(key string, value T) T {
	if len(key) > dynamicCacheMaxKeyBytes {
		return value
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if previous, ok := cache.entries[key]; ok {
		return previous
	}
	if len(cache.entries) >= dynamicCacheMaxEntries || cache.bytes+len(key) > dynamicCacheMaxBytes {
		cache.entries = nil
		cache.bytes = 0
	}
	if cache.entries == nil {
		cache.entries = make(map[string]T)
	}
	cache.entries[strings.Clone(key)] = value
	cache.bytes += len(key)
	return value
}

func (cache *boundedCache[T]) Clear() {
	cache.mu.Lock()
	cache.entries = nil
	cache.bytes = 0
	cache.mu.Unlock()
}
