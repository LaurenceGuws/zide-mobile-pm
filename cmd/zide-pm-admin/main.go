// Command zide-pm-admin manages mobile artifact manifests for Zide consumers.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaurenceGuws/zide-mobile-pm/internal/androidprefix"
	"github.com/LaurenceGuws/zide-mobile-pm/internal/androidrepo"
	"github.com/LaurenceGuws/zide-mobile-pm/internal/manifest"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "help", "-h", "--help":
		printHelp()
	case "version":
		fmt.Println(version)
	case "contract":
		if err := printContract(os.Args[2:]); err != nil {
			die(err)
		}
	case "android-dev-manifest":
		if err := androidDevManifest(os.Args[2:]); err != nil {
			die(err)
		}
	case "android-prefix-archive":
		if err := androidPrefixArchive(os.Args[2:]); err != nil {
			die(err)
		}
	case "android-product-candidate-probe":
		if err := androidProductCandidateProbe(os.Args[2:]); err != nil {
			die(err)
		}
	case "android-dev-snapshot-release":
		if err := androidDevRelease(os.Args[2:]); err != nil {
			die(err)
		}
	case "validate":
		if err := validate(os.Args[2:]); err != nil {
			die(err)
		}
	default:
		die(fmt.Errorf("unknown command %q", os.Args[1]))
	}
}

func printHelp() {
	fmt.Println(`zide-pm-admin manages mobile package/artifact manifests for Zide consumers.

Usage:
  zide-pm-admin <command> [options]

Commands:
  android-dev-manifest
             Fetch/cache Android package metadata and emit a pinned dev manifest.
  android-prefix-archive
             Build a dev Android prefix archive from a pinned dev manifest.
  android-product-candidate-probe
             MP-A6 evidence: run prefix archive with hardcoded-policy=fail (temp outputs).
  android-dev-snapshot-release
             Publish a fast Android dev snapshot prerelease with generated artifacts.
  contract   Print the current artifact contract skeleton as JSON.
  validate   Validate a manifest JSON file against the current schema floor.
  version    Print the tool version.
  help       Show this help.

This tool is backend/admin tooling. It does not run inside the Zide mobile apps
and does not replace platform-native runtime integration.`)
}

func androidDevManifest(args []string) error {
	fs := flag.NewFlagSet("android-dev-manifest", flag.ExitOnError)
	channel := fs.String("channel", "dev", "artifact channel name")
	cacheDir := fs.String("cache-dir", ".cache/android/termux-main/aarch64", "package index cache directory")
	out := fs.String("out", "dist/android-dev.manifest.json", "output manifest path, or - for stdout")
	indexURL := fs.String("index-url", androidrepo.DefaultIndexURL, "Android package index URL")
	baseURL := fs.String("base-url", androidrepo.DefaultBaseURL, "base URL for package filenames")
	roots := fs.String("packages", "bash,neovim,git,ripgrep,htop,gotop", "comma-separated root packages for the dev channel")
	refresh := fs.Bool("refresh", false, "refresh the cached package index")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	indexBytes, indexPath, err := loadOrFetchIndex(ctx, *cacheDir, *indexURL, *refresh)
	if err != nil {
		return err
	}
	index, err := androidrepo.ParseIndex(bytes.NewReader(indexBytes))
	if err != nil {
		return err
	}
	rootPackages := splitCSV(*roots)
	packages, err := androidrepo.ResolveClosure(index, rootPackages)
	if err != nil {
		return err
	}

	doc, err := newAndroidDevManifest(*channel, *indexURL, *baseURL, indexBytes, packages, rootPackages)
	if err != nil {
		return err
	}
	if err := writeManifest(*out, doc); err != nil {
		return err
	}
	if *out != "-" {
		fmt.Printf("wrote %s packages=%d index_cache=%s\n", *out, len(packages), indexPath)
	}
	return nil
}

