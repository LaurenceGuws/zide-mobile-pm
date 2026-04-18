// Package androidprefix materializes Android app-prefix archives from pinned
// package artifacts.
package androidprefix

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

const (
	TermuxUSRPrefix = "data/data/com.termux/files/usr/"
	ZideUSRPrefix   = "usr/"
	AppPackageName  = "uk.laurencegouws.zide"
	AppUSRPath      = "/data/data/" + AppPackageName + "/files/usr"
	RuntimeAliasDir = "/data/user/0/" + AppPackageName + "/t"
)

type ExtractStats struct {
	Entries             int
	Files               int
	Dirs                int
	Symlinks            int
	Hardlinks           int
	Skipped             int
	TextRewrites        int
	BinaryRewrites      int
	HardcodedTermuxHits []string
}

func ExtractDebUSR(debPath string, stagingRoot string) (ExtractStats, error) {
	dataName, dataBytes, err := debDataMember(debPath)
	if err != nil {
		return ExtractStats{}, err
	}
	reader, closeReader, err := dataTarReader(dataName, bytes.NewReader(dataBytes))
	if err != nil {
		return ExtractStats{}, err
	}
	defer closeReader()
	return extractUSR(reader, stagingRoot)
}

func debDataMember(path string) (string, []byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	header := make([]byte, 8)
	if _, err := io.ReadFull(file, header); err != nil {
		return "", nil, err
	}
	if string(header) != "!<arch>\n" {
		return "", nil, fmt.Errorf("%s: not an ar archive", path)
	}

	for {
		memberHeader := make([]byte, 60)
		_, err := io.ReadFull(file, memberHeader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", nil, err
		}
		if string(memberHeader[58:60]) != "`\n" {
			return "", nil, fmt.Errorf("%s: invalid ar member header", path)
		}
		name := strings.TrimSpace(string(memberHeader[0:16]))
		name = strings.TrimSuffix(name, "/")
		sizeText := strings.TrimSpace(string(memberHeader[48:58]))
		var size int64
		if _, err := fmt.Sscanf(sizeText, "%d", &size); err != nil {
			return "", nil, fmt.Errorf("%s: invalid ar member size %q", path, sizeText)
		}
		if size < 0 {
			return "", nil, fmt.Errorf("%s: invalid negative ar member size", path)
		}
		data := make([]byte, size)
		if _, err := io.ReadFull(file, data); err != nil {
			return "", nil, err
		}
		if size%2 != 0 {
			if _, err := file.Seek(1, io.SeekCurrent); err != nil {
				return "", nil, err
			}
		}
		if strings.HasPrefix(name, "data.tar") {
			return name, data, nil
		}
	}

	return "", nil, fmt.Errorf("%s: missing data.tar member", path)
}

func dataTarReader(name string, reader io.Reader) (io.Reader, func() error, error) {
	if strings.HasSuffix(name, ".gz") {
		gz, err := gzip.NewReader(reader)
		if err != nil {
			return nil, nil, err
		}
		return gz, gz.Close, nil
	}
	if strings.HasSuffix(name, ".xz") {
		xzr, err := xz.NewReader(reader)
		if err != nil {
			return nil, nil, err
		}
		return xzr, func() error { return nil }, nil
	}
	if strings.HasSuffix(name, ".zst") || strings.HasSuffix(name, ".zstd") {
		zr, err := zstd.NewReader(reader)
		if err != nil {
			return nil, nil, err
		}
		return zr, func() error { zr.Close(); return nil }, nil
	}
	if strings.HasSuffix(name, ".tar") {
		return reader, func() error { return nil }, nil
	}
	return nil, nil, fmt.Errorf("unsupported data archive compression: %s", name)
}

func extractUSR(reader io.Reader, stagingRoot string) (ExtractStats, error) {
	var stats ExtractStats
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return stats, err
		}
		stats.Entries++

		relative, ok := usrRelativePath(header.Name)
		if !ok {
			stats.Skipped++
			continue
		}
		targetPath, err := safeJoin(stagingRoot, relative)
		if err != nil {
			return stats, err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)&0o777); err != nil {
				return stats, err
			}
			stats.Dirs++
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return stats, err
			}
			bytes, err := io.ReadAll(tarReader)
			if err != nil {
				return stats, err
			}
			rewritten, textChanged, binaryRewrites, binaryHit := rewriteTermuxBytes(relative, bytes)
			if binaryRewrites > 0 {
				stats.BinaryRewrites += binaryRewrites
			}
			if textChanged {
				stats.TextRewrites++
			}
			if binaryHit {
				stats.HardcodedTermuxHits = append(stats.HardcodedTermuxHits, relative)
			}
			if err := os.WriteFile(targetPath, rewritten, os.FileMode(header.Mode)&0o777); err != nil {
				return stats, err
			}
			stats.Files++
		case tar.TypeSymlink:
			linkname, hit := rewriteTermuxLink(relative, header.Linkname)
			if hit {
				stats.TextRewrites++
			}
			if strings.Contains(linkname, "com.termux") {
				stats.HardcodedTermuxHits = append(stats.HardcodedTermuxHits, relative)
			}
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return stats, err
			}
			_ = os.Remove(targetPath)
			if err := os.Symlink(linkname, targetPath); err != nil {
				return stats, err
			}
			stats.Symlinks++
		case tar.TypeLink:
			linkRelative, ok := usrRelativePath(header.Linkname)
			if !ok {
				stats.HardcodedTermuxHits = append(stats.HardcodedTermuxHits, relative)
				stats.Skipped++
				continue
			}
			sourcePath, err := safeJoin(stagingRoot, linkRelative)
			if err != nil {
				return stats, err
			}
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return stats, err
			}
			_ = os.Remove(targetPath)
			if err := os.Link(sourcePath, targetPath); err != nil {
				return stats, err
			}
			stats.Hardlinks++
		default:
			stats.Skipped++
		}
	}
	return stats, nil
}

