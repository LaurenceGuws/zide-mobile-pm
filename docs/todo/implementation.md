# Implementation Queue

This is the active queue for `zide-mobile-pm`.

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
- default roots are `bash,neovim,htop,gotop`
- dependency closure is resolved from the package index
- generated manifest records package index URL/checksum plus each selected
  package URL/version/size/checksum
- generated output validates with `zide-mobile-pm validate`

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
- archive checksum and size are emitted into
  `dist/android-dev-prefix.manifest.json`
- audit metadata is emitted into `dist/zide-android-dev-prefix.audit.json`
- default hardcoded-prefix policy is `fail`; current dev archive generation
  must opt into `-hardcoded-policy audit`

Boundary:

- the current upstream package set still contains compiled-in `com.termux`
  prefix hits
- audit mode is acceptable for development artifact production only
- product archive work must either remove those hits, own a forked package feed,
  or document a narrower approved runtime surface

## Current Priority

### MP-A4 Android Product Provider Decision

Decide whether Android product prefixes come from the current `termux-main`
provider, from a controlled mirror/fork, or from a Zide-owned Android provider.

Acceptance:

- Bash startup path is explicit
- compiled-in `com.termux` assumptions are removed or deliberately owned
- docs name which provider is authoritative for dev and product channels
- `android-prefix-archive` can run with `-hardcoded-policy fail` for the chosen
  product candidate

## Next Tickets

### MP-A3 Zide Consumer Contract

Acceptance:

- `zide` does not parse package internals
- `zide` can stage a produced artifact by manifest path
- version stamp and compatibility metadata are explicit

### MP-I1 iOS Artifact Research

Open only when there is a concrete iOS consumer.

Acceptance:

- define what iOS can legally execute/load
- define artifact kinds for iOS without copying Android package assumptions
