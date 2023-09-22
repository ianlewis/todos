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

package main

import (
	"fmt"
	"strings"

	"github.com/ianlewis/todos/internal/scanner"
)

func main() {
	fmt.Println("# Supported Languages")
	fmt.Println("")
	fmt.Printf("%d languages are currently supported.\n", len(scanner.LanguagesConfig))
	fmt.Println("")

	fmt.Println("| File type | Supported comments |")
	fmt.Println("| -- | -- |")
	for lang, config := range scanner.LanguagesConfig {
		var supported []string
		for _, c := range config.LineCommentStart {
			supported = append(supported, fmt.Sprintf("`%s`", c))
		}
		if config.MultilineComment.Start != "" {
			s := fmt.Sprintf("`%s %s`", config.MultilineComment.Start, config.MultilineComment.End)
			supported = append(supported, s)
		}

		fmt.Printf("| %s | %s |\n", lang, strings.Join(supported, ", "))
	}
}
