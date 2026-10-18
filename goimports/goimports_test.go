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

package goimports_test

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"

	"piko.sh/goastutil"
	"piko.sh/goastutil/goimports"
)

func TestFormatAST(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		file     *ast.File
		contains []string
		excludes []string
	}{
		{
			name: "prints the package and its declarations",
			file: &ast.File{
				Name:  ast.NewIdent("main"),
				Decls: []ast.Decl{goastutil.FuncDecl("main", nil, nil, goastutil.BlockStmt())},
			},
			contains: []string{"package main", "func main()"},
		},
		{
			name: "adds a missing standard library import",
			file: &ast.File{
				Name: ast.NewIdent("main"),
				Decls: []ast.Decl{goastutil.FuncDecl("main", nil, nil, goastutil.BlockStmt(
					goastutil.ExprStmt(goastutil.CallExpr(goastutil.SelectorExpr("strings", "ToUpper"), goastutil.StrLit("gull"))),
				))},
			},
			contains: []string{`import "strings"`, `strings.ToUpper("gull")`},
		},
		{
			name: "removes an unused import",
			file: &ast.File{
				Name: ast.NewIdent("main"),
				Decls: []ast.Decl{
					&ast.GenDecl{Tok: token.IMPORT, Specs: []ast.Spec{&ast.ImportSpec{Path: goastutil.StrLit("fmt")}}},
					goastutil.FuncDecl("main", nil, nil, goastutil.BlockStmt()),
				},
			},
			contains: []string{"func main()"},
			excludes: []string{`"fmt"`},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			output, err := goimports.FormatAST(token.NewFileSet(), testCase.file)
			require.NoError(t, err)
			for _, fragment := range testCase.contains {
				require.Contains(t, string(output), fragment)
			}
			for _, fragment := range testCase.excludes {
				require.NotContains(t, string(output), fragment)
			}
		})
	}
}
