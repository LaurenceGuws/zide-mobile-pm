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
	if err := writeSampleDeb(debPath, sampleDebEntries()); err != nil {
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
	if got := string(script); got != "/data/data/uk.laurencegouws.zide/files/usr/bin\n" {
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

func TestExtractDebUSRRewritesKnownBinaryTermuxPaths(t *testing.T) {
	debPath := filepath.Join(t.TempDir(), "sample.deb")
	body := append([]byte{0x7f, 'E', 'L', 'F', 0}, []byte("/data/data/com.termux/files/usr/var/htop/stat")...)
	body = append(body, 0, 1, 2, 3)
	if err := writeSampleDeb(debPath, []sampleDebEntry{{
		name:     "data/data/com.termux/files/usr/bin/htop",
		mode:     0o755,
		body:     body,
		typeflag: tar.TypeReg,
	}}); err != nil {
		t.Fatal(err)
	}

	stagingRoot := t.TempDir()
	stats, err := ExtractDebUSR(debPath, stagingRoot)
	if err != nil {
		t.Fatal(err)
	}
	if stats.BinaryRewrites != 1 {
		t.Fatalf("binary rewrites=%d want=1", stats.BinaryRewrites)
	}
	if len(stats.HardcodedTermuxHits) != 0 {
		t.Fatalf("unexpected hardcoded hits: %v", stats.HardcodedTermuxHits)
	}

	rewritten, err := os.ReadFile(filepath.Join(stagingRoot, "usr/bin/htop"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(rewritten, []byte("/data/data/com.termux/files/usr/var/htop/stat")) {
		t.Fatalf("old htop path remained in binary payload: %q", rewritten)
	}
	if !bytes.Contains(rewritten, []byte("/data/user/0/uk.laurencegouws.zide/t/hs")) {
		t.Fatalf("new htop path missing from binary payload: %q", rewritten)
	}
}

func TestPruneTermuxPrefixedBinaries(t *testing.T) {
	stagingRoot := t.TempDir()
	binDir := filepath.Join(stagingRoot, "usr/bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "termux-open"), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "bash"), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}

	removed, err := PruneTermuxPrefixedBinaries(stagingRoot)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("removed=%d want=1", removed)
	}
	if _, err := os.Stat(filepath.Join(binDir, "termux-open")); !os.IsNotExist(err) {
		t.Fatalf("expected termux-open removed, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(binDir, "bash")); err != nil {
		t.Fatalf("expected bash preserved, err=%v", err)
	}
}

func TestReplaceFixedWidthCStringCStringOnlySkipsExtendedPath(t *testing.T) {
	payload := []byte("/data/data/com.termux/files/usr/lib/extra\x00")
	old := []byte("/data/data/com.termux/files/usr/lib")
	newPath := []byte("/data/data/uk.laurencegouws.zide/ul")
	got, changed := replaceFixedWidthCString(append([]byte(nil), payload...), old, newPath, true)
	if changed {
		t.Fatal("expected no rewrite when '/' follows usr/lib")
	}
	if string(got) != string(payload) {
		t.Fatalf("payload mutated: %q", got)
	}
}

func TestExtractDebUSRRewritesUnknownBinaryTermuxUSRPrefix(t *testing.T) {
	debPath := filepath.Join(t.TempDir(), "sample.deb")
	body := append([]byte{0x7f, 'E', 'L', 'F', 0}, []byte("/data/data/com.termux/files/usr/lib/unknown")...)
	if err := writeSampleDeb(debPath, []sampleDebEntry{{
		name:     "data/data/com.termux/files/usr/bin/unknown",
		mode:     0o755,
		body:     body,
		typeflag: tar.TypeReg,
	}}); err != nil {
		t.Fatal(err)
	}

	stagingRoot := t.TempDir()
	stats, err := ExtractDebUSR(debPath, stagingRoot)
	if err != nil {
		t.Fatal(err)
	}
	if stats.BinaryRewrites < 1 {
		t.Fatalf("binary rewrites=%d want>=1", stats.BinaryRewrites)
	}
	if len(stats.HardcodedTermuxHits) != 0 {
		t.Fatalf("unexpected hardcoded hits: %v", stats.HardcodedTermuxHits)
	}

	rewritten, err := os.ReadFile(filepath.Join(stagingRoot, "usr/bin/unknown"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(rewritten, []byte("/data/data/com.termux/files/usr")) {
		t.Fatalf("termux usr root remained in binary: %q", rewritten)
	}
	if !bytes.Contains(rewritten, []byte("/data/data/zide.embed/files/usr")) {
		t.Fatalf("embed usr root missing from binary: %q", rewritten)
	}
}

type sampleDebEntry struct {
	name     string
	mode     int64
	body     []byte
	linkname string
	typeflag byte
}

func sampleDebEntries() []sampleDebEntry {
	return []sampleDebEntry{
		{name: "data/data/com.termux/files/usr/bin/", mode: 0o755, typeflag: tar.TypeDir},
		{name: "data/data/com.termux/files/usr/bin/sample", mode: 0o755, body: []byte("/data/data/com.termux/files/usr/bin\n"), typeflag: tar.TypeReg},
		{name: "data/data/com.termux/files/usr/bin/sample-link", mode: 0o777, linkname: "/data/data/com.termux/files/usr/bin/sample", typeflag: tar.TypeSymlink},
	}
}

func writeSampleDeb(path string, entries []sampleDebEntry) error {
	var tarPayload bytes.Buffer
	gzipWriter := gzip.NewWriter(&tarPayload)
	tarWriter := tar.NewWriter(gzipWriter)
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
		if len(entry.body) > 0 {
			if _, err := tarWriter.Write(entry.body); err != nil {
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
