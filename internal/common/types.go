/*
This file adapted from
https://github.com/graph-gophers/graphql-go/blob/6d71ad7559729f427b045403dddd7bdeb4ecac3b/internal/common/types.go.

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
	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/types"
)

// ParseType returns an interface Type of a passed in lexer instance
func ParseType(l *Lexer) types.Type {
	t := parseNullType(l)
	if l.Peek() == '!' {
		l.ConsumeToken('!')
		return &types.NonNull{OfType: t}
	}
	return t
}

func parseNullType(l *Lexer) types.Type {
	if l.Peek() == '[' {
		l.ConsumeToken('[')
		ofType := ParseType(l)
		l.ConsumeToken(']')
		return &types.List{OfType: ofType}
	}

	return &types.TypeName{Ident: l.ConsumeIdentWithLoc()}
}

// Resolver returns interface Type
type Resolver func(name string) types.Type

// ResolveType attempts to resolve a type's name against a resolving function.
// This function is used when one needs to check if a TypeName exists in the resolver (typically a Schema).
//
// In the example below, ResolveType would be used to check if the resolving function
// returns a valid type for Dimension:
//
//	type Profile {
//	   picture(dimensions: Dimension): Url
//	}
//
// ResolveType recursively unwraps List and NonNull types until a NamedType is reached.
func ResolveType(t types.Type, resolver Resolver) (types.Type, *errors.QueryError) {
	switch t := t.(type) {
	case *types.List:
		ofType, err := ResolveType(t.OfType, resolver)
		if err != nil {
			return nil, err
		}
		return &types.List{OfType: ofType}, nil
	case *types.NonNull:
		ofType, err := ResolveType(t.OfType, resolver)
		if err != nil {
			return nil, err
		}
		return &types.NonNull{OfType: ofType}, nil
	case *types.TypeName:
		refT := resolver(t.Name)
		if refT == nil {
			err := errors.Errorf("unknown type %q", t.Name)
			err.Rule = "KnownTypeNames"
			err.Locations = []errors.Location{t.Loc}
			return nil, err
		}
		return refT, nil
	default:
		return t, nil
	}
}
