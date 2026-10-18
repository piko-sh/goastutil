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
	"sync/atomic"
)

const (
	// dynamicCacheMaxEntries bounds how many entries each dynamic cache holds before it is
	// emptied.
	dynamicCacheMaxEntries = 4096

	// dynamicCacheMaxBytes bounds the key text each dynamic cache holds before it is
	// emptied.
	dynamicCacheMaxBytes = 1 << 20

	// dynamicCacheMaxKeyBytes is the longest key a dynamic cache accepts. Longer keys bypass
	// the cache, so one large value cannot take the whole byte budget.
	dynamicCacheMaxKeyBytes = 4 << 10
)

// boundedCache is a concurrent string-keyed cache that empties itself when it passes
// dynamicCacheMaxEntries entries or dynamicCacheMaxBytes of key text.
type boundedCache[T any] struct {
	// entries maps each cached key to its value of type T.
	entries sync.Map

	// count is the number of entries stored since the cache was last emptied.
	count atomic.Int64

	// keyBytes is the total length of the keys stored since the cache was last emptied.
	keyBytes atomic.Int64
}

// Load returns the value cached for key.
//
// Takes key (string) which is the key to look up.
//
// Returns T which is the cached value, or the zero value when key is not cached.
// Returns bool which is true when key was cached.
func (cache *boundedCache[T]) Load(key string) (T, bool) {
	var zero T
	if len(key) > dynamicCacheMaxKeyBytes {
		return zero, false
	}
	value, ok := cache.entries.Load(key)
	if !ok {
		return zero, false
	}
	typed, isT := value.(T)
	return typed, isT
}

// LoadOrStore returns the value cached for key, building and caching one on a miss.
//
// Takes key (string) which is the key to look up.
// Takes newValue (func(string) T) which builds the value to cache from the copied key.
//
// Returns T which is the cached value, or the built value when key bypasses the cache.
func (cache *boundedCache[T]) LoadOrStore(key string, newValue func(key string) T) T {
	if value, ok := cache.Load(key); ok {
		return value
	}
	if len(key) > dynamicCacheMaxKeyBytes {
		return newValue(key)
	}
	key = strings.Clone(key)
	return cache.insert(key, newValue(key))
}

// Store caches value under a copy of key unless another value is already cached.
//
// Takes key (string) which is the key to cache value under.
// Takes value (T) which is the value to cache.
//
// Returns T which is the value cached for key: an earlier value when one was already
// stored, value otherwise. A key longer than dynamicCacheMaxKeyBytes is not cached and
// value is returned.
func (cache *boundedCache[T]) Store(key string, value T) T {
	if len(key) > dynamicCacheMaxKeyBytes {
		return value
	}
	return cache.insert(strings.Clone(key), value)
}

// Clear empties the cache.
//
// Values already returned stay valid.
func (cache *boundedCache[T]) Clear() {
	cache.entries.Clear()
	cache.count.Store(0)
	cache.keyBytes.Store(0)
}

// insert caches value under key, first emptying the cache when the new entry would pass
// either limit.
//
// Takes key (string) which is a key the cache may keep, already copied by the caller.
// Takes value (T) which is the value to cache.
//
// Returns T which is the value cached for key, which is an earlier value when another
// call stored one first.
func (cache *boundedCache[T]) insert(key string, value T) T {
	keyLength := int64(len(key))
	if cache.count.Load() >= dynamicCacheMaxEntries || cache.keyBytes.Load()+keyLength > dynamicCacheMaxBytes {
		cache.Clear()
	}
	actual, loaded := cache.entries.LoadOrStore(key, value)
	if !loaded {
		cache.count.Add(1)
		cache.keyBytes.Add(keyLength)
		return value
	}
	if previous, isT := actual.(T); isT {
		return previous
	}
	return value
}
