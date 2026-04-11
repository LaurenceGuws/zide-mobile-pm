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

`hardcoded_termux_policy=fail` is the product-clean default. Development
archives may use `audit` only when the emitted audit file is treated as a real
blocker list, not as compatibility debt to ignore.

## iOS Initial Kinds

- `ios-bundle-manifest`

Meaning:

- a placeholder contract for future iOS-safe artifact bundles
- not an executable userland promise

## Consumer Rule

Consumers should treat manifests as immutable pinned inputs.

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
