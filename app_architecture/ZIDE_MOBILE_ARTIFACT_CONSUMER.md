# Zide mobile artifact consumer contract (MP-A7)

Purpose: define what the **`zide` repository** (runtime) may rely on when
staging mobile artifacts, and what it must **not** implement.

`zide-mobile-pm` owns manifests and admin tooling; **`zide` owns integration**
(UX, lifecycle, staging orchestration). This contract is the seam between them.

## Inputs (allowed)

`zide` may consume, for Android:

- A **manifest** (`schema_version: 1`, `project: zide-mobile-pm`, `platform:
  android`) as a **local path** and/or **HTTPS URL** pinned by release or
  config.
- For bootstrap / prefix materialization: the single **`android-prefix-archive`**
  artifact selected by the same cardinality rule as `zide-pm` (exactly one per
  manifest for that workflow).
- **Checksums and sizes** on that artifact (`sha256`, `size`) for verification
  after download.
- **Version stamp:** `artifact.version` (string) plus `artifact.sha256` as the
  immutable identity of the staged bits. When present,
  `metadata.source_manifest_sha256` ties the prefix archive back to the
  generating dev manifest.
- **Compatibility metadata** on the prefix artifact, including at minimum:
  - `package_name` (Android application id the tree is shaped for)
  - `prefix` (intended on-device prefix path)
  - `target_sdk`
  - `archive_root` (expected `usr` for current dev archives)
  - `hardcoded_termux_policy` and numeric `hardcoded_termux_hits` for gating
  - `limitations` array on the artifact for human-visible constraints

### `android-prefix-archive` metadata emitted today (`zide-pm-admin`)

The following keys are written on **`android-prefix-archive`** artifacts produced
by `zide-pm-admin android-prefix-archive` / dev snapshot releases (authoritative
for MP-A7 alignment with real manifests):

| Key | Role |
|-----|------|
| `package_name` | Android app id the tree is shaped for |
| `prefix` | Intended on-device prefix root |
| `archive_root` | `usr` for current archives |
| `target_sdk` | Terminal/userland compatibility posture |
| `provider` | Source provider id (e.g. `termux-main`) |
| `provider_role` | Channel role (e.g. `android-dev-bootstrap`) |
| `provider_platform` | `android` |
| `provider_architecture` | e.g. `aarch64` |
| `source_manifest_sha256` | Hex sha256 of generating MP-A1-style manifest |
| `source_package_count` | Count of `android-termux-deb` inputs |
| `hardcoded_termux_hits` | Count of remaining compiled-in `com.termux` hits |
| `hardcoded_termux_policy` | `audit` or `fail` for this build |
| `text_rewrites` / `binary_rewrites` | Prefix rewrite tallies |
| `runtime_support_files` | Comma-separated app-owned paths to materialize |
| `runtime_support_links` | `source=>target` symlink specs for the consumer |
| `removed_termux_prefixed_binaries` | Pruned Termux-prefixed binaries count |
| `extracted_*` / `archive_*` | File/symlink counts from build vs archive |
| `zide_pm_cli` | Whether `usr/bin/zide-pm` was bundled (`included` when set) |

`zide` should treat unknown metadata keys as **opaque** unless this contract or
`ARTIFACT_CONTRACT.md` promotes them; do not infer provider package-manager
behavior from them.

`zide` may surface **`zide-pm`** inside the staged prefix for user-driven
install/catalog flows; the Zig/runtime layer should still treat the **manifest
+ prefix archive** as the authoritative bootstrap contract.

## Staging behavior (must)

- Download (or copy) the archive payload for the selected artifact using the
  manifest `url` (absolute or resolved relative to the manifest URL).
- **Verify** `size` and `sha256` before extraction.
- Extract **`usr/`** tree into the app-private prefix layout expected by
  `metadata.prefix` / deployment policy (same shape as `zide-pm` extract).
- Apply **runtime support** obligations: if `runtime_support_files` or
  `runtime_support_links` are set, materialize those paths **outside** the
  extracted `usr/` tree as specified. Do not reinterpret them through Termux or
  provider logic.

## Forbidden (must not)

- Parse **`.deb`**, **dpkg**, **apt**, or Termux **Packages** index formats in
  product paths.
- Treat `android-termux-deb` artifacts as direct runtime install units (those
  are **host-side production inputs** for `zide-pm-admin`).
- Infer cross-platform semantics (Android kinds on iOS or vice versa).
- Bypass checksum verification for downloaded artifacts.

## `android-test-binary` (optional consumer note)

Catalog-mode **`android-test-binary`** artifacts are consumed by **`zide-pm`**
when `ZIDE_PM_HOST_PLATFORM=android`. The Zig app does not need a second
installer: it runs `zide-pm` with that environment. Staging policy for those
files is identical to pinned URL + hash + install path semantics documented in
`ARTIFACT_CONTRACT.md`.

## MP-A7 acceptance

MP-A7 is **met** when `zide` implementation is audited against this document:
no package-internal parsers on product paths, staging is manifest-driven, and
version/compatibility fields above are honored or explicitly surfaced to the
operator on mismatch.
