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
	"sort"
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/go-enry/go-enry/v2/data"

	"github.com/ianlewis/todos/internal/scanner"
)

type langConfig struct {
	lang   string
	config *scanner.Config
	info   *data.LanguageInfo
}

type langConfigs []langConfig

func (l langConfigs) Len() int {
	return len(l)
}

func (l langConfigs) Less(i, j int) bool {
	return l[i].lang < l[j].lang
}

func (l langConfigs) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func main() {
	var langs langConfigs
	for lang, config := range scanner.LanguagesConfig {
		info, err := enry.GetLanguageInfo(lang)
		if err != nil {
			panic(err)
		}

		langs = append(langs, langConfig{
			lang:   lang,
			config: config,
			info:   &info,
		})
	}
	sort.Sort(langs)

	fmt.Println("# Supported Languages")
	fmt.Println("")
	fmt.Printf("%d languages are currently supported.\n", len(scanner.LanguagesConfig))
	fmt.Println("")

	fmt.Println("| File type | Extension | Supported comments |")
	fmt.Println("| -- | -- | -- |")

	for _, l := range langs {
		var supported []string
		for _, c := range l.config.LineComments {
			supported = append(supported, fmt.Sprintf("`%s`", string(c.Start)))
		}
		for _, c := range l.config.MultilineComments {
			s := fmt.Sprintf("`%s %s`", string(c.Start), string(c.End))
			supported = append(supported, s)
		}

		var extensions []string
		for _, ext := range l.info.Extensions {
			extensions = append(extensions, fmt.Sprintf("`%s`", ext))
		}

		fmt.Printf("| %s | %s | %s |\n", l.lang, strings.Join(extensions, ", "), strings.Join(supported, ", "))
	}
}
