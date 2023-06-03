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

//go:build windows

package main

import (
	"os"
	"syscall"
	"testing"

	"github.com/ianlewis/todos/internal/testutils"
)

func setHidden(path string) error {
	filenameW, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	err = syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN)
	if err != nil {
		return err
	}

	return nil
}

func Test_isHidden(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		hiddenAttr bool
		expected   bool
	}{
		{
			name:       "foo.bar",
			hiddenAttr: false,
			expected:   false,
		},
		{
			name:       "some/path/foo.bar",
			hiddenAttr: false,
			expected:   false,
		},
		{
			name:       "foo.bar",
			hiddenAttr: true,
			expected:   true,
		},
		{
			name:       "some/path/foo.bar",
			hiddenAttr: true,
			expected:   true,
		},
		{
			name:       ".foo.bar",
			hiddenAttr: false,
			expected:   true,
		},
		{
			name:       "some/path/.foo.bar",
			hiddenAttr: false,
			expected:   true,
		},
		{
			name:       ".foo.bar",
			hiddenAttr: true,
			expected:   true,
		},
		{
			name:       "some/path/.foo.bar",
			hiddenAttr: true,
			expected:   true,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := testutils.Must(os.MkdirTemp("", tc.name))
			path := filepath.Join(dir, tc.name)
			testutils.Check(os.WriteFile(path, "", 0600))
			if tc.hiddenAttr {
				testutils.Check(setHidden(path))
			}
			defer os.RemoveAll(dir)

			if got, want := testutils.Must(isHidden(tc.name)), tc.expected; got != want {
				t.Errorf("unexpected result for %q, got: %v, want: %v", tc.name, got, want)
			}
		})
	}
}