func androidDevRelease(args []string) error {
	fs := flag.NewFlagSet("android-dev-snapshot-release", flag.ExitOnError)
	tag := fs.String("tag", defaultAndroidDevTag(), "release tag to create")
	title := fs.String("title", "", "release title, defaults to tag")
	refresh := fs.Bool("refresh", false, "refresh provider package index before release")
	dryRun := fs.Bool("dry-run", false, "generate assets and print release command without publishing")
	repo := fs.String("repo", "LaurenceGuws/zide-mobile-pm", "GitHub repository for release publishing")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*tag) == "" {
		return fmt.Errorf("tag must not be empty")
	}
	releaseTitle := *title
	if releaseTitle == "" {
		releaseTitle = *tag
	}

	devManifestPath := "dist/android-dev.manifest.json"
	prefixArchivePath := "dist/zide-android-dev-prefix.tar.gz"
	prefixManifestPath := "dist/android-dev-prefix.manifest.json"
	auditPath := "dist/zide-android-dev-prefix.audit.json"
	releaseManifestPath := "dist/android-dev-prefix.release.manifest.json"

	manifestArgs := []string{
		"-out", devManifestPath,
	}
	if *refresh {
		manifestArgs = append(manifestArgs, "-refresh")
	}
	if err := androidDevManifest(manifestArgs); err != nil {
		return err
	}
	if err := buildAndroidZidePM("dist/zide-pm-android-arm64"); err != nil {
		return err
	}
	if err := androidPrefixArchive([]string{
		"-manifest", devManifestPath,
		"-out", prefixArchivePath,
		"-out-manifest", prefixManifestPath,
		"-audit-out", auditPath,
		"-zide-pm-bin", "dist/zide-pm-android-arm64",
		"-hardcoded-policy", "audit",
	}); err != nil {
		return err
	}
	if err := writeAndroidDevReleaseManifest(prefixManifestPath, releaseManifestPath, filepath.Base(prefixArchivePath)); err != nil {
		return err
	}
	if err := manifestPathValid(releaseManifestPath); err != nil {
		return err
	}

	body := androidDevReleaseBody(*tag, releaseManifestPath, auditPath)
	bodyPath := filepath.Join("dist", *tag+".release-notes.md")
	if err := os.WriteFile(bodyPath, []byte(body), 0o644); err != nil {
		return err
	}

	assetArgs := []string{
		releaseManifestPath,
		devManifestPath,
		prefixArchivePath,
		catalogSmokeReleaseAssetPath(),
		auditPath,
	}
	fmt.Printf("prepared android dev snapshot release tag=%s manifest=%s archive=%s\n", *tag, releaseManifestPath, prefixArchivePath)
	if *dryRun {
		fmt.Printf("dry-run: git tag %s\n", *tag)
		fmt.Printf("dry-run: git push origin %s\n", *tag)
		fmt.Printf("dry-run: gh release create %s %s --repo %s --title %q --notes-file %s --prerelease\n",
			*tag,
			strings.Join(assetArgs, " "),
			*repo,
			releaseTitle,
			bodyPath,
		)
		return nil
	}

	if err := ensureCleanWorktree(); err != nil {
		return err
	}
	if err := runCommand("git", "tag", "--annotate", *tag, "--message", releaseTitle); err != nil {
		return err
	}
	if err := runCommand("git", "push", "origin", *tag); err != nil {
		return err
	}
	argsForGH := []string{
		"release", "create", *tag,
		"--repo", *repo,
		"--title", releaseTitle,
		"--notes-file", bodyPath,
		"--prerelease",
	}
	argsForGH = append(argsForGH, assetArgs...)
	if err := runCommand("gh", argsForGH...); err != nil {
		return err
	}
	fmt.Printf("published https://github.com/%s/releases/tag/%s\n", *repo, *tag)
	fmt.Printf("consumer manifest URL: https://github.com/%s/releases/download/%s/%s\n", *repo, *tag, filepath.Base(releaseManifestPath))
	return nil
}

