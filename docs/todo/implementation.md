# Implementation Queue

This is the active queue for the `zide-mobile-pm` repo.

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
  `dev.zide.terminal` runtime contract
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
    `run-as dev.zide.terminal`
  - `zide-pm list-available --manifest /no/such/manifest --prefix $PREFIX`
    prints `dev-baseline`

Boundary:

- on-device `install dev-baseline` is not claimed yet
- this proves the product CLI exists in the shell and can report installed
  package state

## Current Priority

### MP-A6 Android Product Provider Decision

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

Acceptance:

- `zide` does not parse package internals
- `zide` can stage a produced artifact by manifest path
- version stamp and compatibility metadata are explicit

### MP-I1 iOS Artifact Research

Open only when there is a concrete iOS consumer.

Acceptance:

- define what iOS can legally execute/load
- define artifact kinds for iOS without copying Android package assumptions
