# Releases

Release notes and published artifact records live here once `zide-pm-admin`
starts producing versioned mobile artifacts.

## Android Dev Releases

Android dev snapshot releases are automated prereleases for Zide Android terminal
bringup.

They are not official/product release artifacts.

Command:

```bash
go run ./cmd/zide-pm-admin android-dev-snapshot-release
```

Use `-dry-run` before publishing when changing release plumbing.

Dev snapshot release assets:

- `android-dev-prefix.release.manifest.json`
- `android-dev.manifest.json`
- `zide-android-dev-prefix.tar.gz`
- `zide-android-catalog-smoke.sh` (pinned `android-test-binary` for catalog mode)
- `zide-android-dev-prefix.audit.json`

The prefix archive includes the Android `zide-pm` binary and a package install
stamp for offline on-device `doctor` / `list-available` support.

Dev snapshot release policy:

- provider: `termux-main`
- provider role: `android-dev-bootstrap`
- hardcoded prefix policy: `audit`
- GitHub release type: prerelease

The release manifest uses an archive URL relative to the manifest location so
Zide can consume either a release URL or a local file override with the same
resolution rule.

## Formal Product Releases

No product artifacts are published yet.

Formal product releases must not use audit mode as a substitute for product review.
They must pass the stricter provider/source decision and hardcoded-prefix
policy documented in `docs/todo/implementation.md`.
