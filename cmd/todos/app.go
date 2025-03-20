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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/gobwas/glob"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/encoding/ianaindex"
	"sigs.k8s.io/release-utils/version"

	"github.com/ianlewis/todos/internal/todos"
	"github.com/ianlewis/todos/internal/utils"
	"github.com/ianlewis/todos/internal/walker"
)

const defaultCharset = "UTF-8"

var defaultIgnoreFilenames = []string{".gitignore", ".todosignore"}

// TODO(github.com/urfave/cli/issues/1809): Remove init func when upstream bug is fixed.
//
//nolint:gochecknoinits // init needed needed for global variable.
func init() {
	// Set the HelpFlag to a random name so that it isn't used. `cli` handles
	// the flag with the root command such that it takes a command name argument
	// but we don't use commands.
	// This flag is hidden by the help output.
	// See: #442
	cli.HelpFlag = &cli.BoolFlag{
		// NOTE: Use a random name no one would guess.
		Name:               "d41d8cd98f00b204e980",
		DisableDefaultText: true,
	}
}

// newTODOsApp returns a new `todos` application.
func newTODOsApp() *cli.App {
	defaultOutput := "default"
	gha := os.Getenv("GITHUB_ACTIONS")
	if gha == "true" {
		defaultOutput = "github"
	}

	copyrightNames := []string{
		"Google LLC",
		"Ian Lewis",
	}

	app := &cli.App{
		Name:  filepath.Base(os.Args[0]),
		Usage: "Search for TODOS in code.",
		Flags: []cli.Flag{
			// Flags for functionality are in alphabetical order.
			&cli.BoolFlag{
				Name:               "blame",
				Usage:              "[BETA] attempt to find committer info",
				Value:              false,
				DisableDefaultText: true,
			},
			&cli.StringFlag{
				Name:    "charset",
				Usage:   "character set to use when reading files ('detect' to perform charset detection)",
				Value:   defaultCharset,
				Aliases: []string{"c"},
			},
			&cli.StringSliceFlag{
				Name:  "exclude",
				Usage: "exclude files that match `GLOB`",
			},
			&cli.StringSliceFlag{
				Name:  "exclude-dir",
				Usage: "exclude directories that match `GLOB`",
			},
			&cli.BoolFlag{
				Name:               "exclude-hidden",
				Usage:              "exclude hidden files and directories",
				DisableDefaultText: true,
			},
			&cli.StringSliceFlag{
				Name:  "ignore-file-name",
				Usage: "name of files with ignore patterns (.gitignore format)",
				Value: cli.NewStringSlice(defaultIgnoreFilenames...),
			},
			&cli.BoolFlag{
				Name:               "include-vcs",
				Usage:              "include version control directories (.git, .hg, .svn)",
				DisableDefaultText: true,
			},
			&cli.BoolFlag{
				Name:               "include-generated",
				Usage:              "include generated files",
				Value:              false,
				DisableDefaultText: true,
			},
			&cli.BoolFlag{
				Name:               "include-vendored",
				Usage:              "include vendored directories",
				Value:              false,
				DisableDefaultText: true,
			},
			&cli.StringSliceFlag{
				Name:    "label",
				Usage:   "only output TODOs that match `GLOB`",
				Aliases: []string{"l"},
			},
			&cli.StringFlag{
				Name:    "output",
				Usage:   "output `TYPE` (default, github, json)",
				Value:   defaultOutput,
				Aliases: []string{"o"},
			},
			&cli.StringFlag{
				Name:  "todo-types",
				Usage: "comma separated list of TODO `TYPES`",
				Value: strings.Join(todos.DefaultTypes, ","),
			},

			// Special flags are shown at the end.
			&cli.BoolFlag{
				Name:               "help",
				Usage:              "print this help text and exit",
				Aliases:            []string{"h"},
				DisableDefaultText: true,
			},
			&cli.BoolFlag{
				Name:               "version",
				Usage:              "print version information and exit",
				Aliases:            []string{"v"},
				DisableDefaultText: true,
			},
		},
		ArgsUsage:       "[PATH]...",
		Copyright:       strings.Join(copyrightNames, "\n"),
		HideHelp:        true,
		HideHelpCommand: true,
		Action: func(c *cli.Context) error {
			if c.Bool("help") {
				utils.Check(cli.ShowAppHelp(c))
				return nil
			}

			if c.Bool("version") {
				versionInfo := version.GetVersionInfo()
				copyright := ""
				for _, name := range copyrightNames {
					copyright += "Copyright (c) " + name + "\n"
				}

				_ = utils.Must(fmt.Fprintf(c.App.Writer, `%s %s
%s

%s`, c.App.Name, versionInfo.GitVersion, copyright, versionInfo.String()))
				return nil
			}

			opts, err := walkerOptionsFromContext(c)
			if err != nil {
				return err
			}

			todosFound := false
			todoFunc := opts.TODOFunc
			opts.TODOFunc = func(o *walker.TODORef) error {
				todosFound = true
				return todoFunc(o)
			}

			w := walker.New(opts)
			if w.Walk() {
				return ErrWalk
			}

			if todosFound {
				return ErrTODOsFound
			}

			return nil
		},
		OnUsageError:   OnUsageError,
		ExitErrHandler: ExitErrHandler,
	}

	setupProfiling(app)

	return app
}

