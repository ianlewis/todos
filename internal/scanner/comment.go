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

// Comment is a generic Comment implementation.
type Comment struct {
	text string
	line int
}

// Line implements Comment.Line.
func (c *Comment) Line() int {
	return c.line
}

// String implements Comment.String.
func (c *Comment) String() string {
	return c.text
}
