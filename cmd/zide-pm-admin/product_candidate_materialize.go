package main

import (
	"flag"
	"fmt"
	"path/filepath"
)

const productCandidateOutputDir = "dist/product-candidate"

// androidProductCandidateMaterialize runs android-prefix-archive with
// hardcoded-policy=fail and default outputs under dist/product-candidate/.
// It is the MP-A6 materialization step: on success, leaves tarball + prefix
// manifest + audit for downstream packaging or zide staging experiments.
func androidProductCandidateMaterialize(args []string) error {
	fs := flag.NewFlagSet("android-product-candidate-materialize", flag.ExitOnError)
	manifestPath := fs.String("manifest", "dist/android-dev.manifest.json", "input MP-A1 Android dev manifest")
	cacheDir := fs.String("cache-dir", ".cache/android/packages", "downloaded package cache directory")
	workDir := fs.String("work-dir", ".cache/android/product-candidate-work", "temporary extraction directory")
	out := fs.String("out", filepath.Join(productCandidateOutputDir, "zide-android-prefix.tar.gz"), "output prefix archive path")
	outManifest := fs.String("out-manifest", filepath.Join(productCandidateOutputDir, "android-prefix.manifest.json"), "output archive manifest path")
	auditOut := fs.String("audit-out", filepath.Join(productCandidateOutputDir, "prefix.audit.json"), "output archive audit path")
	zidePMBin := fs.String("zide-pm-bin", "", "optional Android zide-pm binary to include as usr/bin/zide-pm")
	if err := fs.Parse(args); err != nil {
		return err
	}

	archiveArgs := []string{
		"-manifest", *manifestPath,
		"-cache-dir", *cacheDir,
		"-work-dir", *workDir,
		"-out", *out,
		"-out-manifest", *outManifest,
		"-audit-out", *auditOut,
		"-hardcoded-policy", "fail",
	}
	if *zidePMBin != "" {
		archiveArgs = append(archiveArgs, "-zide-pm-bin", *zidePMBin)
	}

	fmt.Printf("mp_a6_product_candidate_materialize manifest=%s hardcoded_termux_policy=fail\n", *manifestPath)
	fmt.Printf("mp_a6_product_candidate_outputs dir=%s\n", filepath.Dir(*out))
	if err := androidPrefixArchive(archiveArgs); err != nil {
		fmt.Printf("mp_a6_product_candidate_materialize=fail audit=%s\n", *auditOut)
		return err
	}
	fmt.Println("mp_a6_product_candidate_materialize=pass")
	return nil
}
