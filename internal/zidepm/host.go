package zidepm

import (
	"os"
	"strings"
)

// EnvHostPlatform is the environment variable Zide sets for in-app zide-pm runs
// so catalog and test-binary install paths stay Android-scoped without forking
// the CLI surface.
const EnvHostPlatform = "ZIDE_PM_HOST_PLATFORM"

const (
	HostPlatformAndroid = "android"
	HostPlatformHost    = "host"
)

// CurrentHostPlatform returns the normalized host execution class for zide-pm.
// Empty or unset ZIDE_PM_HOST_PLATFORM means developer/generic host (not the
// Android in-app catalog mode).
func CurrentHostPlatform() string {
	v := strings.TrimSpace(os.Getenv(EnvHostPlatform))
	if v == "" {
		return HostPlatformHost
	}
	return strings.ToLower(v)
}

// AndroidCatalogActive reports whether Android-scoped catalog entries (such as
// android-test-binary) are visible and installable.
func AndroidCatalogActive() bool {
	return CurrentHostPlatform() == HostPlatformAndroid
}
