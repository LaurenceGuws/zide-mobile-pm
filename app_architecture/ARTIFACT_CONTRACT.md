# Artifact Contract

Purpose: define the language-neutral contract that `zide` and other consumers
can rely on without importing this repo's implementation.

## Manifest

Current schema version: `1`

Required top-level fields:

- `schema_version`
- `project`
- `platform`
- `channel`
- `artifacts`

Required artifact fields:

- `name`
- `kind`
- `version`
- `url`
- `sha256`
- `size`

Optional artifact fields:

- `metadata`
- `limitations`

## Platform Values

Supported initial platform values:

- `android`
- `ios`

Android and iOS artifacts may have different `kind` values. Consumers must not
infer that a kind from one platform is valid for another platform.

## Android Initial Kinds

- `android-prefix-archive`
- `android-termux-package-index`
- `android-termux-deb`
- `android-test-binary`

Meaning:

- `android-prefix-archive`: an archive that materializes a complete or partial
  app-private prefix. Intended consumer prefix, package name, and target SDK
  assumptions are recorded in metadata. Archive root is currently `usr/` for
  Android dev archives, intended to be staged under the app files directory.
- `android-termux-package-index`: a pinned package index snapshot used as
  provider source metadata for a dev channel.
- `android-termux-deb`: a pinned upstream package payload. This is source
  material for host-side artifact production, not a direct product runtime
  promise.
- `android-test-binary`: a single pinned file (executable or data) installed
  under the app-private prefix at `metadata.install_relative_path` (relative to
  the prefix root, same layout as extracted `usr/` contents). Intended for
  Android catalog mode only: `zide-pm` lists and installs these when
  `ZIDE_PM_HOST_PLATFORM=android`. iOS does not reuse this kind.

The current kind names are intentionally specific because they describe the
payload format. Provider identity is recorded separately in metadata.

## Providers

A provider is a build-time source of artifact metadata and payloads.

Required provider metadata for provider-derived artifacts:

- `provider`
- `provider_role`
- `provider_platform`
- `provider_architecture`

Initial provider:

- `termux-main`

Initial provider role:

- `android-dev-bootstrap`

Initial Android metadata keys:

- `package_name`
- `prefix`
- `target_sdk`
- `architecture`
- `filename`
- `package`
- `depends`
- `pre_depends`
- `archive_root`
- `source_manifest_sha256`
- `hardcoded_termux_hits`
- `hardcoded_termux_policy`
- `text_rewrites`
- `binary_rewrites`
- `runtime_support_files`
- `runtime_support_links`
- `install_relative_path` (required for `android-test-binary`; relative path under the prefix)
- `unix_mode` (optional octal mode for installed test binary files, default `0755`)

`hardcoded_termux_policy=fail` is the product-clean default. Development
archives may use `audit` only when the emitted audit file is treated as a real
blocker list, not as compatibility debt to ignore.

`text_rewrites` counts safe text/symlink prefix rewrites. `binary_rewrites`
counts fixed-width ELF-safe rewrites: known compiled provider paths (including
dash/elvish-style `usr/lib`, `usr/bin`, `RfPATH`, and `usr/bin/sh` targets, with
C-string boundary rules so `usr/lib/...` paths are not corrupted), plus a
same-width blanket swap of `/data/data/com.termux/files/usr` to
`/data/data/zide.embed/files/usr` for any remaining compiled occurrences.
`hardcoded_termux_hits` lists paths that still embed `/data/data/com.termux/files/usr`
after those passes (for example symlink targets or non-binary payloads that were
not rewritten).
`runtime_support_files` lists app-owned files the consumer must materialize
outside the `usr/` archive root for those known binary rewrites.
`runtime_support_links` lists `source=>target` symlinks that let shortened
runtime paths point back at files staged from the archive (including the
`/data/data/zide.embed/files/usr` bridge to the real `metadata.prefix` root).

## iOS Initial Kinds

- `ios-bundle-manifest`

Meaning:

- a placeholder contract for future iOS-safe artifact bundles
- not an executable userland promise

## Product candidate channel (MP-A6)

Product candidate materialization uses the **same** artifact kinds as dev
(`android-prefix-archive`, etc.) but must use **`hardcoded_termux_policy=fail`**
on the prefix artifact and hit counts consistent with the MP-A6 gate. Dev
snapshot prereleases that use **`audit`** are not interchangeable with product
claims. See `ANDROID_PRODUCT_PROVIDER_DECISION.md` and
`docs/product-candidate/README.md`.

## Consumer Rule

Consumers should treat manifests as immutable pinned inputs.

Normative **`zide` runtime** obligations (staging, forbidden parsers, metadata
fields used as version/compatibility truth) live in
`ZIDE_MOBILE_ARTIFACT_CONSUMER.md`.

`zide` should consume:

- a manifest URL or local path
- artifact hashes
- artifact version/stamp
- platform compatibility metadata
- provider provenance metadata

`zide` should not consume:

- package recipe internals
- temporary cache layout
- provider internals or repository assumptions
- host-side implementation details
