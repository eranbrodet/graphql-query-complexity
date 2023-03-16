/*
This file adapted from
https://github.com/graph-gophers/graphql-go/blob/6d71ad7559729f427b045403dddd7bdeb4ecac3b/internal/common/lexer.go.

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
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/types"
)

type syntaxError string

// Lexer fields
type Lexer struct {
	sc                    *scanner.Scanner
	next                  rune
	comment               bytes.Buffer
	useStringDescriptions bool
}

// Ident Fields
type Ident struct {
	Name string
	Loc  errors.Location
}

// NewLexer creates a new instance of Lexer
func NewLexer(s string, useStringDescriptions bool) *Lexer {
	sc := &scanner.Scanner{
		Mode: scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanStrings,
	}
	sc.Init(strings.NewReader(s))

	l := Lexer{sc: sc, useStringDescriptions: useStringDescriptions}
	l.sc.Error = l.CatchScannerError

	return &l
}

// CatchSyntaxError provides error checking for Lexer instance
func (l *Lexer) CatchSyntaxError(f func()) (errRes *errors.QueryError) {
	defer func() {
		if err := recover(); err != nil {
			if err, ok := err.(syntaxError); ok {
				errRes = errors.Errorf("syntax error: %s", err)
				errRes.Locations = []errors.Location{l.Location()}
				return
			}
			panic(err)
		}
	}()

	f()
	return
}

// Peek looks as next token
func (l *Lexer) Peek() rune {
	return l.next
}

// ConsumeWhitespace consumes whitespace and tokens equivalent to whitespace (e.g. commas and comments).
//
// Consumed comment characters will build the description for the next type or field encountered.
// The description is available from `DescComment()`, and will be reset every time `ConsumeWhitespace()` is
// executed unless l.useStringDescriptions is set.
func (l *Lexer) ConsumeWhitespace() {
	l.comment.Reset()
	for {
		l.next = l.sc.Scan()

		if l.next == ',' {
			// Similar to white space and line terminators, commas (',') are used to improve the
			// legibility of source text and separate lexical tokens but are otherwise syntactically and
			// semantically insignificant within GraphQL documents.
			//
			// http://facebook.github.io/graphql/draft/#sec-Insignificant-Commas
			continue
		}

		if l.next == '#' {
			// GraphQL source documents may contain single-line comments, starting with the '#' marker.
			//
			// A comment can contain any Unicode code point except `LineTerminator` so a comment always
			// consists of all code points starting with the '#' character up to but not including the
			// line terminator.
			l.consumeComment()
			continue
		}

		break
	}
}

// consumeDescription optionally consumes a description based on the June 2018 graphql spec if any are present.
//
// Single quote strings are also single line. Triple quote strings can be multi-line. Triple quote strings
// whitespace trimmed on both ends.
// If a description is found, consume any following comments as well
//
// http://facebook.github.io/graphql/June2018/#sec-Descriptions
func (l *Lexer) consumeDescription() string {
	// If the next token is not a string, we don't consume it
	if l.next != scanner.String {
		return ""
	}
	// Triple quote string is an empty "string" followed by an open quote due to the way the parser treats strings as one token
	var desc string
	if l.sc.Peek() == '"' {
		desc = l.consumeTripleQuoteComment()
	} else {
		desc = l.consumeStringComment()
	}
	l.ConsumeWhitespace()
	return desc
}

// ConsumeIdent consumes identifer
func (l *Lexer) ConsumeIdent() string {
	name := l.sc.TokenText()
	l.ConsumeToken(scanner.Ident)
	return name
}

// ConsumeIdentWithLoc consumes indentifer and returns Ident with location
func (l *Lexer) ConsumeIdentWithLoc() types.Ident {
	loc := l.Location()
	name := l.sc.TokenText()
	l.ConsumeToken(scanner.Ident)
	return types.Ident{Name: name, Loc: loc}
}

// ConsumeKeyword consumes keyword
func (l *Lexer) ConsumeKeyword(keyword string) {
	if l.next != scanner.Ident || l.sc.TokenText() != keyword {
		l.SyntaxError(fmt.Sprintf("unexpected %q, expecting %q", l.sc.TokenText(), keyword))
	}
	l.ConsumeWhitespace()
}

// ConsumeLiteral consumes literal and returns primitive value
func (l *Lexer) ConsumeLiteral() *types.PrimitiveValue {
	lit := &types.PrimitiveValue{Type: l.next, Text: l.sc.TokenText()}
	l.ConsumeWhitespace()
	return lit
}

// ConsumeToken consumes tokens
func (l *Lexer) ConsumeToken(expected rune) {
	if l.next != expected {
		l.SyntaxError(fmt.Sprintf("unexpected %q, expecting %s", l.sc.TokenText(), scanner.TokenString(expected)))
	}
	l.ConsumeWhitespace()
}

// DescComment consumes comment
func (l *Lexer) DescComment() string {
	comment := l.comment.String()
	desc := l.consumeDescription()
	if l.useStringDescriptions {
		return desc
	}
	return comment
}

// SyntaxError panics if inputer message has error
func (l *Lexer) SyntaxError(message string) {
	panic(syntaxError(message))
}

// Location returns location
func (l *Lexer) Location() errors.Location {
	return errors.Location{
		Line:   l.sc.Line,
		Column: l.sc.Column,
	}
}

func (l *Lexer) consumeTripleQuoteComment() string {
	l.next = l.sc.Next()
	if l.next != '"' {
		panic("consumeTripleQuoteComment used in wrong context: no third quote?")
	}

	var buf bytes.Buffer
	var numQuotes int
	for {
		l.next = l.sc.Next()
		if l.next == '"' {
			numQuotes++
		} else {
			numQuotes = 0
		}
		buf.WriteRune(l.next)
		if numQuotes == 3 || l.next == scanner.EOF {
			break
		}
	}
	val := buf.String()
	val = val[:len(val)-numQuotes]
	return blockString(val)
}

func (l *Lexer) consumeStringComment() string {
	val, err := strconv.Unquote(l.sc.TokenText())
	if err != nil {
		panic(err)
	}
	return val
}

// consumeComment consumes all characters from `#` to the first encountered line terminator.
// The characters are appended to `l.comment`.
func (l *Lexer) consumeComment() {
	if l.next != '#' {
		panic("consumeComment used in wrong context")
	}

	// TODO: count and trim whitespace so we can dedent any following lines.
	if l.sc.Peek() == ' ' {
		l.sc.Next()
	}

	if l.comment.Len() > 0 {
		l.comment.WriteRune('\n')
	}

	for {
		next := l.sc.Next()
		if next == '\r' || next == '\n' || next == scanner.EOF {
			break
		}
		l.comment.WriteRune(next)
	}
}

// CatchScannerError handles errors during message scanning
func (l *Lexer) CatchScannerError(_ *scanner.Scanner, msg string) {
	l.SyntaxError(msg)
}
