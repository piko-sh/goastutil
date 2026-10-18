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

// Package goimports formats a Go AST with goimports processing.
//
// It is a separate package because goimports links golang.org/x/tools/imports, and with
// it go/types and the module loader, into every program that imports it. Code that only
// builds AST nodes imports piko.sh/goastutil and stays small enough for constrained
// targets such as js/wasm.
package goimports

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"

	"golang.org/x/tools/imports"
)

const (
	// tabWidth is the number of spaces used for tab indentation in Go code.
	tabWidth = 4
)

// FormatAST formats an AST file to Go source code with goimports processing.
//
// Takes fileSet (*token.FileSet) which is the file set.
// Takes file (*ast.File) which is the file to format.
//
// Returns []byte which is the formatted Go source code.
// Returns error when formatting or import processing fails.
func FormatAST(fileSet *token.FileSet, file *ast.File) ([]byte, error) {
	buffer := new(bytes.Buffer)
	printerConfig := printer.Config{
		Mode:     printer.TabIndent | printer.UseSpaces,
		Tabwidth: tabWidth,
	}
	if err := printerConfig.Fprint(buffer, fileSet, file); err != nil {
		return nil, fmt.Errorf("printing Go AST: %w", err)
	}
	return imports.Process("", buffer.Bytes(), nil)
}
