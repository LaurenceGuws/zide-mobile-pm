package main

import (
	"path/filepath"
	"testing"
)

func TestProductCandidateOutputDirConstant(t *testing.T) {
	if filepath.Clean(productCandidateOutputDir) != "dist/product-candidate" {
		t.Fatalf("unexpected dir %q", productCandidateOutputDir)
	}
}
