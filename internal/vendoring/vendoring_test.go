// Copyright 2024 Google LLC
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

package vendoring

import (
	"testing"
)

func TestIsVendor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "normal_file",
			path:     "internal/file.go",
			expected: false,
		},
		{
			name:     "vendor_dir_exact",
			path:     "vendor/",
			expected: true,
		},
		{
			name:     "minified_js",
			path:     "internal/mini.min.js",
			expected: true,
		},
		{
			name:     "github_workflow",
			path:     ".github/workflows/test.yml",
			expected: false,
		},
		{
			name:     "node_modules",
			path:     "somepackage/node_modules/someotherpackage/somefile.js",
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got, want := IsVendor(tc.path), tc.expected; got != want {
				t.Errorf("IsVendor(%q); got: %v, want: %v", tc.path, got, want)
			}
		})
	}
}
