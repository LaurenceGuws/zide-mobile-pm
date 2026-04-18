package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LaurenceGuws/zide-mobile-pm/internal/manifest"
)

const (
	catalogSmokeSourceRel  = "assets/zide-android-catalog-smoke.sh"
	catalogSmokeDistName   = "zide-android-catalog-smoke.sh"
	catalogSmokeArtifactID = "zide-android-catalog-smoke"
)

func moduleRootDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", cwd)
		}
		dir = parent
	}
}

func catalogSmokeReleaseAssetPath() string {
	return filepath.Join("dist", catalogSmokeDistName)
}

func materializeCatalogSmoke(distDir string) (sha256hex string, size int64, distPath string, err error) {
	root, err := moduleRootDir()
	if err != nil {
		return "", 0, "", err
	}
	srcPath := filepath.Join(root, catalogSmokeSourceRel)
	payload, err := os.ReadFile(srcPath)
	if err != nil {
		return "", 0, "", fmt.Errorf("catalog smoke source: %w", err)
	}
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		return "", 0, "", err
	}
	distPath = filepath.Join(distDir, catalogSmokeDistName)
	if err := os.WriteFile(distPath, payload, 0o644); err != nil {
		return "", 0, "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), int64(len(payload)), distPath, nil
}

func applyAndroidDevReleaseEdits(doc manifest.Document, archiveAssetBaseName, smokeHash string, smokeSize int64) (manifest.Document, error) {
	for i := range doc.Artifacts {
		if doc.Artifacts[i].Kind == "android-prefix-archive" {
			doc.Artifacts[i].URL = archiveAssetBaseName
		}
	}
	version := "sha256-" + smokeHash[:12]
	doc.Artifacts = append(doc.Artifacts, manifest.Artifact{
		Name:    catalogSmokeArtifactID,
		Kind:    "android-test-binary",
		Version: version,
		URL:     catalogSmokeDistName,
		SHA256:  smokeHash,
		Size:    smokeSize,
		Metadata: map[string]string{
			"provider":                "termux-main",
			"provider_role":           "android-dev-bootstrap",
			"provider_platform":       "android",
			"provider_architecture":   "aarch64",
			"install_relative_path":   "libexec/zide-pm/zide-android-catalog-smoke.sh",
			"unix_mode":               "0755",
		},
		Limitations: []string{
			"Development snapshot payload for zide-pm android-test-binary pull/install validation only.",
		},
	})
	doc.Notes = append(doc.Notes,
		"Artifact URLs in this release manifest are relative to the manifest location.",
		"Includes android-test-binary "+catalogSmokeArtifactID+" for Android catalog mode (ZIDE_PM_HOST_PLATFORM=android).",
	)
	if err := doc.Validate(); err != nil {
		return manifest.Document{}, err
	}
	return doc, nil
}

func writeAndroidDevReleaseManifest(prefixManifestPath, releaseManifestPath, archiveAssetBaseName string) error {
	hash, size, _, err := materializeCatalogSmoke(filepath.Dir(releaseManifestPath))
	if err != nil {
		return err
	}
	doc, err := manifest.Load(prefixManifestPath)
	if err != nil {
		return err
	}
	if err := doc.Validate(); err != nil {
		return err
	}
	doc, err = applyAndroidDevReleaseEdits(doc, archiveAssetBaseName, hash, size)
	if err != nil {
		return err
	}
	return writeManifest(releaseManifestPath, doc)
}
