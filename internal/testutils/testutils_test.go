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

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var errTest = errors.New("error")

func TestCheck(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpected panic: %v", r)
			}
		}()
		Check(nil)
	})

	t.Run("fail", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			got, ok := r.(error)
			want := errTest
			if !ok {
				t.Errorf("expected panic, got: %v, want: %v", r, want)
			}
			if !errors.Is(got, want) {
				t.Errorf("expected panic, got: %v, want: %v", got, want)
			}
		}()
		Check(errTest)
	})
}

func TestMust(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpected panic: %v", r)
			}
		}()
		if got, want := Must("test", nil), "test"; got != want {
			t.Errorf("unexpected return value, got: %v, want: %v", got, want)
		}
	})

	t.Run("fail", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			got, ok := r.(error)
			want := errTest
			if !ok {
				t.Errorf("expected panic, got: %v, want: %v", r, want)
			}
			if !errors.Is(got, want) {
				t.Errorf("expected panic, got: %v, want: %v", got, want)
			}
		}()
		Must("test", errTest)
	})
}

func TestAsPtr(t *testing.T) {
	t.Parallel()

	t.Run("int", func(t *testing.T) {
		t.Parallel()

		v := AsPtr(123)
		if got, want := *v, 123; got != want {
			t.Errorf("unexpected return value, got: %v, want: %v", got, want)
		}
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		v := AsPtr("test")
		if got, want := *v, "test"; got != want {
			t.Errorf("unexpected return value, got: %v, want: %v", got, want)
		}
	})

	t.Run("struct", func(t *testing.T) {
		t.Parallel()

		// NOTE: can't use struct{}{} since it's a constant value.
		// NOTE: the struct is copied as it's passed to AsPtr by value.
		input := struct{ Foo int }{}
		got := AsPtr(input)
		want := &input
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected return value (-want, +got)\n%s", diff)
		}
	})
}
