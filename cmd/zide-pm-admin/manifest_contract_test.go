package main

import (
	"strings"
	"testing"

	"github.com/LaurenceGuws/zide-mobile-pm/internal/androidprefix"
)

func TestNewAndroidPrefixManifestRuntimeSupportMetadataMatchesAuthority(t *testing.T) {
	doc := newAndroidPrefixManifest(
		"dev",
		"dist/zide-android-dev-prefix.tar.gz",
		androidprefix.ArchiveStats{
			SHA256:   strings.Repeat("a", 64),
			Size:     1,
			Files:    1,
			Dirs:     0,
			Symlinks: 0,
		},
		strings.Repeat("b", 64),
		prefixAudit{},
	)
	md := doc.Artifacts[0].Metadata
	if got, want := md["runtime_support_links"], androidprefix.PrefixArchiveRuntimeSupportLinks(); got != want {
		t.Fatalf("runtime_support_links mismatch\ngot:  %s\nwant: %s", got, want)
	}
	if got, want := md["runtime_support_files"], androidprefix.PrefixArchiveRuntimeSupportFiles(); got != want {
		t.Fatalf("runtime_support_files mismatch\ngot:  %s\nwant: %s", got, want)
	}
}
