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

//go:build !windows

package walker

import (
	"testing"

	"github.com/ianlewis/todos/internal/testutils"
)

func Test_isHidden(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		hidden bool
	}{
		{
			name:   "foo.bar",
			hidden: false,
		},
		{
			name:   "some/path/foo.bar",
			hidden: false,
		},
		{
			name:   ".foo",
			hidden: true,
		},
		{
			name:   "some/path/.foo",
			hidden: true,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// NOTE: For *nix we don't actually have to create the file.
			if got, want := testutils.Must(isHidden(tc.name)), tc.hidden; got != want {
				t.Errorf("unexpected result for %q, got: %v, want: %v", tc.name, got, want)
			}
		})
	}
}
