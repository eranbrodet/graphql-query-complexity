/*
This file adapted from
https://github.com/graph-gophers/graphql-go/blob/6d71ad7559729f427b045403dddd7bdeb4ecac3b/internal/common/blockstring.go.

Copyright (c) 2016 Richard Musiol. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

/* Original file below */

// MIT License
//
// Copyright (c) 2019 GraphQL Contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// This implementation has been adapted from the graphql-js reference implementation
// https://github.com/graphql/graphql-js/blob/5eb7c4ded7ceb83ac742149cbe0dae07a8af9a30/src/language/blockString.js
// which is released under the MIT License above.

// Package common adapted from https://github.com/graph-gophers/graphql-go/blob/master/internal/common.
package common

import (
	"strings"
)

// Produces the value of a block string from its parsed raw value, similar to
// CoffeeScript's block string, Python's docstring trim or Ruby's strip_heredoc.
//
// This implements the GraphQL spec's BlockStringValue() static algorithm.
func blockString(raw string) string {
	lines := strings.Split(raw, "\n")

	// Remove common indentation from all lines except the first (which has none)
	ind := blockStringIndentation(lines)
	if ind > 0 {
		for i := 1; i < len(lines); i++ {
			l := lines[i]
			if len(l) < ind {
				lines[i] = ""
				continue
			}
			lines[i] = l[ind:]
		}
	}

	// Remove leading and trailing blank lines
	trimStart := 0
	for i := 0; i < len(lines) && isBlank(lines[i]); i++ {
		trimStart++
	}
	lines = lines[trimStart:]
	trimEnd := 0
	for i := len(lines) - 1; i > 0 && isBlank(lines[i]); i-- {
		trimEnd++
	}
	lines = lines[:len(lines)-trimEnd]

	return strings.Join(lines, "\n")
}

func blockStringIndentation(lines []string) int {
	var commonIndent *int
	for i := 1; i < len(lines); i++ {
		l := lines[i]
		indent := leadingWhitespace(l)
		if indent == len(l) {
			// don't consider blank/empty lines
			continue
		}
		if indent == 0 {
			return 0
		}
		if commonIndent == nil || indent < *commonIndent {
			commonIndent = &indent
		}
	}
	if commonIndent == nil {
		return 0
	}
	return *commonIndent
}

func isBlank(s string) bool {
	return len(s) == 0 || leadingWhitespace(s) == len(s)
}

func leadingWhitespace(s string) int {
	i := 0
	for _, r := range s {
		if r != '\t' && r != ' ' {
			break
		}
		i++
	}
	return i
}
