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

package reopener

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func matchEqual(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}

	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}

func Test_labelMatch(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		label string
		match []string
	}{
		"no match": {
			label: "No Match",
			match: nil,
		},
		"http url match": {
			label: "http://github.com/ianlewis/todos/issues/123",
			match: []string{
				"http://github.com/ianlewis/todos/issues/123",
				"http://github.com/ianlewis/todos/issues/",
				"http://",
				"ianlewis",
				"todos",
				"123",
			},
		},
		"https url match": {
			label: "https://github.com/ianlewis/todos/issues/123",
			match: []string{
				"https://github.com/ianlewis/todos/issues/123",
				"https://github.com/ianlewis/todos/issues/",
				"https://",
				"ianlewis",
				"todos",
				"123",
			},
		},
		"no scheme url match": {
			label: "github.com/ianlewis/todos/issues/123",
			match: []string{
				"github.com/ianlewis/todos/issues/123",
				"github.com/ianlewis/todos/issues/",
				"",
				"ianlewis",
				"todos",
				"123",
			},
		},
		"# match": {
			label: "#123",
			match: []string{
				"#123",
				"#",
				"",
				"",
				"",
				"123",
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			if got, want := labelMatch.FindStringSubmatch(tc.label), tc.match; !matchEqual(got, want) {
				t.Errorf("unexpected match (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
