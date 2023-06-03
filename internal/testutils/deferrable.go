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

package testutils

// DeferFunc is a deferrable function.
type DeferFunc func()

// CancelFunc is a cancel function.
type CancelFunc func()

// WithCancel returns a deferrable function that can be cancelled.
func WithCancel(d DeferFunc, f func()) (DeferFunc, CancelFunc) {
	var c bool
	newFunc := func() {
		if !c {
			if f != nil {
				f()
			}
			if d != nil {
				d()
			}
		}
	}
	cancel := func() {
		c = true
	}
	return newFunc, cancel
}
