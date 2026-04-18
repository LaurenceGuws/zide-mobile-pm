package zidepm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/LaurenceGuws/zide-mobile-pm/internal/manifest"
)

// InstallAndroidTestBinary downloads a pinned android-test-binary artifact and
// writes it under prefix using install_relative_path from manifest metadata.
// It requires Android catalog mode (ZIDE_PM_HOST_PLATFORM=android) and an
// Android manifest document.
func InstallAndroidTestBinary(ctx context.Context, source Source, packageName string, prefix string, cacheDir string) (InstallResult, error) {
	if strings.TrimSpace(prefix) == "" {
		return InstallResult{}, errors.New("prefix must not be empty")
	}
	if source.Document.Platform != "android" {
		return InstallResult{}, fmt.Errorf("android-test-binary install requires manifest platform android, got %q", source.Document.Platform)
	}
	if !AndroidCatalogActive() {
		return InstallResult{}, fmt.Errorf("android test binaries require %s=%s (got host platform %q)",
			EnvHostPlatform, HostPlatformAndroid, CurrentHostPlatform())
	}
	artifact, err := findTestBinaryArtifact(source, packageName)
	if err != nil {
		return InstallResult{}, err
	}
	rel := artifact.Metadata["install_relative_path"]
	url, err := ResolveURL(source.Location, artifact.URL)
	if err != nil {
		return InstallResult{}, err
	}
	cachePath, err := FetchArtifact(ctx, artifact, url, cacheDir)
	if err != nil {
		return InstallResult{}, err
	}
	prefix = filepath.Clean(prefix)
	target, err := safeJoin(prefix, filepath.ToSlash(rel))
	if err != nil {
		return InstallResult{}, err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return InstallResult{}, err
	}
	if err := copyFile(cachePath, target); err != nil {
		return InstallResult{}, err
	}
	mode := testBinaryFileMode(artifact)
	if err := os.Chmod(target, mode); err != nil {
		return InstallResult{}, err
	}
	return InstallResult{
		Package:       packageName,
		Prefix:        prefix,
		Manifest:      source.Location,
		Provider:      artifact.Metadata["provider"],
		Version:       artifact.Version,
		InstalledPath: target,
		FileCount:     1,
		DirCount:      0,
		SymlinkCount:  0,
	}, nil
}

func findTestBinaryArtifact(source Source, packageName string) (manifest.Artifact, error) {
	for _, a := range source.Document.Artifacts {
		if a.Kind != "android-test-binary" || a.Name != packageName {
			continue
		}
		return a, nil
	}
	return manifest.Artifact{}, fmt.Errorf("unknown android-test-binary %q", packageName)
}

func testBinaryFileMode(a manifest.Artifact) os.FileMode {
	raw := strings.TrimSpace(a.Metadata["unix_mode"])
	if raw == "" {
		return 0o755
	}
	v, err := strconv.ParseUint(raw, 8, 32)
	if err != nil || v == 0 {
		return 0o755
	}
	return os.FileMode(v) & 0o777
}

func testBinaryPackageNames(source Source) []string {
	if source.Document.Platform != "android" {
		return nil
	}
	var names []string
	for _, a := range source.Document.Artifacts {
		if a.Kind == "android-test-binary" {
			names = append(names, a.Name)
		}
	}
	return names
}
