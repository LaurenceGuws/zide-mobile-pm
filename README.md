# zide-mobile-pm

Mobile package authority for the Zide project family.

This repo starts Android-first because Zide's Android terminal needs a real
app-private Bash/tool userland. The boundary is intentionally broader than
Android: future mobile products may need iOS-safe tool bundles, editor/IDE
assets, LSP bundles, or other platform-specific artifacts. This repo owns those
artifact contracts without forcing every platform to pretend it has the same
package model.

## Scope

This repo owns:

- `zide-pm`, the user-facing mobile package CLI
- `zide-pm-admin`, the backend/admin tool for manifests, archives, and release
  publishing
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
- Android package name: `uk.laurencegouws.zide`
- Android prefix: `/data/data/uk.laurencegouws.zide/files/usr`
- Android execution posture: Termux-compatible target SDK 28 until a modern
  userland execution model is proven

Current CLI surface:

```bash
go run ./cmd/zide-pm help
go run ./cmd/zide-pm doctor
go run ./cmd/zide-pm list-available
go run ./cmd/zide-pm install dev-baseline --prefix ./tmp/usr

go run ./cmd/zide-pm-admin help
go run ./cmd/zide-pm-admin contract -platform android
go run ./cmd/zide-pm-admin android-dev-manifest
go run ./cmd/zide-pm-admin android-prefix-archive -hardcoded-policy audit
go run ./cmd/zide-pm-admin android-dev-snapshot-release -dry-run
go run ./cmd/zide-pm-admin android-product-candidate-probe
go run ./cmd/zide-pm-admin validate examples/android-dev.manifest.json
```

`zide-pm` is the product CLI surface. The MVP supports `doctor`,
`list-available`, and `install dev-baseline --prefix <path>`. It consumes the
same manifest/archive contract as Zide and does not parse provider package
internals.

`dev-baseline` is current bringup naming for the first bootstrap/recommended
profile. It should not be treated as the final long-term product name for
onboarding/default package policy.

`zide-pm-admin` is the backend/admin tool. It owns provider snapshotting,
manifest generation, prefix archive production, validation, and dev snapshot
publishing. It is not the product shell command.

Android dev prefix archives now include `usr/bin/zide-pm` and an install stamp,
so `zide-pm doctor` and `zide-pm list-available` work inside the staged
app-private shell without requiring private GitHub release access from the
device.

`android-dev-manifest` fetches or reuses the cached Termux main aarch64 package
index through the `termux-main` provider, verifies the cache checksum, resolves
the default dev package roots `bash,neovim,git,ripgrep,htop,gotop`, and writes
pinned package metadata to `dist/android-dev.manifest.json`.

`android-prefix-archive` consumes that manifest, downloads and verifies the
pinned `.deb` payloads, extracts only `data/data/com.termux/files/usr/*`, and
writes an archive rooted at `usr/`. Its default `-hardcoded-policy fail` is
strict; use `-hardcoded-policy audit` only for the current development archive
while remaining compiled-in Termux prefix hits are still being reviewed. Known
compiled provider paths may be rewritten only through explicit fixed-width
binary rules, and those rewrites are reported separately from text rewrites.

`android-dev-snapshot-release` automates the fast Android development snapshot
prerelease lane. It generates
the dev manifest and prefix archive in audit mode, rewrites the release manifest
so the archive URL is relative to the manifest location, adds a pinned
`android-test-binary` (`zide-android-catalog-smoke`) for catalog-mode validation,
creates a tag, and
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

`zide-pm` is the product-facing package command intended to run inside the Zide
shell. Provider commands and release commands are backend machinery beneath
that UX.

## Near-Term Plan

1. Decide whether product Android packages come from the current `termux-main`
   provider, a Zide-owned provider, or a controlled mirror/fork. Decision space
   and interim dev authority: `app_architecture/ANDROID_PRODUCT_PROVIDER_DECISION.md`.
   Zide runtime staging contract: `app_architecture/ZIDE_MOBILE_ARTIFACT_CONSUMER.md`.
2. Remove or formally own remaining compiled-in Termux prefix assumptions for
   product archives.
3. Add real on-device `zide-pm install/remove/upgrade` semantics after the
   product provider decision.
4. Add iOS artifact-contract notes only when a concrete iOS consumer exists.

## Related Projects

- `zide`: runtime consumer for Android terminal/userland artifacts.
- `zide-tree-sitter`: grammar-pack producer for Zide-family consumers.
- `zlua-portable`: reusable Zig Lua embedding helper package.
- `docs-explorer`: shared local documentation explorer.
