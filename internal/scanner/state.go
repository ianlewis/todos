// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scanner

type state interface {
	stateMustImplement()
}

type stateCode struct{}

func (s *stateCode) stateMustImplement() {}

type stateLineComment struct {
	// index is the index for the type of line comment.
	index int
}

func (s *stateLineComment) stateMustImplement() {}

type stateMultilineComment struct {
	// line is the line of the start of the multi-line comment.
	line int

	// index is the index for the type of multiline comment.
	index int
}

func (s *stateMultilineComment) stateMustImplement() {}

// stateLineCommentOrString implements the special case when strings and line
// comments start with the same character. e.g. Vim Script.
type stateLineCommentOrString struct {
	// lcIndex is the index for the type of line comment.
	lcIndex int

	// sIndex is the index for the type of string.
	sIndex int
}

func (s *stateLineCommentOrString) stateMustImplement() {}

type stateString struct {
	// index is the index for the type of string.
	index int
}

func (s *stateString) stateMustImplement() {}
