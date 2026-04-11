# zide-mobile-pm

Mobile package and artifact authority for the Zide project family.

This repo starts Android-first because Zide's Android terminal needs a real
app-private Bash/tool userland. The boundary is intentionally broader than
Android: future mobile products may need iOS-safe tool bundles, editor/IDE
assets, LSP bundles, or other platform-specific artifacts. This repo owns those
artifact contracts without forcing every platform to pretend it has the same
package model.

## Scope

This repo owns:

- mobile artifact manifests
- provider metadata and trust boundaries
- bootstrap/package artifact metadata
- checksum and version authority
- package index snapshots
- host-side artifact materialization tooling
- release/channel docs for mobile packages and tool bundles

This repo does not own:

- Zide terminal rendering
- Zide PTY/session runtime
- Android Java/Kotlin lifecycle or UI integration
- iOS app runtime policy
- package execution inside mobile apps
- the core Zide Zig codebase

Those stay in `zide` or platform-native app repos.

## Current Status

Initial Android dev manifest support is in place.

The first real consumer is:

- `../zide`
- Android package name: `dev.zide.terminal`
- Android prefix: `/data/data/dev.zide.terminal/files/usr`
- Android execution posture: Termux-compatible target SDK 28 until a modern
  userland execution model is proven

Current CLI surface:

```bash
go run ./cmd/zide-mobile-pm help
go run ./cmd/zide-mobile-pm contract -platform android
go run ./cmd/zide-mobile-pm android-dev-manifest
go run ./cmd/zide-mobile-pm android-prefix-archive -hardcoded-policy audit
go run ./cmd/zide-mobile-pm android-dev-release -dry-run
go run ./cmd/zide-mobile-pm validate examples/android-dev.manifest.json
```

`android-dev-manifest` fetches or reuses the cached Termux main aarch64 package
index through the `termux-main` provider, verifies the cache checksum, resolves
the default dev package roots `bash,neovim,htop,gotop`, and writes pinned
package metadata to `dist/android-dev.manifest.json`.

`android-prefix-archive` consumes that manifest, downloads and verifies the
pinned `.deb` payloads, extracts only `data/data/com.termux/files/usr/*`, and
writes an archive rooted at `usr/`. Its default `-hardcoded-policy fail` is
strict; use `-hardcoded-policy audit` only for the current development archive
while remaining compiled-in Termux prefix hits are still being reviewed.

`android-dev-release` automates the fast development release lane. It generates
the dev manifest and prefix archive in audit mode, rewrites the release manifest
so the archive URL is relative to the manifest location, creates a tag, and
publishes a GitHub prerelease with the generated assets. Use `-dry-run` to
prepare assets without publishing.

## Design Rule

Do not couple platforms through implementation details.

Android may use executable prefix archives and package indexes. iOS likely will
not. The shared layer is the artifact contract:

- platform
- channel
- version
- URL
- checksum
- size
- compatibility metadata
- limitations

Platform-specific package mechanics live behind that contract.

Providers are build-time input sources. `termux-main` is the first Android
provider and is currently a dev/bootstrap source, not Zide's product package
manager. Future providers can include a Zide-owned Android feed, signed mirrors,
or iOS-safe bundle sources without changing the consumer contract.

## Near-Term Plan

1. Decide whether product Android packages come from the current `termux-main`
   provider, a Zide-owned provider, or a controlled mirror/fork.
2. Remove or formally own remaining compiled-in Termux prefix assumptions for
   product archives.
3. Add iOS artifact-contract notes only when a concrete iOS consumer exists.

## Related Projects

- `zide`: runtime consumer for Android terminal/userland artifacts.
- `zide-tree-sitter`: grammar-pack producer for Zide-family consumers.
- `zlua-portable`: reusable Zig Lua embedding helper package.
- `docs-explorer`: shared local documentation explorer.
