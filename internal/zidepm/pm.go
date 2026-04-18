// Package zidepm implements the user-facing Zide mobile package CLI surface.
package zidepm

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/LaurenceGuws/zide-mobile-pm/internal/manifest"
)

const (
	DefaultAndroidDevManifestURL = "https://github.com/LaurenceGuws/zide-mobile-pm/releases/download/android-dev-2026.04.18.140021/android-dev-prefix.release.manifest.json"
	DevBaselinePackage           = "dev-baseline"
)

type Source struct {
	Location string
	Document manifest.Document
}

type PrefixArtifact struct {
	Artifact manifest.Artifact
	URL      string
}

type InstallResult struct {
	Package       string
	Prefix        string
	Manifest      string
	Provider      string
	Version       string
	InstalledPath string
	FileCount     int
	DirCount      int
	SymlinkCount  int
}

type InstallStamp struct {
	InstalledAt string `json:"installed_at"`
	Package     string `json:"package"`
	Manifest    string `json:"manifest"`
	Artifact    string `json:"artifact"`
	Version     string `json:"version"`
	Provider    string `json:"provider"`
	Files       int    `json:"files"`
	Dirs        int    `json:"dirs"`
	Symlinks    int    `json:"symlinks"`
}

func LoadSource(ctx context.Context, location string) (Source, error) {
	if strings.TrimSpace(location) == "" {
		return Source{}, errors.New("manifest location must not be empty")
	}

	var payload []byte
	var err error
	if IsURL(location) {
		payload, err = readURL(ctx, location)
	} else {
		payload, err = os.ReadFile(filepath.Clean(location))
	}
	if err != nil {
		return Source{}, err
	}

	var doc manifest.Document
	if err := json.Unmarshal(payload, &doc); err != nil {
		return Source{}, err
	}
	if err := doc.Validate(); err != nil {
		return Source{}, err
	}
	return Source{Location: location, Document: doc}, nil
}

func AvailablePackages(source Source) []string {
	var out []string
	if _, err := AndroidPrefixArtifact(source); err == nil {
		out = append(out, DevBaselinePackage)
	}
	if AndroidCatalogActive() {
		tb := testBinaryPackageNames(source)
		sort.Strings(tb)
		out = append(out, tb...)
	}
	if len(out) > 0 {
		return out
	}

	seen := map[string]bool{}
	var packages []string
	for _, artifact := range source.Document.Artifacts {
		if artifact.Kind != "android-termux-deb" {
			continue
		}
		name := artifact.Metadata["package"]
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		packages = append(packages, name)
	}
	sort.Strings(packages)
	return packages
}

func artifactCacheSuffix(artifact manifest.Artifact) string {
	if artifact.Kind == "android-test-binary" {
		return ".bin"
	}
	return ".tar.gz"
}

func LoadInstallStamp(prefix string) (InstallStamp, error) {
	if strings.TrimSpace(prefix) == "" {
		return InstallStamp{}, errors.New("prefix must not be empty")
	}
	payload, err := os.ReadFile(filepath.Join(prefix, ".zide-pm-install.json"))
	if err != nil {
		return InstallStamp{}, err
	}
	var stamp InstallStamp
	if err := json.Unmarshal(payload, &stamp); err != nil {
		return InstallStamp{}, err
	}
	if stamp.Package == "" {
		return InstallStamp{}, errors.New("install stamp missing package")
	}
	return stamp, nil
}

func AndroidPrefixArtifact(source Source) (PrefixArtifact, error) {
	var selected []manifest.Artifact
	for _, artifact := range source.Document.Artifacts {
		if artifact.Kind == "android-prefix-archive" {
			selected = append(selected, artifact)
		}
	}
	if len(selected) != 1 {
		return PrefixArtifact{}, fmt.Errorf("manifest must contain exactly one android-prefix-archive, found %d", len(selected))
	}
	artifact := selected[0]
	if artifact.Metadata["archive_root"] != "usr" {
		return PrefixArtifact{}, fmt.Errorf("unsupported archive_root %q", artifact.Metadata["archive_root"])
	}
	if artifact.Metadata["provider"] == "" {
		return PrefixArtifact{}, errors.New("android-prefix-archive missing provider metadata")
	}
	artifactURL, err := ResolveURL(source.Location, artifact.URL)
	if err != nil {
		return PrefixArtifact{}, err
	}
	return PrefixArtifact{Artifact: artifact, URL: artifactURL}, nil
}

func InstallDevBaseline(ctx context.Context, source Source, prefix string, cacheDir string) (InstallResult, error) {
	if filepath.Clean(prefix) == "." || strings.TrimSpace(prefix) == "" {
		return InstallResult{}, errors.New("prefix must not be empty")
	}
	artifact, err := AndroidPrefixArtifact(source)
	if err != nil {
		return InstallResult{}, err
	}
	archivePath, err := FetchArtifact(ctx, artifact.Artifact, artifact.URL, cacheDir)
	if err != nil {
		return InstallResult{}, err
	}
	stats, err := ExtractUSRToPrefix(archivePath, prefix)
	if err != nil {
		return InstallResult{}, err
	}
	if err := writeInstallStamp(prefix, source, artifact.Artifact, stats); err != nil {
		return InstallResult{}, err
	}
	return InstallResult{
		Package:       DevBaselinePackage,
		Prefix:        prefix,
		Manifest:      source.Location,
		Provider:      artifact.Artifact.Metadata["provider"],
		Version:       artifact.Artifact.Version,
		InstalledPath: archivePath,
		FileCount:     stats.files,
		DirCount:      stats.dirs,
		SymlinkCount:  stats.symlinks,
	}, nil
}

