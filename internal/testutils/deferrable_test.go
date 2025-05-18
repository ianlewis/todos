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

import "testing"

func TestWithCancel(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		t.Parallel()

		var called int

		func() {
			d, _ := WithCancel(nil, func() {
				called++
			})
			defer d()
		}()

		if want, got := 1, called; want != got {
			t.Errorf("unexpected calls, want: %d, got: %d", want, got)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		t.Parallel()

		var called int

		func() {
			d, cancel := WithCancel(nil, func() {
				called++
			})
			defer d()
			cancel()
		}()

		if want, got := 0, called; want != got {
			t.Errorf("unexpected calls, want: %d, got: %d", want, got)
		}
	})

	t.Run("nested", func(t *testing.T) {
		t.Parallel()

		var parentCalled int

		var childCalled int

		func() {
			p := func() {
				parentCalled++
			}
			c := func() {
				childCalled++
			}
			parentDefer, _ := WithCancel(nil, p)

			childDefer, _ := WithCancel(parentDefer, c)
			defer childDefer()
		}()

		if want, got := 1, parentCalled; want != got {
			t.Errorf("unexpected parent calls, want: %d, got: %d", want, got)
		}

		if want, got := 1, childCalled; want != got {
			t.Errorf("unexpected child calls, want: %d, got: %d", want, got)
		}
	})

	t.Run("nested child cancel", func(t *testing.T) {
		t.Parallel()

		var parentCalled int

		var childCalled int

		func() {
			p := func() {
				parentCalled++
			}
			c := func() {
				childCalled++
			}
			parentDefer, _ := WithCancel(nil, p)

			childDefer, cancel := WithCancel(parentDefer, c)
			defer childDefer()
			cancel()
		}()

		if want, got := 0, parentCalled; want != got {
			t.Errorf("unexpected parent calls, want: %d, got: %d", want, got)
		}

		if want, got := 0, childCalled; want != got {
			t.Errorf("unexpected child calls, want: %d, got: %d", want, got)
		}
	})

	t.Run("nested parent cancel", func(t *testing.T) {
		t.Parallel()

		var parentCalled int

		var childCalled int

		func() {
			p := func() {
				parentCalled++
			}
			c := func() {
				childCalled++
			}
			parentDefer, cancel := WithCancel(nil, p)

			childDefer, _ := WithCancel(parentDefer, c)
			defer childDefer()
			cancel()
		}()

		if want, got := 0, parentCalled; want != got {
			t.Errorf("unexpected parent calls, want: %d, got: %d", want, got)
		}

		if want, got := 1, childCalled; want != got {
			t.Errorf("unexpected child calls, want: %d, got: %d", want, got)
		}
	})
}