func outCLI(w io.Writer) walker.TODOHandler {
	return func(o *walker.TODORef) error {
		if o == nil {
			return nil
		}
		var blameInfo string
		if o.GitUser != nil {
			blameInfo = color.GreenString(o.GitUser.Name)
			if o.GitUser.Email != "" {
				blameInfo += color.GreenString(" <" + o.GitUser.Email + ">")
			}
			blameInfo += color.CyanString(":")
		}
		_ = utils.Must(fmt.Fprintf(w, "%s%s%s%s%s%s\n",
			color.MagentaString(o.FileName),
			color.CyanString(":"),
			color.GreenString(fmt.Sprintf("%d", o.TODO.Line)),
			color.CyanString(":"),
			blameInfo,
			o.TODO.Text,
		))
		return nil
	}
}

func outGithub(w io.Writer) walker.TODOHandler {
	return func(o *walker.TODORef) error {
		if o == nil {
			return nil
		}
		typ := "notice"
		switch o.TODO.Type {
		case "TODO", "HACK", "COMBAK":
			typ = "warning"
		case "FIXME", "XXX", "BUG":
			typ = "error"
		}
		_ = utils.Must(fmt.Fprintf(w, "::%s file=%s,line=%d::%s\n", typ, o.FileName, o.TODO.Line, o.TODO.Text))
		return nil
	}
}

type outUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type outTODO struct {
	// Path is the path to the file where the TODO was found.
	Path string `json:"path"`

	// Type is the todo type, such as "FIXME", "BUG", etc.
	Type string `json:"type"`

	// Text is the full comment text.
	Text string `json:"text"`

	// Label is the label part (the part in parenthesis)
	Label string `json:"label"`

	// Message is the comment message (the part after the parenthesis).
	Message string `json:"message"`

	// Line is the line number where todo was found..
	Line int `json:"line"`

	// CommentLine is the line where the comment starts.
	CommentLine int `json:"comment_line"`

	// GitUser is the committer of the TODO.
	GitUser *outUser `json:"git_user,omitempty"`
}

func outJSON(w io.Writer) walker.TODOHandler {
	return func(o *walker.TODORef) error {
		if o == nil {
			return nil
		}

		out := outTODO{
			Path:        o.FileName,
			Type:        o.TODO.Type,
			Text:        o.TODO.Text,
			Label:       o.TODO.Label,
			Message:     o.TODO.Message,
			Line:        o.TODO.Line,
			CommentLine: o.TODO.CommentLine,
		}
		if o.GitUser != nil {
			out.GitUser = &outUser{
				Name:  o.GitUser.Name,
				Email: o.GitUser.Email,
			}
		}

		b := utils.Must(json.Marshal(out))

		_ = utils.Must(w.Write(b))
		_ = utils.Must(w.Write([]byte("\n")))

		return nil
	}
}

var outTypes = map[string]func(io.Writer) walker.TODOHandler{
	// NOTE: An empty value is treated as the default value.
	"":        outCLI,
	"default": outCLI,
	"github":  outGithub,
	"json":    outJSON,
}

func walkerOptionsFromContext(c *cli.Context) (*walker.Options, error) {
	o := walker.Options{}

	// Valdidate the character set.
	charset := c.String("charset")
	if charset != "detect" {
		if charset == "ISO-8859-1" {
			charset = "UTF-8"
		}
		// See: https://github.com/saintfish/chardet/issues/2
		if charset == "GB-18030" {
			charset = "GB18030"
		}
		e, err := ianaindex.IANA.Encoding(charset)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: unsupported character set: %w", ErrFlagParse, charset, err)
		}
		if e == nil {
			return nil, fmt.Errorf("%w: %s: unsupported character set", ErrFlagParse, charset)
		}
	}
	o.Charset = charset

	for _, gs := range c.StringSlice("exclude") {
		g, err := glob.Compile(gs)
		if err != nil {
			return nil, fmt.Errorf("%w: exclude: %w", ErrFlagParse, err)
		}
		o.ExcludeGlobs = append(o.ExcludeGlobs, g)
	}

	for _, gs := range c.StringSlice("exclude-dir") {
		g, err := glob.Compile(strings.TrimRight(gs, string(os.PathSeparator)))
		if err != nil {
			return nil, fmt.Errorf("%w: exclude-dir: %w", ErrFlagParse, err)
		}
		o.ExcludeDirGlobs = append(o.ExcludeDirGlobs, g)
	}

	o.Blame = c.Bool("blame")

	// File Includes
	o.IncludeGenerated = c.Bool("include-generated")
	o.IncludeHidden = !c.Bool("exclude-hidden")
	o.IncludeVCS = c.Bool("include-vcs")
	o.IncludeVendored = c.Bool("include-vendored")

	o.IgnoreFileNames = c.StringSlice("ignore-file-name")

	// Filters
	for _, label := range c.StringSlice("label") {
		g, err := glob.Compile(label)
		if err != nil {
			return nil, fmt.Errorf("%w: label: %w", ErrFlagParse, err)
		}
		o.LabelGlobs = append(o.LabelGlobs, g)
	}

	outType := c.String("output")
	outFunc, ok := outTypes[outType]
	if !ok {
		return nil, fmt.Errorf("%w: invalid output type: %v", ErrFlagParse, outType)
	}

	o.TODOFunc = outFunc(c.App.Writer)
	o.ErrorFunc = func(err error) error {
		_ = utils.Must(fmt.Fprintf(c.App.ErrWriter, "%s: %v\n", c.App.Name, err))
		return nil
	}

	o.Config = &todos.Config{}

	todoTypesStr := c.String("todo-types")
	if todoTypesStr != "" {
		for _, todoType := range strings.Split(todoTypesStr, ",") {
			o.Config.Types = append(o.Config.Types, strings.TrimSpace(todoType))
		}
	}

	o.Paths = c.Args().Slice()
	if len(o.Paths) == 0 {
		o.Paths = []string{"."}
	}

	return &o, nil
}
