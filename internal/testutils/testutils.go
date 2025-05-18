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

// Package testutils implements utilities used for automated testing.
package testutils

// Check checks the error and panics if not nil.
func Check(err error) {
	if err != nil {
		panic(err)
	}
}

// Must checks the error and panics if not nil.
func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

// AsPtr returns the value as a pointer.
func AsPtr[T any](i T) *T {
	return &i
}
