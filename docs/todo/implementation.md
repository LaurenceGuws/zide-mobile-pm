# Implementation Queue

This is the active queue for the `zide-mobile-pm` repo.

## Reporting Contract

Required report headers:

- `LABELS`
- `#DONE`
- `#OUTSTANDING`
- `COMMITS`
- `VALIDATION`
- `Blocked by Archtect review needed: true|false`

Required `LABELS` fields:

- `Lane: zide_mobile_pm`
- `Batch: <MP-*>`
- `Gate: in_progress|super_gate`
- `Focus: <one-line>`
- `Blockers: none|<summary>`

Label meanings:

- `Planned`: queued/in-progress scope not yet validated
- `Confirmed`: implemented and validated in this run
- `Deferred`: explicitly moved out of this batch
- `Blocked`: cannot proceed without external decision/dependency
- `ReviewRequired`: super-gate reached, architect verdict pending
- `Accepted`: architect reviewed and approved
- `Rejected`: architect reviewed and not approved

## Completed

### MP-A0 Provider Model — met

Result:

- providers are defined as build-time artifact sources
- `termux-main` is the initial Android dev/bootstrap provider
- provider outputs must become Zide artifact contracts before Zide consumes
  them
- docs forbid treating provider internals as Zide runtime package-manager
  behavior

### MP-A1 Android Dev Artifact Manifest — met, first cut

Result:

- `android-dev-manifest` fetches/caches the Termux main aarch64 package index
  under `.cache/`
- generated artifacts record provider provenance metadata
- cache checksum is written and verified on reuse
- default roots are `bash,neovim,git,ripgrep,htop,gotop`
- dependency closure is resolved from the package index
- generated manifest records package index URL/checksum plus each selected
  package URL/version/size/checksum
- generated output validates with `zide-pm-admin validate`

Boundary:

- this is a dev artifact channel
- `android-termux-deb` artifacts are pinned source payloads, not a final
  `uk.laurencegouws.zide` runtime contract
- product prefix materialization stays in MP-A2

Known limitation:

- `btop` is not currently present in the Termux main aarch64 package index

### MP-A2 Android Prefix Archive Producer — met, dev audit mode

Result:

- `android-prefix-archive` consumes the MP-A1 manifest
- pinned `.deb` payloads are downloaded, size-checked, and SHA-256 verified
- only `data/data/com.termux/files/usr/*` payload paths are extracted
- output archive is rooted at `usr/` for staging under Android app files
- text files and symlink targets that contain the old Termux prefix are
  rewritten to the Zide app prefix where safe
- known compiled Bash/htop provider paths are rewritten as fixed-width
  C-strings and counted separately as binary rewrites
- runtime support files required by those binary rewrites are advertised in
  archive metadata for the app consumer to materialize
- runtime support symlinks are advertised so shortened binary paths can point
  back at archived prefix files without making Zide parse provider internals
- archive checksum and size are emitted into
  `dist/android-dev-prefix.manifest.json`
- audit metadata is emitted into `dist/zide-android-dev-prefix.audit.json`
- default hardcoded-prefix policy is `fail`; current dev archive generation
  must opt into `-hardcoded-policy audit`

Boundary:

- the current upstream package set still contains compiled-in `com.termux`
  prefix hits outside the known rewritten Bash/htop paths
- audit mode is acceptable for development artifact production only
- product archive work must either remove those hits, own a forked package feed,
  or document a narrower approved runtime surface

### MP-A3 Android Dev Release Lane — met

Result:

- `android-dev-snapshot-release` automates the fast dev snapshot release lane
- release lane regenerates the Android dev manifest and prefix archive
- release manifest rewrites the archive URL to a release-local asset name
- dry-run mode prepares assets without publishing
- non-dry-run mode creates a tag and GitHub prerelease with generated assets

Boundary:

- dev snapshot releases are prereleases
- dev snapshot releases use `termux-main` provider with hardcoded-prefix audit
  mode
- product/official releases still require product-clean provider policy

### MP-A4 zide-pm MVP — met

Result:

- `zide-pm` exists as the user-facing package CLI
- commands:
  - `doctor`
  - `list-available`
  - `install dev-baseline --prefix <path>`
- the MVP consumes the Android prefix manifest/archive contract
- it does not parse `.deb` payloads or provider package internals
- install writes `.zide-pm-install.json` state into the target prefix

Boundary:

- only `dev-baseline` is supported
- `dev-baseline` is current MVP/bootstrap naming, not settled product naming
- arbitrary package install/remove/upgrade is not implemented
- arbitrary package install/remove/upgrade is not implemented