func FetchArtifact(ctx context.Context, artifact manifest.Artifact, artifactURL string, cacheDir string) (string, error) {
	if cacheDir == "" {
		cacheDir = DefaultCacheDir()
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}
	cacheName := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(
		artifact.Name + "-" + artifact.Version + "-" + artifact.SHA256[:12] + artifactCacheSuffix(artifact),
	)
	cachePath := filepath.Join(cacheDir, cacheName)
	if verifyPath(cachePath, artifact.Size, artifact.SHA256) {
		return cachePath, nil
	}

	tempPath := cachePath + ".tmp"
	if IsURL(artifactURL) {
		if err := downloadURL(ctx, artifactURL, tempPath); err != nil {
			_ = os.Remove(tempPath)
			return "", err
		}
	} else {
		if err := copyFile(artifactURL, tempPath); err != nil {
			_ = os.Remove(tempPath)
			return "", err
		}
	}
	if !verifyPath(tempPath, artifact.Size, artifact.SHA256) {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("artifact verification failed for %s", artifact.Name)
	}
	if err := os.Rename(tempPath, cachePath); err != nil {
		_ = os.Remove(tempPath)
		return "", err
	}
	return cachePath, nil
}

func ExtractUSRToPrefix(archivePath string, prefix string) (extractStats, error) {
	prefix = filepath.Clean(prefix)
	if err := os.MkdirAll(prefix, 0o755); err != nil {
		return extractStats{}, err
	}
	file, err := os.Open(archivePath)
	if err != nil {
		return extractStats{}, err
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return extractStats{}, err
	}
	defer gzipReader.Close()

	var stats extractStats
	reader := tar.NewReader(gzipReader)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return stats, err
		}
		relative, ok := strings.CutPrefix(filepath.ToSlash(filepath.Clean(header.Name)), "usr/")
		if !ok || relative == "" {
			continue
		}
		target, err := safeJoin(prefix, relative)
		if err != nil {
			return stats, err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)&0o777); err != nil {
				return stats, err
			}
			stats.dirs++
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return stats, err
			}
			if err := writeRegularFile(target, reader, os.FileMode(header.Mode)&0o777); err != nil {
				return stats, err
			}
			stats.files++
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return stats, err
			}
			_ = os.Remove(target)
			if err := os.Symlink(header.Linkname, target); err != nil {
				return stats, err
			}
			stats.symlinks++
		default:
			continue
		}
	}
	return stats, nil
}

func ResolveURL(base string, value string) (string, error) {
	if IsURL(value) {
		return value, nil
	}
	if IsURL(base) {
		parsedBase, err := url.Parse(base)
		if err != nil {
			return "", err
		}
		parsedValue, err := url.Parse(value)
		if err != nil {
			return "", err
		}
		return parsedBase.ResolveReference(parsedValue).String(), nil
	}
	return filepath.Join(filepath.Dir(base), value), nil
}

func IsURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func DefaultCacheDir() string {
	if value := os.Getenv("ZIDE_PM_CACHE"); value != "" {
		return value
	}
	if value := os.Getenv("XDG_CACHE_HOME"); value != "" {
		return filepath.Join(value, "zide-pm")
	}
	if runtime.GOOS == "android" {
		if value := os.Getenv("TMPDIR"); value != "" {
			return filepath.Join(value, "zide-pm-cache")
		}
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".cache", "zide-pm")
	}
	return filepath.Join(os.TempDir(), "zide-pm-cache")
}

type extractStats struct {
	files    int
	dirs     int
	symlinks int
}

func readURL(ctx context.Context, target string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	addAuthHeaders(request)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", target, response.Status)
	}
	return io.ReadAll(response.Body)
}

func downloadURL(ctx context.Context, target string, outputPath string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	addAuthHeaders(request)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", target, response.Status)
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, response.Body)
	return err
}

func addAuthHeaders(request *http.Request) {
	request.Header.Set("User-Agent", "zide-pm")
	if request.URL.Host != "github.com" {
		return
	}
	for _, name := range []string{"ZIDE_PM_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"} {
		if token := os.Getenv(name); token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
			return
		}
	}
	if token := ghAuthToken(); token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
}

func ghAuthToken() string {
	gh, err := exec.LookPath("gh")
	if err != nil {
		return ""
	}
	output, err := exec.Command(gh, "auth", "token").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func verifyPath(path string, wantSize int64, wantSHA256 string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if wantSize >= 0 && info.Size() != wantSize {
		return false
	}
	got, err := sha256Path(path)
	return err == nil && got == wantSHA256
}

func sha256Path(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func copyFile(source string, target string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(target)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func writeRegularFile(path string, reader io.Reader, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(file, reader); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func safeJoin(root string, relative string) (string, error) {
	clean := filepath.Clean(relative)
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe archive path %q", relative)
	}
	target := filepath.Join(root, clean)
	resolvedRelative, err := filepath.Rel(root, target)
	if err != nil {
		return "", err
	}
	if resolvedRelative == ".." || strings.HasPrefix(resolvedRelative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("archive path escapes prefix: %q", relative)
	}
	return target, nil
}

func writeInstallStamp(prefix string, source Source, artifact manifest.Artifact, stats extractStats) error {
	stamp := InstallStamp{
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
		Package:     DevBaselinePackage,
		Manifest:    source.Location,
		Artifact:    artifact.Name,
		Version:     artifact.Version,
		Provider:    artifact.Metadata["provider"],
		Files:       stats.files,
		Dirs:        stats.dirs,
		Symlinks:    stats.symlinks,
	}
	payload, err := json.MarshalIndent(stamp, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(prefix, ".zide-pm-install.json"), append(payload, '\n'), 0o644)
}
