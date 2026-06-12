package parser

import (
	"context"
	"testing"
)

func TestLoadModuleFiles(t *testing.T) {
	t.Parallel()

	mod, err := loadModule(context.Background(), minimoduleRoot(t), defaultTestConfig().Ignore)
	if err != nil {
		t.Fatalf("loadModule() error = %v", err)
	}

	paths := make([]string, 0, len(mod.files))
	for _, file := range mod.files {
		paths = append(paths, file.relPath)
	}

	want := []string{
		"pkg/alpha/alpha.go",
		"pkg/beta/beta.go",
		"pkg/beta/beta_test.go",
		"pkg/beta/helper.go",
		"pkg/gamma/gamma.go",
	}
	if len(paths) != len(want) {
		t.Fatalf("files = %#v, want %#v", paths, want)
	}
	for i, path := range want {
		if paths[i] != path {
			t.Fatalf("files[%d] = %q, want %q (all=%#v)", i, paths[i], path, paths)
		}
	}

	for _, excluded := range []string{
		"pkg/mock_foo.go",
		"pkg/beta/beta_string.go",
		"vendor/example.com/vendorpkg/vendor.go",
	} {
		if containsPath(paths, excluded) {
			t.Fatalf("unexpected ignored file in scope: %q", excluded)
		}
	}
}

func TestMatchesIgnore(t *testing.T) {
	t.Parallel()

	patterns := defaultTestConfig().Ignore
	cases := []struct {
		path string
		want bool
	}{
		{path: "vendor/foo.go", want: true},
		{path: "pkg/mock_foo.go", want: true},
		{path: "pkg/foo_string.go", want: true},
		{path: "pkg/beta/beta_string.go", want: true},
		{path: "pkg/alpha/alpha.go", want: false},
	}

	for _, tc := range cases {
		if got := matchesIgnore(tc.path, patterns); got != tc.want {
			t.Fatalf("matchesIgnore(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}
