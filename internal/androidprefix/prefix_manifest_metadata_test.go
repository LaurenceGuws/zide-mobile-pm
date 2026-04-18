package androidprefix

import (
	"strings"
	"testing"
)

// Golden string must match historical zide-pm-admin emission so manifest
// consumers do not see surprise ordering or spelling changes.
const prefixRuntimeSupportLinksGolden = "/data/data/uk.laurencegouws.zide/.z=>/data/data/uk.laurencegouws.zide/files/usr,/data/user/0/uk.laurencegouws.zide/t/b=>/data/user/0/uk.laurencegouws.zide/files/usr/etc/bash.bashrc,/data/user/0/uk.laurencegouws.zide/t/p=>/data/user/0/uk.laurencegouws.zide/files/usr/etc/profile,/data/user/0/uk.laurencegouws.zide/t/h=>/data/user/0/uk.laurencegouws.zide/files/usr/etc/hosts,/data/user/0/uk.laurencegouws.zide/t/hs=>/data/user/0/uk.laurencegouws.zide/files/usr/var/htop/stat,/data/data/uk.laurencegouws.zide/ul=>/data/data/uk.laurencegouws.zide/files/usr/lib,/data/data/uk.laurencegouws.zide/ub=>/data/data/uk.laurencegouws.zide/files/usr/bin,/data/data/uk.laurencegouws.zide/b=>/data/data/uk.laurencegouws.zide/files/usr/bin,/data/data/uk.laurencegouws.zide/u/bsh=>/data/data/uk.laurencegouws.zide/files/usr/bin/sh"

const prefixRuntimeSupportFilesGolden = "/data/user/0/uk.laurencegouws.zide/t/b,/data/user/0/uk.laurencegouws.zide/t/p,/data/user/0/uk.laurencegouws.zide/t/h,/data/user/0/uk.laurencegouws.zide/t/hs"

func TestPrefixArchiveRuntimeSupportLinksGolden(t *testing.T) {
	got := PrefixArchiveRuntimeSupportLinks()
	if got != prefixRuntimeSupportLinksGolden {
		t.Fatalf("runtime_support_links drift\ngot:  %s\nwant: %s", got, prefixRuntimeSupportLinksGolden)
	}
}

func TestPrefixArchiveRuntimeSupportFilesGolden(t *testing.T) {
	got := PrefixArchiveRuntimeSupportFiles()
	if got != prefixRuntimeSupportFilesGolden {
		t.Fatalf("runtime_support_files drift\ngot:  %s\nwant: %s", got, prefixRuntimeSupportFilesGolden)
	}
}

func TestPrefixArchiveRuntimeSupportLinksEmbedsBridgeFirst(t *testing.T) {
	got := PrefixArchiveRuntimeSupportLinks()
	prefix := BinaryUSRBridgePath + "=>" + AppUSRPath
	if !strings.HasPrefix(got, prefix+",") {
		t.Fatalf("expected first pair %q with trailing comma, got %q", prefix, got)
	}
}
