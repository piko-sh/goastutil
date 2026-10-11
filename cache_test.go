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
	"strings"
	"sync"
	"testing"
)

func TestDynamicCacheBounds(t *testing.T) {
	for _, size := range []int{16, 1024, 4097} {
		var cache boundedCache[int]
		for index := 0; index < 10000; index++ {
			key := fmt.Sprintf("%d_%s", index, strings.Repeat("x", size))
			cache.Store(key, index)
			if len(cache.entries) > dynamicCacheMaxEntries || cache.bytes > dynamicCacheMaxBytes {
				t.Fatalf("unbounded cache: %d entries, %d bytes", len(cache.entries), cache.bytes)
			}
		}
		if size > dynamicCacheMaxKeyBytes && len(cache.entries) != 0 {
			t.Fatal("oversized keys retained")
		}
	}
}

func TestDynamicCachesConcurrentReset(t *testing.T) {
	var workers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		workers.Go(func() {
			for index := 0; index < 2000; index++ {
				name := fmt.Sprintf("cache_%d", index)
				if CachedIdent(name).Name != name {
					t.Error("identifier changed")
				}
				if ASTToTypeString(TypeStringToAST("[]*pkg."+name)) != "[]*pkg."+name {
					t.Error("type changed")
				}
				StrLit(name)
				if index%19 == 0 {
					ResetDynamicCaches()
				}
			}
		})
	}
	workers.Wait()
}
