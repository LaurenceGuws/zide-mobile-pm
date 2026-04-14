// Command zide-pm is the user-facing Zide mobile package CLI.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LaurenceGuws/zide-mobile-pm/internal/zidepm"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	var err error
	switch os.Args[1] {
	case "help", "-h", "--help":
		printHelp()
	case "version":
		fmt.Println(version)
	case "doctor":
		err = doctor(os.Args[2:])
	case "list-available", "list":
		err = listAvailable(os.Args[2:])
	case "install":
		err = install(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`zide-pm manages Zide mobile packages from pinned artifact manifests.

Usage:
  zide-pm <command> [options]

Commands:
  doctor          Validate the configured package manifest and print provider info.
  list-available  List packages/groups available from the manifest.
  install         Install a supported package/group into a prefix.
  version         Print the tool version.
  help            Show this help.

Current MVP package:
  dev-baseline    Bash + Neovim + Git + ripgrep + htop + gotop baseline.

Examples:
  zide-pm doctor
  zide-pm list-available
  zide-pm install dev-baseline --prefix /data/data/uk.laurencegouws.zide/files/usr
  zide-pm install dev-baseline --manifest ./android-dev-prefix.release.manifest.json --prefix ./tmp/usr

zide-pm is the product CLI surface. Provider/package internals stay behind the
manifest contract.`)
}

func doctor(args []string) error {
	fs := commonFlagSet("doctor")
	manifestPath := fs.String("manifest", zidepm.DefaultAndroidDevManifestURL, "artifact manifest URL/path")
	prefix := fs.String("prefix", defaultPrefix(), "installed prefix used for offline doctor fallback")
	if err := fs.Parse(args); err != nil {
		return err
	}
	source, err := loadSource(*manifestPath)
	if err != nil {
		return doctorInstalled(*prefix, err)
	}
	fmt.Printf("manifest=%s\n", source.Location)
	fmt.Printf("platform=%s\n", source.Document.Platform)
	fmt.Printf("channel=%s\n", source.Document.Channel)
	if artifact, err := zidepm.AndroidPrefixArtifact(source); err == nil {
		fmt.Printf("artifact=%s\n", artifact.Artifact.Name)
		fmt.Printf("version=%s\n", artifact.Artifact.Version)
		fmt.Printf("provider=%s\n", artifact.Artifact.Metadata["provider"])
		fmt.Printf("provider_role=%s\n", artifact.Artifact.Metadata["provider_role"])
		fmt.Printf("archive_url=%s\n", artifact.URL)
		fmt.Printf("hardcoded_termux_policy=%s\n", artifact.Artifact.Metadata["hardcoded_termux_policy"])
	}
	fmt.Println("ok=true")
	return nil
}

func listAvailable(args []string) error {
	fs := commonFlagSet("list-available")
	manifestPath := fs.String("manifest", zidepm.DefaultAndroidDevManifestURL, "artifact manifest URL/path")
	prefix := fs.String("prefix", defaultPrefix(), "installed prefix used for offline list fallback")
	if err := fs.Parse(args); err != nil {
		return err
	}
	source, err := loadSource(*manifestPath)
	if err != nil {
		return listInstalled(*prefix, err)
	}
	for _, name := range zidepm.AvailablePackages(source) {
		fmt.Println(name)
	}
	return nil
}

func install(args []string) error {
	fs := commonFlagSet("install")
	manifestPath := fs.String("manifest", zidepm.DefaultAndroidDevManifestURL, "artifact manifest URL/path")
	prefix := fs.String("prefix", "", "installation prefix, required")
	cacheDir := fs.String("cache-dir", zidepm.DefaultCacheDir(), "download/cache directory")
	args = reorderFlags(args, map[string]bool{
		"manifest":  true,
		"prefix":    true,
		"cache-dir": true,
	})
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("install expects exactly one package")
	}
	pkg := fs.Arg(0)
	if pkg != zidepm.DevBaselinePackage {
		return fmt.Errorf("unsupported MVP package %q; supported: %s", pkg, zidepm.DevBaselinePackage)
	}
	if strings.TrimSpace(*prefix) == "" {
		return fmt.Errorf("--prefix is required")
	}

	source, err := loadSource(*manifestPath)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	result, err := zidepm.InstallDevBaseline(ctx, source, *prefix, *cacheDir)
	if err != nil {
		return err
	}
	fmt.Printf("installed=%s\n", result.Package)
	fmt.Printf("prefix=%s\n", result.Prefix)
	fmt.Printf("provider=%s\n", result.Provider)
	fmt.Printf("version=%s\n", result.Version)
	fmt.Printf("files=%d\n", result.FileCount)
	fmt.Printf("dirs=%d\n", result.DirCount)
	fmt.Printf("symlinks=%d\n", result.SymlinkCount)
	return nil
}

func commonFlagSet(name string) *flag.FlagSet {
	return flag.NewFlagSet(name, flag.ExitOnError)
}

func doctorInstalled(prefix string, manifestErr error) error {
	stamp, err := zidepm.LoadInstallStamp(prefix)
	if err != nil {
		return fmt.Errorf("manifest unavailable (%v) and no install stamp at prefix %q: %w", manifestErr, prefix, err)
	}
	fmt.Printf("manifest=%s\n", stamp.Manifest)
	fmt.Printf("installed=true\n")
	fmt.Printf("package=%s\n", stamp.Package)
	fmt.Printf("artifact=%s\n", stamp.Artifact)
	fmt.Printf("version=%s\n", stamp.Version)
	fmt.Printf("provider=%s\n", stamp.Provider)
	fmt.Printf("prefix=%s\n", prefix)
	fmt.Printf("files=%d\n", stamp.Files)
	fmt.Printf("symlinks=%d\n", stamp.Symlinks)
	fmt.Println("ok=true")
	return nil
}

func listInstalled(prefix string, manifestErr error) error {
	stamp, err := zidepm.LoadInstallStamp(prefix)
	if err != nil {
		return fmt.Errorf("manifest unavailable (%v) and no install stamp at prefix %q: %w", manifestErr, prefix, err)
	}
	fmt.Println(stamp.Package)
	return nil
}

func defaultPrefix() string {
	if value := os.Getenv("PREFIX"); value != "" {
		return value
	}
	return ""
}

func reorderFlags(args []string, takesValue map[string]bool) []string {
	var flags []string
	var positionals []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positionals = append(positionals, arg)
			continue
		}
		flags = append(flags, arg)
		name := strings.TrimLeft(arg, "-")
		if before, _, ok := strings.Cut(name, "="); ok {
			name = before
		}
		if strings.Contains(arg, "=") || !takesValue[name] {
			continue
		}
		if i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}
	return append(flags, positionals...)
}

func loadSource(location string) (zidepm.Source, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	return zidepm.LoadSource(ctx, location)
}
