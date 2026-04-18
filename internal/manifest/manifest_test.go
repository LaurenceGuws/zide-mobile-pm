package manifest

import "testing"

func TestAndroidSkeletonValidates(t *testing.T) {
	doc, err := NewSkeleton("android", "dev")
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := doc.Artifacts[0].Metadata["prefix"]; got != "/data/data/uk.laurencegouws.zide/files/usr" {
		t.Fatalf("unexpected android prefix: %s", got)
	}
}

func TestRejectsUnknownPlatform(t *testing.T) {
	if _, err := NewSkeleton("desktop", "dev"); err == nil {
		t.Fatal("expected unknown platform error")
	}
}

func TestProviderArtifactsRequireProviderMetadata(t *testing.T) {
	doc := Document{
		SchemaVersion: SchemaVersion,
		Project:       "zide-mobile-pm",
		Channel:       "dev",
		Platform:      "android",
		Artifacts: []Artifact{{
			Name:    "bad-prefix",
			Kind:    "android-prefix-archive",
			Version: "0",
			URL:     "dist/bad.tar.gz",
			SHA256:  "abc",
			Size:    1,
		}},
	}
	if err := doc.Validate(); err == nil {
		t.Fatal("expected missing provider metadata error")
	}
}

func TestAndroidTestBinaryRequiresInstallPath(t *testing.T) {
	meta := map[string]string{
		"provider":                "termux-main",
		"provider_role":           "android-dev-bootstrap",
		"provider_platform":       "android",
		"provider_architecture":   "aarch64",
		"install_relative_path":   "bin/smoke",
	}
	base := Artifact{
		Name:     "smoke-bin",
		Kind:     "android-test-binary",
		Version:  "1",
		URL:      "https://example.invalid/bin",
		SHA256:   "abc",
		Size:     1,
		Metadata: meta,
	}
	doc := Document{
		SchemaVersion: SchemaVersion,
		Project:       "zide-mobile-pm",
		Platform:      "android",
		Channel:       "dev",
		Artifacts:     []Artifact{base},
	}
	if err := doc.Validate(); err != nil {
		t.Fatal(err)
	}

	metaNoPath := map[string]string{
		"provider":              "termux-main",
		"provider_role":         "android-dev-bootstrap",
		"provider_platform":     "android",
		"provider_architecture": "aarch64",
	}
	doc.Artifacts = []Artifact{{
		Name: "smoke-bin", Kind: "android-test-binary", Version: "1",
		URL: "https://example.invalid/bin", SHA256: "abc", Size: 1, Metadata: metaNoPath,
	}}
	if err := doc.Validate(); err == nil {
		t.Fatal("expected install_relative_path error")
	}

	metaEscape := map[string]string{
		"provider":                "termux-main",
		"provider_role":           "android-dev-bootstrap",
		"provider_platform":       "android",
		"provider_architecture":   "aarch64",
		"install_relative_path":   "../escape",
	}
	doc.Artifacts = []Artifact{{
		Name: "smoke-bin", Kind: "android-test-binary", Version: "1",
		URL: "https://example.invalid/bin", SHA256: "abc", Size: 1, Metadata: metaEscape,
	}}
	if err := doc.Validate(); err == nil {
		t.Fatal("expected escape path error")
	}

	docIOS := Document{
		SchemaVersion: SchemaVersion,
		Project:       "zide-mobile-pm",
		Platform:      "ios",
		Channel:       "dev",
		Artifacts: []Artifact{{
			Name: "smoke-bin", Kind: "android-test-binary", Version: "1",
			URL: "https://example.invalid/bin", SHA256: "abc", Size: 1, Metadata: meta,
		}},
	}
	if err := docIOS.Validate(); err == nil {
		t.Fatal("expected ios platform rejection for android-test-binary")
	}
}
