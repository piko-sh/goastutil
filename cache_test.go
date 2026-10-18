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
	"fmt"
	"go/ast"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cachedEntries[T any](cache *boundedCache[T]) (entries int, keyBytes int) {
	cache.entries.Range(func(key, _ any) bool {
		entries++
		keyBytes += len(key.(string))
		return true
	})
	return entries, keyBytes
}

func TestBoundedCacheStaysWithinLimits(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		keySize int
	}{
		{name: "short keys reach the entry limit", keySize: 16},
		{name: "long keys reach the byte limit", keySize: 1024},
		{name: "oversized keys are never kept", keySize: dynamicCacheMaxKeyBytes + 1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var cache boundedCache[int]
			padding := strings.Repeat("x", testCase.keySize)
			for index := range 3 * dynamicCacheMaxEntries {
				cache.Store(fmt.Sprintf("%d_%s", index, padding), index)

				require.LessOrEqual(t, cache.count.Load(), int64(dynamicCacheMaxEntries))
				require.LessOrEqual(t, cache.keyBytes.Load(), int64(dynamicCacheMaxBytes))
				if index%512 == 0 {
					entries, keyBytes := cachedEntries(&cache)
					require.Equal(t, int64(entries), cache.count.Load(), "the entry counter drifted from the map")
					require.Equal(t, int64(keyBytes), cache.keyBytes.Load(), "the byte counter drifted from the map")
				}
			}
			if testCase.keySize > dynamicCacheMaxKeyBytes {
				entries, _ := cachedEntries(&cache)
				assert.Zero(t, entries)
			}
		})
	}
}

func TestBoundedCacheKeepsTheFirstValue(t *testing.T) {
	t.Parallel()

	var cache boundedCache[*ast.Ident]
	first := &ast.Ident{Name: "first"}

	assert.Same(t, first, cache.Store("name", first))
	assert.Same(t, first, cache.Store("name", &ast.Ident{Name: "second"}))
	assert.Same(t, first, cache.LoadOrStore("name", newIdent))

	cached, ok := cache.Load("name")
	require.True(t, ok)
	assert.Same(t, first, cached)
}

func TestBoundedCacheLoadOrStoreCopiesTheKey(t *testing.T) {
	t.Parallel()

	var cache boundedCache[*ast.Ident]
	source := "prefix_identifier_suffix"
	name := source[len("prefix_") : len(source)-len("_suffix")]

	identifier := cache.LoadOrStore(name, newIdent)
	assert.Equal(t, "identifier", identifier.Name)
	assert.NotSame(t, unsafe.StringData(name), unsafe.StringData(identifier.Name),
		"the cached identifier must not share memory with the caller's string")

	cache.entries.Range(func(key, _ any) bool {
		assert.Same(t, unsafe.StringData(identifier.Name), unsafe.StringData(key.(string)),
			"the key and the identifier should share one copy")
		return true
	})
}

func TestBoundedCacheBypassesOversizedKeys(t *testing.T) {
	t.Parallel()

	var cache boundedCache[*ast.Ident]
	name := strings.Repeat("x", dynamicCacheMaxKeyBytes+1)

	first := cache.LoadOrStore(name, newIdent)
	second := cache.LoadOrStore(name, newIdent)
	assert.Equal(t, name, first.Name)
	assert.NotSame(t, first, second)

	_, ok := cache.Load(name)
	assert.False(t, ok)
	entries, _ := cachedEntries(&cache)
	assert.Zero(t, entries)
}

func TestBoundedCacheClear(t *testing.T) {
	t.Parallel()

	var cache boundedCache[int]
	cache.Store("one", 1)
	cache.Store("two", 2)

	cache.Clear()

	_, ok := cache.Load("one")
	assert.False(t, ok)
	assert.Zero(t, cache.count.Load())
	assert.Zero(t, cache.keyBytes.Load())
	assert.Equal(t, 3, cache.Store("one", 3))
}

func TestDynamicCachesConcurrentReset(t *testing.T) {
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() {
			for index := range 2000 {
				name := fmt.Sprintf("cache_%d", index)
				assert.Equal(t, name, CachedIdent(name).Name)
				assert.Equal(t, `"`+name+`"`, StrLit(name).Value)
				assert.Equal(t, "[]*pkg."+name, ASTToTypeString(TypeStringToAST("[]*pkg."+name)))
				if index%19 == 0 {
					ResetDynamicCaches()
				}
			}
		})
	}
	workers.Wait()
}
