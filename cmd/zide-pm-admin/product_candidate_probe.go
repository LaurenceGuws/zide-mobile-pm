package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// androidProductCandidateProbe runs android-prefix-archive with
// hardcoded-policy=fail under a disposable temp directory. It is the MP-A6
// executable gate: exit 0 means the manifest/build inputs are product-candidate
// clean under fail policy; non-zero means hits remain (audit JSON is written).
func androidProductCandidateProbe(args []string) error {
	fs := flag.NewFlagSet("android-product-candidate-probe", flag.ExitOnError)
	manifestPath := fs.String("manifest", "dist/android-dev.manifest.json", "input MP-A1 Android dev manifest")
	cacheDir := fs.String("cache-dir", ".cache/android/packages", "downloaded package cache directory")
	auditOut := fs.String("audit-out", "dist/mp-a6-product-candidate.audit.json", "written on fail (and on pass); MP-A6 evidence path")
	zidePMBin := fs.String("zide-pm-bin", "", "optional Android zide-pm binary to include as usr/bin/zide-pm")
	if err := fs.Parse(args); err != nil {
		return err
	}

	td, err := os.MkdirTemp("", "zide-pm-admin-product-candidate-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(td) }()

	out := filepath.Join(td, "zide-android-dev-prefix.tar.gz")
	outManifest := filepath.Join(td, "android-dev-prefix.manifest.json")
	workDir := filepath.Join(td, "prefix-work")

	probeArgs := []string{
		"-manifest", *manifestPath,
		"-cache-dir", *cacheDir,
		"-work-dir", workDir,
		"-out", out,
		"-out-manifest", outManifest,
		"-audit-out", *auditOut,
		"-hardcoded-policy", "fail",
	}
	if *zidePMBin != "" {
		probeArgs = append(probeArgs, "-zide-pm-bin", *zidePMBin)
	}

	fmt.Printf("mp_a6_product_candidate_probe manifest=%s hardcoded_termux_policy=fail\n", *manifestPath)
	if err := androidPrefixArchive(probeArgs); err != nil {
		fmt.Printf("mp_a6_product_candidate=fail\n")
		fmt.Printf("mp_a6_product_candidate_audit=%s\n", *auditOut)
		return err
	}
	fmt.Println("mp_a6_product_candidate=pass")
	return nil
}
