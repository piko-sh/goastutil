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

// Package goastutil provides Go AST construction, type-string parsing, and type
// classification utilities.
//
// It offers AST node builders for programmatic code generation, type string parsing and
// formatting with package qualification support, and helpers for recognising primitive
// and built-in type names. It also checks Go internal package accessibility and sanitises
// user-controlled names into legal Go identifiers.
//
// The package links neither go/types nor golang.org/x/tools/imports, so programs that
// only build AST nodes stay small, including on js/wasm. Classifying go/types types lives
// in piko.sh/goastutil/typecheck, and goimports formatting in
// piko.sh/goastutil/goimports.
//
// # Usage
//
// Parse a type string and convert it back:
//
//	expr := goastutil.TypeStringToAST("map[string]*User")
//	str := goastutil.ASTToTypeString(expr, "models")
//
// Build AST nodes programmatically:
//
//	call := goastutil.CallExpr(
//	    goastutil.SelectorExpr("fmt", "Sprintf"),
//	    goastutil.StrLit("hello %s"),
//	    goastutil.CachedIdent("name"),
//	)
//
// # Thread safety
//
// The identifier, literal and parsed type caches (CachedIdent, StrLit, TypeStringToAST)
// are safe for concurrent use, as is ResetDynamicCaches. The static identifier cache is
// populated at init time and is read-only thereafter. RegisterIdent must only be called
// during package initialisation.
//
// # Performance
//
// The package caches pre-parsed AST expressions for primitive types and reuses a shared
// token.FileSet to minimise allocations in hot paths. The two-tier identifier cache
// (static map + dynamic cache) avoids repeated heap allocations for commonly occurring
// identifiers.
//
// Each dynamic cache holds about 4096 entries and 1 MiB of key text, and is emptied when
// it passes either, so a long-lived process does not grow without bound. Keys longer than
// 4 KiB are not cached. Because a cache can be emptied at any time, a repeated call may
// return a different node than the one before.
package goastutil
