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

package testutils

import "os"

func compareMode(l, r os.FileMode) bool {
	if l&0o600 == 0o600 && r&0o600 == 0o600 {
		return true
	}
	if l&0o400 == 0o400 && r&0o400 == 0o400 {
		return true
	}

	return false
}
