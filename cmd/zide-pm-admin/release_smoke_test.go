package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LaurenceGuws/zide-mobile-pm/internal/manifest"
)

func TestMaterializeCatalogSmokeWritesDist(t *testing.T) {
	dist := filepath.Join(t.TempDir(), "dist")
	hash, size, outPath, err := materializeCatalogSmoke(dist)
	if err != nil {
		t.Fatal(err)
	}
	wantPath := filepath.Join(dist, catalogSmokeDistName)
	if outPath != wantPath {
		t.Fatalf("outPath=%s want %s", outPath, wantPath)
	}
	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if int64(len(b)) != size {
		t.Fatalf("size %d vs len %d", size, len(b))
	}
	root, err := moduleRootDir()
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(filepath.Join(root, catalogSmokeSourceRel))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != string(src) {
		t.Fatal("dist payload mismatch")
	}
	if len(hash) != 64 {
		t.Fatalf("hash len %d", len(hash))
	}
}

func TestApplyAndroidDevReleaseEditsValidates(t *testing.T) {
	doc := manifest.Document{
		SchemaVersion: manifest.SchemaVersion,
		Project:       "zide-mobile-pm",
		Platform:      "android",
		Channel:       "dev",
		Artifacts: []manifest.Artifact{{
			Name:    "zide-android-dev-prefix",
			Kind:    "android-prefix-archive",
			Version: "sha256-aaaaaaaaaaaa",
			URL:     "zide-android-dev-prefix.tar.gz",
			SHA256:  strings.Repeat("a", 64),
			Size:    1,
			Metadata: map[string]string{
				"archive_root":          "usr",
				"package_name":          "uk.laurencegouws.zide",
				"prefix":                "/data/data/uk.laurencegouws.zide/files/usr",
				"target_sdk":            "28",
				"provider":              "termux-main",
				"provider_role":         "android-dev-bootstrap",
				"provider_platform":     "android",
				"provider_architecture": "aarch64",
			},
		}},
	}
	if err := doc.Validate(); err != nil {
		t.Fatal(err)
	}
	smokeHash := strings.Repeat("c", 64)
	merged, err := applyAndroidDevReleaseEdits(doc, "zide-android-dev-prefix.tar.gz", smokeHash, 99)
	if err != nil {
		t.Fatal(err)
	}
	if len(merged.Artifacts) != 2 {
		t.Fatalf("artifacts %d", len(merged.Artifacts))
	}
	if merged.Artifacts[1].Kind != "android-test-binary" || merged.Artifacts[1].Name != catalogSmokeArtifactID {
		t.Fatalf("second artifact: %#v", merged.Artifacts[1])
	}
	if merged.Artifacts[1].SHA256 != smokeHash {
		t.Fatal("smoke hash")
	}
	if merged.Artifacts[0].URL != "zide-android-dev-prefix.tar.gz" {
		t.Fatal("archive url rewrite")
	}
}