func androidPrefixArchive(args []string) error {
	fs := flag.NewFlagSet("android-prefix-archive", flag.ExitOnError)
	manifestPath := fs.String("manifest", "dist/android-dev.manifest.json", "input MP-A1 Android dev manifest")
	cacheDir := fs.String("cache-dir", ".cache/android/packages", "downloaded package cache directory")
	workDir := fs.String("work-dir", ".cache/android/prefix-work", "temporary extraction directory")
	out := fs.String("out", "dist/zide-android-dev-prefix.tar.gz", "output prefix archive path")
	outManifest := fs.String("out-manifest", "dist/android-dev-prefix.manifest.json", "output archive manifest path")
	auditOut := fs.String("audit-out", "dist/zide-android-dev-prefix.audit.json", "output archive audit path")
	hardcodedPolicy := fs.String("hardcoded-policy", "fail", "hardcoded com.termux policy: audit or fail")
	zidePMBin := fs.String("zide-pm-bin", "", "optional Android zide-pm binary to include as usr/bin/zide-pm")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *hardcodedPolicy != "audit" && *hardcodedPolicy != "fail" {
		return fmt.Errorf("unsupported hardcoded-policy %q", *hardcodedPolicy)
	}

	sourceBytes, err := os.ReadFile(*manifestPath)
	if err != nil {
		return err
	}
	sourceDoc, err := manifest.Load(*manifestPath)
	if err != nil {
		return err
	}
	if err := sourceDoc.Validate(); err != nil {
		return err
	}
	debArtifacts := androidDebArtifacts(sourceDoc)
	if len(debArtifacts) == 0 {
		return fmt.Errorf("%s has no android-termux-deb artifacts", *manifestPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	stagingRoot := filepath.Join(*workDir, "staging")
	if err := os.RemoveAll(stagingRoot); err != nil {
		return err
	}
	if err := os.MkdirAll(stagingRoot, 0o755); err != nil {
		return err
	}

	audit := prefixAudit{
		SourceManifest:  *manifestPath,
		PackageCount:    len(debArtifacts),
		HardcodedPolicy: *hardcodedPolicy,
	}
	sourceHash := fmt.Sprintf("%x", sha256.Sum256(sourceBytes))
	for _, artifact := range debArtifacts {
		debPath, err := downloadArtifact(ctx, artifact, *cacheDir)
		if err != nil {
			return err
		}
		stats, err := androidprefix.ExtractDebUSR(debPath, stagingRoot)
		if err != nil {
			return fmt.Errorf("%s: extract: %w", artifact.Name, err)
		}
		audit.ExtractedEntries += stats.Entries
		audit.ExtractedFiles += stats.Files
		audit.ExtractedDirs += stats.Dirs
		audit.ExtractedSymlinks += stats.Symlinks
		audit.ExtractedHardlinks += stats.Hardlinks
		audit.SkippedEntries += stats.Skipped
		audit.TextRewrites += stats.TextRewrites
		audit.BinaryRewrites += stats.BinaryRewrites
		for _, hit := range stats.HardcodedTermuxHits {
			audit.HardcodedTermuxHits = append(audit.HardcodedTermuxHits, artifact.Name+":"+hit)
		}
	}
	if *zidePMBin != "" {
		if err := installZidePMBinary(stagingRoot, *zidePMBin); err != nil {
			return err
		}
	}
	removedTermuxBinaries, err := androidprefix.PruneTermuxPrefixedBinaries(stagingRoot)
	if err != nil {
		return err
	}
	audit.RemovedTermuxBinaries = removedTermuxBinaries
	if err := writeBundledZidePMInstallStamp(stagingRoot, *manifestPath, sourceHash, audit); err != nil {
		return err
	}
	if len(audit.HardcodedTermuxHits) > 0 && *hardcodedPolicy == "fail" {
		if err := writeJSON(*auditOut, audit); err != nil {
			return err
		}
		return fmt.Errorf("hardcoded com.termux hits remain: %d (audit written to %s)", len(audit.HardcodedTermuxHits), *auditOut)
	}

	archiveStats, err := androidprefix.WriteTarGz(stagingRoot, *out)
	if err != nil {
		return err
	}
	audit.ArchivePath = *out
	audit.ArchiveSHA256 = archiveStats.SHA256
	audit.ArchiveSize = archiveStats.Size
	audit.ArchiveFiles = archiveStats.Files
	audit.ArchiveDirs = archiveStats.Dirs
	audit.ArchiveSymlinks = archiveStats.Symlinks
	if err := writeJSON(*auditOut, audit); err != nil {
		return err
	}

	archiveDoc := newAndroidPrefixManifest(sourceDoc.Channel, *out, archiveStats, sourceHash, audit)
	if err := writeManifest(*outManifest, archiveDoc); err != nil {
		return err
	}
	fmt.Printf(
		"wrote %s files=%d symlinks=%d packages=%d text_rewrites=%d binary_rewrites=%d termux_prefixed_binaries_removed=%d hardcoded_termux_hits=%d manifest=%s audit=%s\n",
		*out,
		archiveStats.Files,
		archiveStats.Symlinks,
		len(debArtifacts),
		audit.TextRewrites,
		audit.BinaryRewrites,
		audit.RemovedTermuxBinaries,
		len(audit.HardcodedTermuxHits),
		*outManifest,
		*auditOut,
	)
	return nil
}

func printContract(args []string) error {
	fs := flag.NewFlagSet("contract", flag.ExitOnError)
	platform := fs.String("platform", "android", "target platform: android or ios")
	channel := fs.String("channel", "dev", "artifact channel name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	doc, err := manifest.NewSkeleton(*platform, *channel)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(doc)
}

func validate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("validate expects exactly one manifest path")
	}

	doc, err := manifest.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	if err := doc.Validate(); err != nil {
		return err
	}
	fmt.Printf("ok platform=%s channel=%s artifacts=%d\n", doc.Platform, doc.Channel, len(doc.Artifacts))
	return nil
}

func manifestPathValid(path string) error {
	doc, err := manifest.Load(path)
	if err != nil {
		return err
	}
	return doc.Validate()
}

func androidDevReleaseBody(tag string, manifestPath string, auditPath string) string {
	return fmt.Sprintf(`# %s

Automated Android development artifact release.

This release is for Zide Android terminal bringup and device testing. It is not
a formal product userland release.

Assets:

- %s
- android-dev.manifest.json
- zide-android-dev-prefix.tar.gz
- zide-android-catalog-smoke.sh
- %s

Policy:

- provider: termux-main
- provider role: android-dev-bootstrap
- hardcoded prefix policy: audit
- product releases must pass the stricter hardcoded-prefix policy before being
  treated as product-clean
`, tag, filepath.Base(manifestPath), filepath.Base(auditPath))
}

func defaultAndroidDevTag() string {
	return "android-dev-" + time.Now().UTC().Format("2006.01.02.150405")
}

func ensureCleanWorktree() error {
	output, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(output)) != "" {
		return fmt.Errorf("worktree must be clean before publishing a release")
	}
	return nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

type prefixAudit struct {
	SourceManifest        string   `json:"source_manifest"`
	PackageCount          int      `json:"package_count"`
	HardcodedPolicy       string   `json:"hardcoded_policy"`
	ExtractedEntries      int      `json:"extracted_entries"`
	ExtractedFiles        int      `json:"extracted_files"`
	ExtractedDirs         int      `json:"extracted_dirs"`
	ExtractedSymlinks     int      `json:"extracted_symlinks"`
	ExtractedHardlinks    int      `json:"extracted_hardlinks"`
	SkippedEntries        int      `json:"skipped_entries"`
	TextRewrites          int      `json:"text_rewrites"`
	BinaryRewrites        int      `json:"binary_rewrites"`
	RemovedTermuxBinaries int      `json:"removed_termux_prefixed_binaries"`
	HardcodedTermuxHits   []string `json:"hardcoded_termux_hits,omitempty"`
	ArchivePath           string   `json:"archive_path"`
	ArchiveSHA256         string   `json:"archive_sha256"`
	ArchiveSize           int64    `json:"archive_size"`
	ArchiveFiles          int      `json:"archive_files"`
	ArchiveDirs           int      `json:"archive_dirs"`
	ArchiveSymlinks       int      `json:"archive_symlinks"`
}

func androidDebArtifacts(doc manifest.Document) []manifest.Artifact {
	var artifacts []manifest.Artifact
	for _, artifact := range doc.Artifacts {
		if artifact.Kind == "android-termux-deb" {
			artifacts = append(artifacts, artifact)
		}
	}
	return artifacts
}

func downloadArtifact(ctx context.Context, artifact manifest.Artifact, cacheDir string) (string, error) {
	filename := artifact.Metadata["filename"]
	if filename == "" {
		filename = strings.TrimPrefix(artifact.Name, "termux-main/") + ".deb"
	}
	cachePath := filepath.Join(cacheDir, filepath.FromSlash(filename))
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return "", err
	}
	if ok, err := verifyFile(cachePath, artifact.SHA256, artifact.Size); err != nil {
		return "", err
	} else if ok {
		return cachePath, nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, artifact.URL, nil)
	if err != nil {
		return "", err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: download: %s", artifact.Name, response.Status)
	}

	tempPath := cachePath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(file, hash), response.Body)
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(tempPath)
		return "", copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tempPath)
		return "", closeErr
	}
	if artifact.Size >= 0 && written != artifact.Size {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("%s: size mismatch: got %d want %d", artifact.Name, written, artifact.Size)
	}
	gotHash := fmt.Sprintf("%x", hash.Sum(nil))
	if gotHash != artifact.SHA256 {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("%s: sha256 mismatch: got %s want %s", artifact.Name, gotHash, artifact.SHA256)
	}
	if err := os.Rename(tempPath, cachePath); err != nil {
		_ = os.Remove(tempPath)
		return "", err
	}
	return cachePath, nil
}

