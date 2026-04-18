package zidepm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaurenceGuws/zide-mobile-pm/internal/manifest"
)

func TestAvailablePackagesHidesTestBinariesWithoutAndroidHost(t *testing.T) {
	t.Setenv(EnvHostPlatform, "")
	doc := manifest.Document{
		SchemaVersion: manifest.SchemaVersion,
		Project:       "zide-mobile-pm",
		Platform:      "android",
		Channel:       "dev",
		Artifacts: []manifest.Artifact{
			prefixArchiveStub(),
			testBinaryStub("tb-one", "bin/one"),
		},
	}
	source := Source{Location: "file:///tmp/x.json", Document: doc}
	got := AvailablePackages(source)
	if len(got) != 1 || got[0] != DevBaselinePackage {
		t.Fatalf("expected only dev-baseline, got %#v", got)
	}
}

func TestAvailablePackagesListsTestBinariesOnAndroidHost(t *testing.T) {
	t.Setenv(EnvHostPlatform, HostPlatformAndroid)
	doc := manifest.Document{
		SchemaVersion: manifest.SchemaVersion,
		Project:       "zide-mobile-pm",
		Platform:      "android",
		Channel:       "dev",
		Artifacts: []manifest.Artifact{
			prefixArchiveStub(),
			testBinaryStub("tb-b", "bin/b"),
			testBinaryStub("tb-a", "bin/a"),
		},
	}
	source := Source{Location: "file:///tmp/x.json", Document: doc}
	got := AvailablePackages(source)
	want := []string{DevBaselinePackage, "tb-a", "tb-b"}
	if len(got) != len(want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got, want)
		}
	}
}

func TestInstallAndroidTestBinaryWritesPayload(t *testing.T) {
	t.Setenv(EnvHostPlatform, HostPlatformAndroid)
	work := t.TempDir()
	payloadPath := filepath.Join(work, "payload.dat")
	payload := []byte("zide-pm-test-binary-payload\n")
	if err := os.WriteFile(payloadPath, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(payload)
	hexHash := hex.EncodeToString(sum[:])

	doc := manifest.Document{
		SchemaVersion: manifest.SchemaVersion,
		Project:       "zide-mobile-pm",
		Platform:      "android",
		Channel:       "dev",
		Artifacts: []manifest.Artifact{
			testBinaryArtifact("zide-test-payload", "payload.dat", int64(len(payload)), hexHash, "bin/smoke-test"),
		},
	}
	manifestPath := filepath.Join(work, "manifest.json")
	if err := writeManifest(manifestPath, doc); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	source, err := LoadSource(ctx, manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	prefix := filepath.Join(work, "usr")
	cacheDir := filepath.Join(work, "cache")
	res, err := InstallAndroidTestBinary(ctx, source, "zide-test-payload", prefix, cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	if res.FileCount != 1 {
		t.Fatalf("file count: %d", res.FileCount)
	}
	outPath := filepath.Join(prefix, "bin", "smoke-test")
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("payload mismatch")
	}
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("expected executable bits, got %v", info.Mode())
	}
}

func TestInstallAndroidTestBinaryRejectsWithoutHost(t *testing.T) {
	t.Setenv(EnvHostPlatform, "")
	work := t.TempDir()
	doc := manifest.Document{
		SchemaVersion: manifest.SchemaVersion,
		Project:       "zide-mobile-pm",
		Platform:      "android",
		Channel:       "dev",
		Artifacts: []manifest.Artifact{
			testBinaryArtifact("zide-test-payload", "payload.dat", 1, "abc", "bin/x"),
		},
	}
	manifestPath := filepath.Join(work, "manifest.json")
	if err := writeManifest(manifestPath, doc); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	source, err := LoadSource(ctx, manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = InstallAndroidTestBinary(ctx, source, "zide-test-payload", filepath.Join(work, "usr"), filepath.Join(work, "c"))
	if err == nil {
		t.Fatal("expected error without android host platform")
	}
}

func prefixArchiveStub() manifest.Artifact {
	return manifest.Artifact{
		Name:    "pfx",
		Kind:    "android-prefix-archive",
		Version: "1",
		URL:     "x.tar.gz",
		SHA256:  "a" + repeatChar('b', 63),
		Size:    1,
		Metadata: map[string]string{
			"archive_root":          "usr",
			"provider":              "termux-main",
			"provider_role":         "android-dev-bootstrap",
			"provider_platform":     "android",
			"provider_architecture": "aarch64",
		},
	}
}

func testBinaryStub(name, rel string) manifest.Artifact {
	return testBinaryArtifact(name, "http://example.invalid/x", 1, repeatChar('a', 64), rel)
}

func testBinaryArtifact(name, url string, size int64, sha256, rel string) manifest.Artifact {
	return manifest.Artifact{
		Name:    name,
		Kind:    "android-test-binary",
		Version: "1",
		URL:     url,
		SHA256:  sha256,
		Size:    size,
		Metadata: map[string]string{
			"provider":              "termux-main",
			"provider_role":         "android-dev-bootstrap",
			"provider_platform":     "android",
			"provider_architecture": "aarch64",
			"install_relative_path": rel,
		},
	}
}

func repeatChar(c byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}

func writeManifest(path string, doc manifest.Document) error {
	payload, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o644)
}
