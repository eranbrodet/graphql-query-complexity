/*
This file adapted from
https://github.com/graph-gophers/graphql-go/blob/6d71ad7559729f427b045403dddd7bdeb4ecac3b/internal/common/lexer_test.go.

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

package common_test

import (
	"testing"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/graphql-query-complexity/internal/common"
)

type consumeTestCase struct {
	description           string
	definition            string
	expected              string // expected description
	failureExpected       bool
	useStringDescriptions bool
}

// Note that these tests stop as soon as they parse the comments, so even though the rest of the file will fail to parse sometimes, the tests still pass
var consumeTests = []consumeTestCase{{
	description: "no string descriptions allowed in old mode",
	definition: `
# Comment line 1
#Comment line 2
,,,,,, # Commas are insignificant
"New style comments"
type Hello {
	world: String!
}`,
	expected:              "Comment line 1\nComment line 2\nCommas are insignificant",
	useStringDescriptions: false,
}, {
	description: "simple string descriptions allowed in new mode",
	definition: `
# Comment line 1
#Comment line 2
,,,,,, # Commas are insignificant
"New style comments"
type Hello {
	world: String!
}`,
	expected:              "New style comments",
	useStringDescriptions: true,
}, {
	description: "comment after description works",
	definition: `
# Comment line 1
#Comment line 2
,,,,,, # Commas are insignificant
type Hello {
	world: String!
}`,
	expected:              "",
	useStringDescriptions: true,
}, {
	description: "triple quote descriptions allowed in new mode",
	definition: `
# Comment line 1
#Comment line 2
,,,,,, # Commas are insignificant
"""
New style comments
Another line
"""
type Hello {
	world: String!
}`,
	expected:              "New style comments\nAnother line",
	useStringDescriptions: true,
}}

func TestConsume(t *testing.T) {
	for _, test := range consumeTests {
		t.Run(test.description, func(t *testing.T) {
			lex := common.NewLexer(test.definition, test.useStringDescriptions)

			err := lex.CatchSyntaxError(func() { lex.ConsumeWhitespace() })
			if test.failureExpected {
				if err == nil {
					t.Fatalf("schema should have been invalid; comment: %s", lex.DescComment())
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
			}

			if test.expected != lex.DescComment() {
				t.Errorf("wrong description value:\nwant: %q\ngot : %q", test.expected, lex.DescComment())
			}
		})
	}
}

var multilineStringTests = []consumeTestCase{
	{
		description:           "Oneline strings are okay",
		definition:            `"Hello World"`,
		expected:              "",
		failureExpected:       false,
		useStringDescriptions: true,
	},
	{
		description: "Multiline strings are not allowed",
		definition: `"Hello
				 World"`,
		expected:              `graphql: syntax error: literal not terminated (line 1, column 1)`,
		failureExpected:       true,
		useStringDescriptions: true,
	},
}

func TestMultilineString(t *testing.T) {
	for _, test := range multilineStringTests {
		t.Run(test.description, func(t *testing.T) {
			lex := common.NewLexer(test.definition, test.useStringDescriptions)

			err := lex.CatchSyntaxError(func() { lex.ConsumeWhitespace() })
			if test.failureExpected && err == nil {
				t.Fatalf("Test '%s' should fail", test.description)
			} else if test.failureExpected && err != nil {
				if test.expected != err.Error() {
					t.Fatalf("Test '%s' failed with wrong error: '%s'. Error should be: '%s'", test.description, err.Error(), test.expected)
				}
			}

			if !test.failureExpected && err != nil {
				t.Fatalf("Test '%s' failed with error: '%s'", test.description, err.Error())
			}
		})
	}
}