func verifyFile(path string, wantSHA256 string, wantSize int64) (bool, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return false, err
	}
	if wantSize >= 0 && size != wantSize {
		return false, nil
	}
	return fmt.Sprintf("%x", hash.Sum(nil)) == wantSHA256, nil
}

func buildAndroidZidePM(outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-trimpath", "-o", outputPath, "./cmd/zide-pm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), "GOOS=android", "GOARCH=arm64", "CGO_ENABLED=0")
	if err := cmd.Run(); err != nil {
		return err
	}
	return os.Chmod(outputPath, 0o755)
}

func installZidePMBinary(stagingRoot string, sourcePath string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("zide-pm binary path is a directory: %s", sourcePath)
	}
	target := filepath.Join(stagingRoot, "usr", "bin", "zide-pm")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	destination, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(destination, source); err != nil {
		_ = destination.Close()
		return err
	}
	return destination.Close()
}

func writeBundledZidePMInstallStamp(stagingRoot string, manifestPath string, sourceManifestSHA256 string, audit prefixAudit) error {
	stamp := map[string]any{
		"installed_at":            "archive-build",
		"package":                 "dev-baseline",
		"manifest":                manifestPath,
		"artifact":                "zide-android-dev-prefix",
		"version":                 "source-sha256-" + sourceManifestSHA256[:12],
		"provider":                "termux-main",
		"hardcoded_termux_policy": audit.HardcodedPolicy,
		"files":                   audit.ExtractedFiles,
		"dirs":                    audit.ExtractedDirs,
		"symlinks":                audit.ExtractedSymlinks,
	}
	payload, err := json.MarshalIndent(stamp, "", "  ")
	if err != nil {
		return err
	}
	target := filepath.Join(stagingRoot, "usr", ".zide-pm-install.json")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, append(payload, '\n'), 0o644)
}

