# Releases

Release notes and published artifact records live here once `zide-mobile-pm`
starts producing versioned mobile artifacts.

## Android Dev Releases

Android dev releases are automated prereleases for Zide Android terminal
bringup.

Command:

```bash
go run ./cmd/zide-mobile-pm android-dev-release
```

Use `-dry-run` before publishing when changing release plumbing.

Dev release assets:

- `android-dev-prefix.release.manifest.json`
- `android-dev.manifest.json`
- `zide-android-dev-prefix.tar.gz`
- `zide-android-dev-prefix.audit.json`

Dev release policy:

- provider: `termux-main`
- provider role: `android-dev-bootstrap`
- hardcoded prefix policy: `audit`
- GitHub release type: prerelease

The release manifest uses an archive URL relative to the manifest location so
Zide can consume either a release URL or a local file override with the same
resolution rule.

## Formal Releases

No product artifacts are published yet.

Formal releases must not use audit mode as a substitute for product review.
They must pass the stricter provider/source decision and hardcoded-prefix
policy documented in `docs/todo/implementation.md`.
