package androidprefix

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractDebUSRRewritesPrefixPaths(t *testing.T) {
	debPath := filepath.Join(t.TempDir(), "sample.deb")
	if err := writeSampleDeb(debPath); err != nil {
		t.Fatal(err)
	}

	stagingRoot := t.TempDir()
	stats, err := ExtractDebUSR(debPath, stagingRoot)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Files != 1 {
		t.Fatalf("files=%d want=1", stats.Files)
	}
	if stats.Symlinks != 1 {
		t.Fatalf("symlinks=%d want=1", stats.Symlinks)
	}
	if stats.TextRewrites != 2 {
		t.Fatalf("text rewrites=%d want=2", stats.TextRewrites)
	}
	if len(stats.HardcodedTermuxHits) != 0 {
		t.Fatalf("unexpected hardcoded hits: %v", stats.HardcodedTermuxHits)
	}

	script, err := os.ReadFile(filepath.Join(stagingRoot, "usr/bin/sample"))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(script); got != "/data/data/dev.zide.terminal/files/usr/bin\n" {
		t.Fatalf("unexpected rewritten file: %q", got)
	}
	link, err := os.Readlink(filepath.Join(stagingRoot, "usr/bin/sample-link"))
	if err != nil {
		t.Fatal(err)
	}
	if link != "sample" {
		t.Fatalf("unexpected symlink target: %q", link)
	}
}

func writeSampleDeb(path string) error {
	var tarPayload bytes.Buffer
	gzipWriter := gzip.NewWriter(&tarPayload)
	tarWriter := tar.NewWriter(gzipWriter)
	entries := []struct {
		name     string
		mode     int64
		body     string
		linkname string
		typeflag byte
	}{
		{name: "data/data/com.termux/files/usr/bin/", mode: 0o755, typeflag: tar.TypeDir},
		{name: "data/data/com.termux/files/usr/bin/sample", mode: 0o755, body: "/data/data/com.termux/files/usr/bin\n", typeflag: tar.TypeReg},
		{name: "data/data/com.termux/files/usr/bin/sample-link", mode: 0o777, linkname: "/data/data/com.termux/files/usr/bin/sample", typeflag: tar.TypeSymlink},
	}
	for _, entry := range entries {
		header := &tar.Header{
			Name:     entry.name,
			Mode:     entry.mode,
			Size:     int64(len(entry.body)),
			Typeflag: entry.typeflag,
			Linkname: entry.linkname,
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if entry.body != "" {
			if _, err := tarWriter.Write([]byte(entry.body)); err != nil {
				return err
			}
		}
	}
	if err := tarWriter.Close(); err != nil {
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		return err
	}

	var deb bytes.Buffer
	deb.WriteString("!<arch>\n")
	writeArMember(&deb, "debian-binary", []byte("2.0\n"))
	writeArMember(&deb, "control.tar.gz", []byte{})
	writeArMember(&deb, "data.tar.gz", tarPayload.Bytes())
	return os.WriteFile(path, deb.Bytes(), 0o644)
}

func writeArMember(buffer *bytes.Buffer, name string, payload []byte) {
	header := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8o%-10d`\n", name, 0, 0, 0, 0o644, len(payload))
	buffer.WriteString(header)
	buffer.Write(payload)
	if len(payload)%2 != 0 {
		buffer.WriteByte('\n')
	}
}
