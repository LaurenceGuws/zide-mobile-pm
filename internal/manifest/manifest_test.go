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
	if got := doc.Artifacts[0].Metadata["prefix"]; got != "/data/data/dev.zide.terminal/files/usr" {
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
		Platform:      "android",
		Channel:       "dev",
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
