# Provider Model

Purpose: define how `zide-mobile-pm` treats upstream package/artifact sources
without making Zide a clone of any provider.

## Definition

A provider is a build-time source of artifact metadata and payloads.

Examples:

- an Android package repository
- a Zide-owned Android package feed
- an iOS-safe tool bundle source
- a signed internal artifact mirror

Providers are not runtime dependencies of Zide. Provider outputs must be
converted into Zide artifact contracts before any Zide app consumes them.

## Initial Provider

Provider id: `termux-main`

Role:

- Android development/bootstrap source
- useful for Bash, Neovim, Git, ripgrep, htop/gotop, and dependency closure
  experiments
- not automatically product-clean for `uk.laurencegouws.zide`

Trust boundary:

- package index is cached and checksummed
- package payloads are pinned by URL, size, and SHA-256
- generated prefix archives include audit metadata

Known limitation:

- upstream payloads may contain compiled-in `/data/data/com.termux/files/usr`
  assumptions
- product archives must remove, replace, or deliberately own those assumptions
- known fixed-width binary strings may be rewritten only when the replacement
  is explicit, app-owned, and recorded in artifact audit metadata

## Provider Metadata

Provider-derived artifacts should record:

- `provider`
- `provider_role`
- `provider_platform`
- `provider_architecture`

Provider-specific mechanics may also record package names, filenames,
dependency strings, repository names, or extraction rules.

Consumers should use provider metadata for provenance and compatibility checks,
not to infer runtime package-manager behavior.

## Dev snapshot channel (interim)

Until **MP-A6** closes with a product provider decision, the **authoritative
dev consumer manifest** for Zide Android bringup is whatever `android-dev-*`
GitHub prerelease `zide-pm-admin android-dev-snapshot-release` publishes most
recently. That lane always snapshots `termux-main`, materializes a prefix
archive under **hardcoded-prefix policy `audit`**, and ships the release-local
`android-dev-prefix.release.manifest.json` plus audit metadata.

`zide-pm` defaults to the latest published dev release manifest URL in code
(`DefaultAndroidDevManifestURL`). Product channels must not treat audit-mode
artifacts as product-clean; they remain development/bootstrap inputs under the
provider model above.

## Product Rule

Zide consumes artifact contracts, not provider internals.

Acceptable:

- `zide-mobile-pm` snapshots `termux-main`
- `zide-mobile-pm` materializes a dev prefix archive
- Zide stages an `android-prefix-archive` described by a manifest

Not acceptable:

- Zide parses `.deb` package internals
- Zide assumes Termux package layout is product-correct
- provider-specific paths leak into product docs as Zide-owned paths

## Future Providers

Likely future providers:

- `zide-android-main`: forked or rebuilt Android feed for product archives
- `zide-mobile-assets`: shared editor/IDE resource bundles
- `ios-bundled-tools`: iOS-safe artifacts with a separate execution model

Android and iOS may share the provider vocabulary. They should not share
execution assumptions.
