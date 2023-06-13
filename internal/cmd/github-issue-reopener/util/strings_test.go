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

package util

import "testing"

func strP(s string) *string {
	return &s
}

func TestMustString(t *testing.T) {
	testCases := map[string]struct {
		strP     *string
		expected string
	}{
		"non-empty": {
			strP:     strP("value"),
			expected: "value",
		},
		"empty": {
			strP:     strP(""),
			expected: "",
		},
		"nil": {
			strP:     nil,
			expected: "",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			if got, want := MustString(tc.strP), tc.expected; got != want {
				t.Errorf("unexpected value, got: %q, want: %q", got, want)
			}
		})
	}
}

func TestFirstString(t *testing.T) {
	testCases := map[string]struct {
		str      []string
		expected string
	}{
		"first value": {
			str:      []string{"value", "", ""},
			expected: "value",
		},
		"last value": {
			str:      []string{"", "", "", "", "value"},
			expected: "value",
		},
		"multiple values": {
			str:      []string{"", "value1", "", "", "value2"},
			expected: "value1",
		},
		"nil": {
			str:      nil,
			expected: "",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			if got, want := FirstString(tc.str...), tc.expected; got != want {
				t.Errorf("unexpected value, got: %q, want: %q", got, want)
			}
		})
	}
}