func usrRelativePath(raw string) (string, bool) {
	name := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(raw, "./")))
	if name == "." {
		return "", false
	}
	if name == strings.TrimSuffix(TermuxUSRPrefix, "/") {
		return strings.TrimSuffix(ZideUSRPrefix, "/"), true
	}
	if strings.HasPrefix(name, TermuxUSRPrefix) {
		rest := strings.TrimPrefix(name, TermuxUSRPrefix)
		if rest == "" {
			return strings.TrimSuffix(ZideUSRPrefix, "/"), true
		}
		return ZideUSRPrefix + rest, true
	}
	return "", false
}

func rewriteTermuxBytes(relative string, payload []byte) ([]byte, bool, int, bool) {
	oldPrefix := []byte("/data/data/com.termux/files/usr")
	if !bytes.Contains(payload, oldPrefix) {
		return payload, false, 0, false
	}
	if !looksText(payload) {
		rewritten, rewrites := rewriteKnownBinaryTermuxPaths(payload)
		return rewritten, false, rewrites, bytes.Contains(rewritten, oldPrefix)
	}
	rewritten := bytes.ReplaceAll(payload, oldPrefix, []byte(AppUSRPath))
	return rewritten, true, 0, false
}

func rewriteKnownBinaryTermuxPaths(payload []byte) ([]byte, int) {
	rewritten := append([]byte(nil), payload...)
	rewrites := 0
	// Longest old strings first so paths that share prefixes (e.g. usr/bin vs
	// usr/bin/sh) rewrite without corrupting longer C strings.
	for _, replacement := range []struct {
		old         string
		new         string
		cStringOnly bool
	}{
		{
			old: "/data/data/com.termux/files/usr/bin/sh",
			new: "/data/data/uk.laurencegouws.zide/u/bsh",
		},
		{
			old: "/data/data/com.termux/files/usr/etc/bash.bashrc",
			new: RuntimeAliasDir + "/b",
		},
		{
			old: "/data/data/com.termux/files/usr/etc/profile",
			new: RuntimeAliasDir + "/p",
		},
		{
			old: "/data/data/com.termux/files/usr/etc/hosts",
			new: RuntimeAliasDir + "/h",
		},
		{
			old: "/data/data/com.termux/files/usr/var/htop/stat",
			new: RuntimeAliasDir + "/hs",
		},
		{
			old: "RfPATH=/data/data/com.termux/files/usr/bin",
			new: "RfPATH=/data/data/uk.laurencegouws.zide/b",
		},
		{
			old:         "/data/data/com.termux/files/usr/lib",
			new:         "/data/data/uk.laurencegouws.zide/ul",
			cStringOnly: true,
		},
		{
			old:         "/data/data/com.termux/files/usr/bin",
			new:         "/data/data/uk.laurencegouws.zide/ub",
			cStringOnly: true,
		},
	} {
		next, changed := replaceFixedWidthCString(rewritten, []byte(replacement.old), []byte(replacement.new), replacement.cStringOnly)
		if changed {
			rewrites++
			rewritten = next
		}
	}
	return rewritten, rewrites
}

func PruneTermuxPrefixedBinaries(stagingRoot string) (int, error) {
	binDir := filepath.Join(stagingRoot, "usr", "bin")
	entries, err := os.ReadDir(binDir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "termux-") {
			continue
		}
		target := filepath.Join(binDir, name)
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

func replaceFixedWidthCString(payload []byte, old []byte, new []byte, cStringOnly bool) ([]byte, bool) {
	if len(new) > len(old) {
		return payload, false
	}
	changed := false
	searchFrom := 0
	for {
		index := bytes.Index(payload[searchFrom:], old)
		if index < 0 {
			break
		}
		start := searchFrom + index
		if cStringOnly {
			end := start + len(old)
			if end < len(payload) && payload[end] != 0 {
				searchFrom = start + 1
				continue
			}
		}
		copy(payload[start:start+len(new)], new)
		for i := start + len(new); i < start+len(old); i++ {
			payload[i] = 0
		}
		searchFrom = start + len(old)
		changed = true
	}
	return payload, changed
}

func rewriteTermuxLink(relative string, linkname string) (string, bool) {
	const oldAbsolute = "/data/data/com.termux/files/usr/"
	if strings.HasPrefix(linkname, oldAbsolute) {
		target := ZideUSRPrefix + strings.TrimPrefix(linkname, oldAbsolute)
		rel, err := filepath.Rel(filepath.Dir(relative), target)
		if err != nil {
			return linkname, false
		}
		return filepath.ToSlash(rel), true
	}
	if strings.HasPrefix(linkname, TermuxUSRPrefix) {
		target := ZideUSRPrefix + strings.TrimPrefix(linkname, TermuxUSRPrefix)
		rel, err := filepath.Rel(filepath.Dir(relative), target)
		if err != nil {
			return linkname, false
		}
		return filepath.ToSlash(rel), true
	}
	return linkname, false
}

func safeJoin(root string, relative string) (string, error) {
	clean := filepath.Clean(relative)
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return "", fmt.Errorf("unsafe archive path %q", relative)
	}
	target := filepath.Join(root, clean)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("archive path escapes staging root: %q", relative)
	}
	return target, nil
}

func looksText(bytes []byte) bool {
	limit := len(bytes)
	if limit > 8192 {
		limit = 8192
	}
	for _, b := range bytes[:limit] {
		if b == 0 {
			return false
		}
	}
	return true
}