func newAndroidPrefixManifest(
	channel string,
	archivePath string,
	archiveStats androidprefix.ArchiveStats,
	sourceManifestSHA256 string,
	audit prefixAudit,
) manifest.Document {
	return manifest.Document{
		SchemaVersion: manifest.SchemaVersion,
		Project:       "zide-mobile-pm",
		Platform:      "android",
		Channel:       channel,
		Artifacts: []manifest.Artifact{{
			Name:    "zide-android-dev-prefix",
			Kind:    "android-prefix-archive",
			Version: "sha256-" + archiveStats.SHA256[:12],
			URL:     filepath.ToSlash(archivePath),
			SHA256:  archiveStats.SHA256,
			Size:    archiveStats.Size,
			Metadata: map[string]string{
				"package_name":                     "uk.laurencegouws.zide",
				"prefix":                           "/data/data/uk.laurencegouws.zide/files/usr",
				"archive_root":                     "usr",
				"target_sdk":                       "28",
				"provider":                         "termux-main",
				"provider_role":                    "android-dev-bootstrap",
				"provider_platform":                "android",
				"provider_architecture":            "aarch64",
				"source_manifest_sha256":           sourceManifestSHA256,
				"source_package_count":             fmt.Sprintf("%d", audit.PackageCount),
				"hardcoded_termux_hits":            fmt.Sprintf("%d", len(audit.HardcodedTermuxHits)),
				"hardcoded_termux_policy":          audit.HardcodedPolicy,
				"text_rewrites":                    fmt.Sprintf("%d", audit.TextRewrites),
				"binary_rewrites":                  fmt.Sprintf("%d", audit.BinaryRewrites),
				"runtime_support_files":            "/data/user/0/uk.laurencegouws.zide/t/b,/data/user/0/uk.laurencegouws.zide/t/p,/data/user/0/uk.laurencegouws.zide/t/h,/data/user/0/uk.laurencegouws.zide/t/hs",
				"runtime_support_links":            "/data/user/0/uk.laurencegouws.zide/t/b=>/data/user/0/uk.laurencegouws.zide/files/usr/etc/bash.bashrc,/data/user/0/uk.laurencegouws.zide/t/p=>/data/user/0/uk.laurencegouws.zide/files/usr/etc/profile,/data/user/0/uk.laurencegouws.zide/t/h=>/data/user/0/uk.laurencegouws.zide/files/usr/etc/hosts,/data/user/0/uk.laurencegouws.zide/t/hs=>/data/user/0/uk.laurencegouws.zide/files/usr/var/htop/stat",
				"removed_termux_prefixed_binaries": fmt.Sprintf("%d", audit.RemovedTermuxBinaries),
				"extracted_regular_files":          fmt.Sprintf("%d", audit.ExtractedFiles),
				"extracted_symlinks":               fmt.Sprintf("%d", audit.ExtractedSymlinks),
				"archive_regular_files":            fmt.Sprintf("%d", archiveStats.Files),
				"archive_symlinks":                 fmt.Sprintf("%d", archiveStats.Symlinks),
				"zide_pm_cli":                      "included",
			},
			Limitations: []string{
				"Development prefix archive for Android terminal bringup.",
				"Generated from pinned upstream package payloads; product channels must review the audit before release.",
			},
		}},
		Notes: []string{
			"Archive root is usr/ and is intended to be staged under the Android app files directory.",
			"Zide should consume this archive by manifest contract instead of parsing package internals.",
		},
	}
}

