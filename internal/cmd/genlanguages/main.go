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

package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v1"
)

var (
	// charEscape indicates the following character is the escape character.
	charEscape = "character"

	// noEscape indicates characters cannot be escaped. This is the default.
	noEscape = "none"

	// doubleEscape indicates that strings can be escaped with double characters.
	doubleEscape = "double"
)

// MultilineCommentConfig describes multi-line comments.
type MultilineCommentConfig struct {
	Start       string `yaml:"start,omitempty"`
	End         string `yaml:"end,omitempty"`
	AtLineStart bool   `yaml:"at_line_start,omitempty"`
}

// StringConfig is config describing types of string.
type StringConfig struct {
	Start  string `yaml:"start,omitempty"`
	End    string `yaml:"end,omitempty"`
	Escape string `yaml:"escape,omitempty"`
}

// Config is configuration for a language comment scanner.
type Config struct {
	LineCommentStart []string               `yaml:"line_comment_start,omitempty"`
	MultilineComment MultilineCommentConfig `yaml:"multiline_comment,omitempty"`
	Strings          []StringConfig         `yaml:"strings,omitempty"`
}

type scannerStringConfig struct {
	Start      []rune // []rune
	End        []rune // []rune
	EscapeFunc string
}

type scannerConfig struct {
	LineCommentStart            [][]rune
	MultilineCommentStart       []rune
	MultilineCommentEnd         []rune
	MultilineCommentAtLineStart bool
	Strings                     []scannerStringConfig
}

type templateData struct {
	CopyrightYear int
	PackageName   string
	Languages     map[string]*scannerConfig
}

var codeTemplate = `
// Copyright {{.CopyrightYear}} Google LLC
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

package {{.PackageName}}

var LanguagesConfig = map[string]*Config{
	{{range $name, $config := .Languages}}{{$name | printf "%q"}}: {
		LineCommentStart: {{if $config.LineCommentStart}}[][]rune{
			{{range $config.LineCommentStart}}{{ "{" }}{{range .}}
				{{. | printf "%q"}},{{end}}
			},{{end}}
		}{{else}}nil{{end}},
		MultilineCommentStart: {{if $config.MultilineCommentStart}}[]rune{
		{{range $config.MultilineCommentStart}}	{{. | printf "%q"}},
		{{end}}}{{else}}nil{{end}},
		MultilineCommentEnd: {{if $config.MultilineCommentEnd}}[]rune{
		{{range $config.MultilineCommentEnd}}	{{. | printf "%q"}},
		{{end}}}{{else}}nil{{end}},
		MultilineCommentAtLineStart: {{$config.MultilineCommentAtLineStart}},
		Strings: []StringConfig{
			{{range .Strings}}{
				Start: []rune{
					{{range .Start}}{{. | printf "%q"}},
				{{end}}},
				End: []rune{
					{{range .End}}{{. | printf "%q"}},
				{{end}}},
				EscapeFunc: {{.EscapeFunc}},
			},{{end}}
		},
	},
	{{end}}
}
`

func stringToRunes(s string) []rune {
	return []rune(s)
}

func stringsToRunes(s []string) [][]rune {
	var runes [][]rune
	for i := range s {
		runes = append(runes, stringToRunes(s[i]))
	}
	return runes
}

func convertConfig(c *Config) *scannerConfig {
	var c2 scannerConfig
	c2.LineCommentStart = stringsToRunes(c.LineCommentStart)
	c2.MultilineCommentStart = stringToRunes(c.MultilineComment.Start)
	c2.MultilineCommentEnd = stringToRunes(c.MultilineComment.End)
	c2.MultilineCommentAtLineStart = c.MultilineComment.AtLineStart

	for i := range c.Strings {
		var escapeFunc string
		switch {
		case strings.HasPrefix(c.Strings[i].Escape, charEscape):

			parts := strings.Split(c.Strings[i].Escape, " ")
			if len(parts[1]) != 1 {
				panic(fmt.Sprintf("invalid escape character %q", parts[1]))
			}
			escapeFunc = fmt.Sprintf("CharEscape(%q)", []rune(parts[1])[0])
		case c.Strings[i].Escape == doubleEscape:
			escapeFunc = "DoubleEscape"
		case c.Strings[i].Escape == "" || c.Strings[i].Escape == noEscape:
			escapeFunc = "NoEscape"
		default:
			panic(fmt.Sprintf("invalid escape %q", c.Strings[i].Escape))
		}

		c2.Strings = append(c2.Strings, scannerStringConfig{
			Start:      stringToRunes(c.Strings[i].Start),
			End:        stringToRunes(c.Strings[i].End),
			EscapeFunc: escapeFunc,
		})
	}
	return &c2
}

func convertConfigs(c map[string]*Config) map[string]*scannerConfig {
	configs := make(map[string]*scannerConfig)
	for k, v := range c {
		configs[k] = convertConfig(v)
	}
	return configs
}

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("usage: genlanguages <PACKAGE> <PATH>")
		os.Exit(1)
	}

	data := templateData{
		CopyrightYear: time.Now().UTC().Year(),
		PackageName:   args[0],
	}

	rawYaml, err := os.ReadFile(args[1])
	if err != nil {
		panic(err)
	}

	languages := make(map[string]*Config)
	if err := yaml.Unmarshal(rawYaml, &languages); err != nil {
		// NOTE: This shouldn't happen.
		panic(err)
	}
	data.Languages = convertConfigs(languages)

	// Generate the license header.
	tmpl := template.Must(template.New("languages").Parse(codeTemplate))
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		panic(err)
	}
}