### MP-A8 Android host-aware test-binary pull/install — met (foundation)

Result:

- artifact kind `android-test-binary` validates on Android manifests with
  `metadata.install_relative_path` and standard provider metadata
- `zide-pm` honors `ZIDE_PM_HOST_PLATFORM=android` for catalog visibility;
  without it, test-binary entries stay out of `list-available` so host runs do
  not imply Android-only install semantics
- `zide-pm install <android-test-binary name>` downloads the pinned payload,
  verifies size/hash, and writes under the prefix at `install_relative_path`
- `zide-pm doctor` prints `zide_pm_host_platform`
- iOS is not assigned this kind; platform/kind pairing stays explicit in
  validation

Boundary:

- no Java/Android UI or app lifecycle code in this repo
- product naming remains `zide-pm` / `zide-pm-admin` only

Follow-on:

- `android-dev-snapshot-release` now materializes `assets/zide-android-catalog-smoke.sh`,
  appends artifact `zide-android-catalog-smoke` (`android-test-binary`) to the
  published `android-dev-prefix.release.manifest.json`, and uploads the payload
  as `zide-android-catalog-smoke.sh` on the same GitHub prerelease.

### MP-A5 Android zide-pm Staging — met

Result:

- Android-compatible `zide-pm` binary is produced during dev snapshot release
  generation
- dev prefix archives include `usr/bin/zide-pm`
- dev prefix archives include `usr/.zide-pm-install.json` so `doctor` and
  `list-available` can work without private GitHub manifest access on-device
- Note10 validation proves:
  - `zide-pm help` runs as an Android binary
  - artifact-staged app prefix includes `zide-pm`
  - `zide-pm doctor --manifest /no/such/manifest --prefix $PREFIX` works under
    `run-as uk.laurencegouws.zide`
  - `zide-pm list-available --manifest /no/such/manifest --prefix $PREFIX`
    prints `dev-baseline`

Boundary:

- on-device `install dev-baseline` is not claimed yet
- this proves the product CLI exists in the shell and can report installed
  package state

## Current Priority

### MP-A6 Android Product Provider Decision

Interim dev channel authority: `termux-main` via `android-dev-snapshot-release`
(audit policy) remains the pinned bootstrap path until this ticket selects the
product provider. Latest published dev snapshot: `android-dev-2026.04.18.140021`
(see `docs/releases/android-dev-2026.04.18.140021.md`).

Groundwork (**MP-A6-doc**):

- `app_architecture/ANDROID_PRODUCT_PROVIDER_DECISION.md` defines dev vs product
  channel vocabulary, explicit Bash startup path expectations, `com.termux`
  hit / policy meaning, and the three provider options under consideration.

Executable gate (**MP-A6-probe**):

- `zide-pm-admin android-product-candidate-probe` runs `android-prefix-archive`
  with **`hardcoded-policy=fail`** and records audit output (default
  `dist/mp-a6-product-candidate.audit.json` on failure) for MP-A6 evidence.

Materialize path (**MP-A6-materialize**):

- `zide-pm-admin android-product-candidate-materialize` runs the same fail-policy
  archive build and writes **`dist/product-candidate/*`** on success (operator
  doc: `docs/product-candidate/README.md`).

Concrete input (**MP-A6-dash-minimal**):

- Termux package root **`dash`** yields a fail-policy-clean archive today (see
  `docs/product-candidate/PACKAGES.md`). Full dev tool closure still fails fail
  policy until additional packages are narrowed or rewritten.

## Next Tickets

Decide whether Android product prefixes come from the current `termux-main`
provider, from a controlled mirror/fork, or from a Zide-owned Android provider.

Acceptance:

- Bash startup path is explicit
- compiled-in `com.termux` assumptions are removed or deliberately owned
- docs name which provider is authoritative for dev and product channels
- `android-prefix-archive` can run with `-hardcoded-policy fail` for the chosen
  product candidate

### MP-A7 Zide Consumer Contract

Groundwork (**MP-A7-doc**):

- `app_architecture/ZIDE_MOBILE_ARTIFACT_CONSUMER.md` states allowed manifest
  inputs, required staging/verification behavior, forbidden package-internal
  parsers, and which metadata fields constitute version/compatibility truth for
  the Zig/runtime consumer (including top-level manifest document fields aligned
  with emitted JSON).

Acceptance:

- `zide` does not parse package internals
- `zide` can stage a produced artifact by manifest path
- version stamp and compatibility metadata are explicit

### MP-I1 iOS Artifact Research

Open only when there is a concrete iOS consumer.

Acceptance:

- define what iOS can legally execute/load
- define artifact kinds for iOS without copying Android package assumptions