func loadOrFetchIndex(ctx context.Context, cacheDir string, indexURL string, refresh bool) ([]byte, string, error) {
	indexPath := filepath.Join(cacheDir, "Packages")
	hashPath := indexPath + ".sha256"
	if !refresh {
		if bytes, err := os.ReadFile(indexPath); err == nil {
			hash := androidrepo.HashBytes(bytes)
			if expectedBytes, err := os.ReadFile(hashPath); err == nil {
				expected := strings.TrimSpace(string(expectedBytes))
				if expected != "" && expected != hash {
					return nil, "", fmt.Errorf("cached package index checksum mismatch: %s", indexPath)
				}
			}
			return bytes, indexPath, nil
		}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, indexURL, nil)
	if err != nil {
		return nil, "", err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("fetch package index: %s", response.Status)
	}
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, "", err
	}
	if err := os.WriteFile(indexPath, bytes, 0o644); err != nil {
		return nil, "", err
	}
	hash := androidrepo.HashBytes(bytes)
	if err := os.WriteFile(hashPath, []byte(hash+"\n"), 0o644); err != nil {
		return nil, "", err
	}
	return bytes, indexPath, nil
}

func newAndroidDevManifest(
	channel string,
	indexURL string,
	baseURL string,
	indexBytes []byte,
	packages []androidrepo.Package,
	roots []string,
) (manifest.Document, error) {
	doc := manifest.Document{
		SchemaVersion: manifest.SchemaVersion,
		Project:       "zide-mobile-pm",
		Platform:      "android",
		Channel:       channel,
		Artifacts: []manifest.Artifact{{
			Name:    "termux-main-aarch64-packages-index",
			Kind:    "android-termux-package-index",
			Version: "sha256-" + androidrepo.HashBytes(indexBytes)[:12],
			URL:     indexURL,
			SHA256:  androidrepo.HashBytes(indexBytes),
			Size:    int64(len(indexBytes)),
			Metadata: map[string]string{
				"architecture":          "aarch64",
				"provider":              "termux-main",
				"provider_role":         "android-dev-bootstrap",
				"provider_platform":     "android",
				"provider_architecture": "aarch64",
				"provider_repository":   "termux-main",
			},
		}},
		Notes: []string{
			"Development channel manifest for Zide Android terminal userland work.",
			"This pins package metadata and payload checksums; it is not a final product package-manager contract.",
			"Root packages: " + strings.Join(roots, ","),
		},
	}

	for _, pkg := range packages {
		packageURL, err := androidrepo.AbsolutePackageURL(baseURL, pkg.Filename)
		if err != nil {
			return manifest.Document{}, fmt.Errorf("%s: build package URL: %w", pkg.Name, err)
		}
		metadata := map[string]string{
			"package":               pkg.Name,
			"architecture":          pkg.Architecture,
			"filename":              pkg.Filename,
			"provider":              "termux-main",
			"provider_role":         "android-dev-bootstrap",
			"provider_platform":     "android",
			"provider_architecture": "aarch64",
			"provider_repository":   "termux-main",
		}
		if pkg.Depends != "" {
			metadata["depends"] = pkg.Depends
		}
		if pkg.PreDepends != "" {
			metadata["pre_depends"] = pkg.PreDepends
		}
		doc.Artifacts = append(doc.Artifacts, manifest.Artifact{
			Name:     "termux-main/" + pkg.Name,
			Kind:     "android-termux-deb",
			Version:  pkg.Version,
			URL:      packageURL,
			SHA256:   pkg.SHA256,
			Size:     pkg.Size,
			Metadata: metadata,
			Limitations: []string{
				"Payload is pinned upstream package data. Product archives must still prove uk.laurencegouws.zide prefix correctness.",
			},
		})
	}

	return doc, nil
}

func writeManifest(path string, doc manifest.Document) error {
	return writeJSON(path, doc)
}

func writeJSON(path string, value any) error {
	writer := io.Writer(os.Stdout)
	var file *os.File
	if path != "-" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		created, err := os.Create(path)
		if err != nil {
			return err
		}
		defer created.Close()
		file = created
		writer = file
	}
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return err
	}
	if file != nil {
		return file.Sync()
	}
	return nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
