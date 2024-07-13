package vendoring

import (
	// embed is required to be imported to use go:embed.
	_ "embed"
	"regexp"
	"strings"

	"gopkg.in/yaml.v1"

	"github.com/ianlewis/todos/internal/utils"
)

//go:embed vendor.yml
var vendorRaw []byte

// VendorMatcher regular expression that is used to check if a
// path is a vendored directory.
var VendorMatcher *regexp.Regexp

// IsVendor returns if the path is in a vendored directory.
func IsVendor(path string) bool {
	return VendorMatcher.MatchString(path)
}

//nolint:gochecknoinits // init needed to load embedded config.
func init() {
	// TODO(#460): Generate Go code rather than loading YAML at runtime.
	var rawRegex []string
	utils.Check(yaml.Unmarshal(vendorRaw, &rawRegex))

	var startPrefix []string
	var slashPrefix []string
	var noPrefix []string

	for _, s := range rawRegex {
		switch {
		case strings.HasPrefix(s, "^"):
			startPrefix = append(startPrefix, `(?:`+strings.TrimPrefix(s, "^")+`)`)
		case strings.HasPrefix(s, "(^|/)"):
			slashPrefix = append(slashPrefix, `(?:`+strings.TrimPrefix(s, "(^|/)")+`)`)
		default:
			noPrefix = append(noPrefix, `(?:`+s+`)`)
		}
	}

	if len(startPrefix) > 0 {
		noPrefix = append(noPrefix, `(?:^(?:`+strings.Join(startPrefix, "|")+`))`)
	}

	if len(slashPrefix) > 0 {
		noPrefix = append(noPrefix, `(?:(?:^|/)(?:`+strings.Join(slashPrefix, "|")+`))`)
	}

	// fmt.Println(strings.Join(noPrefix, "|"))
	VendorMatcher = utils.Must(regexp.Compile(strings.Join(noPrefix, "|")))
}
