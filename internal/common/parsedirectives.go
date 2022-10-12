/*
This file adapted from
https://github.com/graph-gophers/graphql-go/blob/6d71ad7559729f427b045403dddd7bdeb4ecac3b/internal/common/directive.go.

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

import "github.com/graph-gophers/graphql-go/types"

// ParseDirectives returns slice of DirectivesList
func ParseDirectives(l *Lexer) types.DirectiveList {
	var directives types.DirectiveList
	for l.Peek() == '@' {
		l.ConsumeToken('@')
		d := &types.Directive{}
		d.Name = l.ConsumeIdentWithLoc()
		d.Name.Loc.Column--
		if l.Peek() == '(' {
			d.Arguments = ParseArgumentList(l)
		}
		directives = append(directives, d)
	}
	return directives
}
