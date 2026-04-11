package androidrepo

import (
	"strings"
	"testing"
)

const sampleIndex = `Package: bash
Architecture: aarch64
Version: 5.3.9
Depends: ncurses, readline (>= 8.3)
Filename: pool/main/b/bash/bash_5.3.9_aarch64.deb
Size: 100
SHA256: bash-sha

Package: ncurses
Architecture: aarch64
Version: 6.5
Filename: pool/main/n/ncurses/ncurses_6.5_aarch64.deb
Size: 200
SHA256: ncurses-sha

Package: readline
Architecture: aarch64
Version: 8.3
Depends: ncurses | ncurses-static
Filename: pool/main/r/readline/readline_8.3_aarch64.deb
Size: 300
SHA256: readline-sha
Description: line one
 continued line
`

func TestParseIndexAndResolveClosure(t *testing.T) {
	index, err := ParseIndex(strings.NewReader(sampleIndex))
	if err != nil {
		t.Fatal(err)
	}
	if got := index.Packages["readline"].Version; got != "8.3" {
		t.Fatalf("unexpected readline version: %s", got)
	}

	packages, err := ResolveClosure(index, []string{"bash"})
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(packages))
	for _, pkg := range packages {
		names = append(names, pkg.Name)
	}
	want := []string{"bash", "ncurses", "readline"}
	if len(names) != len(want) {
		t.Fatalf("names=%v want=%v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("names=%v want=%v", names, want)
		}
	}
}

func TestDependencyNames(t *testing.T) {
	got := DependencyNames("readline (>= 8.3), ncurses | ncurses-static, zlib:arm64")
	want := []string{"readline", "ncurses", "zlib"}
	if len(got) != len(want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got=%v want=%v", got, want)
		}
	}
}
