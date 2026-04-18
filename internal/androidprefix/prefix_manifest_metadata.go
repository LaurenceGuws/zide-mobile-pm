package androidprefix

import "strings"

// PrefixArchiveRuntimeSupportFiles returns the comma-separated
// metadata.runtime_support_files value for android-prefix-archive manifests
// produced by this repo (paths under RuntimeAliasDir used by fixed-width binary
// rewrites for etc/ and htop support files).
func PrefixArchiveRuntimeSupportFiles() string {
	return strings.Join([]string{
		RuntimeAliasDir + "/b",
		RuntimeAliasDir + "/p",
		RuntimeAliasDir + "/h",
		RuntimeAliasDir + "/hs",
	}, ",")
}

// PrefixArchiveRuntimeSupportLinks returns the metadata.runtime_support_links
// CSV for android-prefix-archive manifests. The first pair is always the
// BinaryUSRBridgePath symlink source mapped to AppUSRPath so it stays aligned
// with rewriteBinaryUSRRootToBridge in deb.go.
func PrefixArchiveRuntimeSupportLinks() string {
	data := "/data/data/" + AppPackageName
	userFiles := "/data/user/0/" + AppPackageName + "/files/usr"
	parts := []string{
		BinaryUSRBridgePath + "=>" + AppUSRPath,
		RuntimeAliasDir + "/b=>" + userFiles + "/etc/bash.bashrc",
		RuntimeAliasDir + "/p=>" + userFiles + "/etc/profile",
		RuntimeAliasDir + "/h=>" + userFiles + "/etc/hosts",
		RuntimeAliasDir + "/hs=>" + userFiles + "/var/htop/stat",
		data + "/ul=>" + data + "/files/usr/lib",
		data + "/ub=>" + data + "/files/usr/bin",
		data + "/b=>" + data + "/files/usr/bin",
		data + "/u/bsh=>" + data + "/files/usr/bin/sh",
	}
	return strings.Join(parts, ",")
}
