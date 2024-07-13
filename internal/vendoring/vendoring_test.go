package vendoring

import (
	"testing"
)

func TestIsVendor(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "normal_file",
			path:     "internal/file.go",
			expected: false,
		},
		{
			name:     "vendor_dir_exact",
			path:     "vendor/",
			expected: true,
		},
		{
			name:     "minified_js",
			path:     "internal/mini.min.js",
			expected: true,
		},
		{
			name:     "github_workflow",
			path:     ".github/workflows/test.yml",
			expected: false,
		},
		{
			name:     "node_modules",
			path:     "somepackage/node_modules/someotherpackage/somefile.js",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := IsVendor(tc.path), tc.expected; got != want {
				t.Errorf("IsVendor(%q); got: %v, want: %v", tc.path, got, want)
			}
		})
	}
}
