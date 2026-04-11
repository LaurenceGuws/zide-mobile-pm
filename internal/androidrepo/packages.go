// Package androidrepo parses Android package-index metadata used to build
// zide-mobile-pm Android artifact manifests.
package androidrepo

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

const (
	DefaultIndexURL = "https://packages.termux.dev/apt/termux-main/dists/stable/main/binary-aarch64/Packages"
	DefaultBaseURL  = "https://packages.termux.dev/apt/termux-main/"
)

type Package struct {
	Name         string
	Version      string
	Architecture string
	Filename     string
	Size         int64
	SHA256       string
	Depends      string
	PreDepends   string
}

type Index struct {
	Packages map[string]Package
}

func ParseIndex(reader io.Reader) (Index, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	index := Index{Packages: map[string]Package{}}
	fields := map[string]string{}
	var previousKey string
	lineNumber := 0

	flush := func() error {
		if len(fields) == 0 {
			return nil
		}
		pkg, err := packageFromFields(fields)
		if err != nil {
			return err
		}
		index.Packages[pkg.Name] = pkg
		fields = map[string]string{}
		previousKey = ""
		return nil
	}

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if line == "" {
			if err := flush(); err != nil {
				return Index{}, fmt.Errorf("paragraph ending at line %d: %w", lineNumber, err)
			}
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if previousKey == "" {
				return Index{}, fmt.Errorf("line %d: continuation without field", lineNumber)
			}
			fields[previousKey] += "\n" + strings.TrimSpace(line)
			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return Index{}, fmt.Errorf("line %d: invalid field", lineNumber)
		}
		key = strings.TrimSpace(key)
		fields[key] = strings.TrimSpace(value)
		previousKey = key
	}
	if err := scanner.Err(); err != nil {
		return Index{}, err
	}
	if err := flush(); err != nil {
		return Index{}, err
	}

	return index, nil
}

func ResolveClosure(index Index, roots []string) ([]Package, error) {
	seen := map[string]bool{}
	var ordered []Package

	var visit func(name string) error
	visit = func(name string) error {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			return nil
		}
		pkg, ok := index.Packages[name]
		if !ok {
			return fmt.Errorf("package %q not found in index", name)
		}
		seen[name] = true
		for _, dep := range DependencyNames(pkg.PreDepends) {
			if err := visit(dep); err != nil {
				return err
			}
		}
		for _, dep := range DependencyNames(pkg.Depends) {
			if err := visit(dep); err != nil {
				return err
			}
		}
		ordered = append(ordered, pkg)
		return nil
	}

	for _, root := range roots {
		if err := visit(root); err != nil {
			return nil, err
		}
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Name < ordered[j].Name
	})
	return ordered, nil
}

func DependencyNames(depends string) []string {
	if strings.TrimSpace(depends) == "" {
		return nil
	}

	var names []string
	seen := map[string]bool{}
	for _, group := range strings.Split(depends, ",") {
		firstAlternative := strings.TrimSpace(strings.Split(group, "|")[0])
		name := normalizeDependencyName(firstAlternative)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names
}

func HashBytes(bytes []byte) string {
	sum := sha256.Sum256(bytes)
	return hex.EncodeToString(sum[:])
}

func AbsolutePackageURL(baseURL string, filename string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	relative, err := url.Parse(filename)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(relative).String(), nil
}

func packageFromFields(fields map[string]string) (Package, error) {
	name := fields["Package"]
	if name == "" {
		return Package{}, fmt.Errorf("missing Package")
	}
	size, err := strconv.ParseInt(fields["Size"], 10, 64)
	if err != nil {
		return Package{}, fmt.Errorf("%s: invalid Size %q", name, fields["Size"])
	}
	pkg := Package{
		Name:         name,
		Version:      fields["Version"],
		Architecture: fields["Architecture"],
		Filename:     fields["Filename"],
		Size:         size,
		SHA256:       fields["SHA256"],
		Depends:      fields["Depends"],
		PreDepends:   fields["Pre-Depends"],
	}
	if pkg.Version == "" {
		return Package{}, fmt.Errorf("%s: missing Version", name)
	}
	if pkg.Filename == "" {
		return Package{}, fmt.Errorf("%s: missing Filename", name)
	}
	if pkg.SHA256 == "" {
		return Package{}, fmt.Errorf("%s: missing SHA256", name)
	}
	return pkg, nil
}

func normalizeDependencyName(dep string) string {
	dep = strings.TrimSpace(dep)
	if dep == "" {
		return ""
	}
	if before, _, ok := strings.Cut(dep, " "); ok {
		dep = before
	}
	if before, _, ok := strings.Cut(dep, "("); ok {
		dep = before
	}
	if before, _, ok := strings.Cut(dep, ":"); ok {
		dep = before
	}
	return strings.TrimSpace(dep)
}
