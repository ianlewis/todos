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
)

func TestCheck(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpected panic: %v", r)
			}
		}()
		Check(nil)
	})

	t.Run("fail", func(t *testing.T) {
		err := errors.New("error")
		defer func() {
			if got, want := recover(), err; got != want {
				t.Errorf("expected panic, got: %v, want: %v", got, want)
			}
		}()
		Check(err)
	})
}

func TestMust(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
		err := errors.New("error")
		defer func() {
			if got, want := recover(), err; got != want {
				t.Errorf("expected panic, got: %v, want: %v", got, want)
			}
		}()
		Must("test", err)
	})
}
