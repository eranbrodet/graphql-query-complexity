/*
This file adapted from
https://github.com/graph-gophers/graphql-go/blob/6d71ad7559729f427b045403dddd7bdeb4ecac3b/internal/common/literals.go.

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

package common

import (
	"text/scanner"

	"github.com/graph-gophers/graphql-go/types"
)

// ParseLiteral returns an interface value
func ParseLiteral(l *Lexer, constOnly bool) types.Value {
	loc := l.Location()
	switch l.Peek() {
	case '$':
		if constOnly {
			l.SyntaxError("variable not allowed")
			panic("unreachable")
		}
		l.ConsumeToken('$')
		return &types.Variable{Name: l.ConsumeIdent(), Loc: loc}

	case scanner.Int, scanner.Float, scanner.String, scanner.Ident:
		lit := l.ConsumeLiteral()
		if lit.Type == scanner.Ident && lit.Text == "null" {
			return &types.NullValue{Loc: loc}
		}
		lit.Loc = loc
		return lit
	case '-':
		l.ConsumeToken('-')
		lit := l.ConsumeLiteral()
		lit.Text = "-" + lit.Text
		lit.Loc = loc
		return lit
	case '[':
		l.ConsumeToken('[')
		var list []types.Value
		for l.Peek() != ']' {
			list = append(list, ParseLiteral(l, constOnly))
		}
		l.ConsumeToken(']')
		return &types.ListValue{Values: list, Loc: loc}

	case '{':
		l.ConsumeToken('{')
		var fields []*types.ObjectField
		for l.Peek() != '}' {
			name := l.ConsumeIdentWithLoc()
			l.ConsumeToken(':')
			value := ParseLiteral(l, constOnly)
			fields = append(fields, &types.ObjectField{Name: name, Value: value})
		}
		l.ConsumeToken('}')
		return &types.ObjectValue{Fields: fields, Loc: loc}

	default:
		l.SyntaxError("invalid value")
		panic("unreachable")
	}
}
