package testmap

import (
	"path/filepath"
	"testing"
)

func TestParseCoverProfile(t *testing.T) {
	t.Parallel()
	path := filepath.Join("testdata", "coverage.out")
	profiles, err := parseCoverProfile(path)
	if err != nil {
		t.Fatalf("parseCoverProfile() error = %v", err)
	}
	betaPct, ok := matchProfilePct(profiles, "pkg/beta/beta.go")
	if !ok {
		t.Fatal("missing beta.go in profile")
	}
	if betaPct < 66 || betaPct > 67 {
		t.Fatalf("beta coverage = %v, want ~66.7", betaPct)
	}
	helperPct, ok := matchProfilePct(profiles, "pkg/beta/helper.go")
	if !ok || helperPct != 100 {
		t.Fatalf("helper coverage = %v, want 100", helperPct)
	}
}

func TestParseCoverProfileMissingFile(t *testing.T) {
	t.Parallel()
	profiles, err := parseCoverProfile("testdata/missing.out")
	if err != nil {
		t.Fatalf("parseCoverProfile() error = %v", err)
	}
	if profiles != nil {
		t.Fatalf("profiles = %#v, want nil", profiles)
	}
}
