# Zide PM CLI

Purpose: define the product-facing package command intended to run inside Zide
mobile shells.

## Product Role

`zide-pm` is the Zide-owned package UX for mobile products.

It is not a replacement for every low-level package tool. It is the stable
front door that Zide users should see first.

The backend may use:

- provider manifests
- prefix archives
- package indexes
- future Zide-owned feeds

The CLI surface should stay Zide-shaped:

- clear package/group names
- honest provider provenance
- explicit install state
- predictable recovery output

## MVP

The first MVP is deliberately narrow.

Commands:

- `zide-pm doctor`
- `zide-pm list-available`
- `zide-pm install dev-baseline --prefix <path>`

Initial bootstrap profile:

- `dev-baseline`

Current `dev-baseline` means:

- Bash
- Neovim
- htop
- gotop
- their pinned provider dependency closure

This is current bringup naming, not long-term product semantics.

The intended product direction is:

- APK/app releases can point at a compatible bootstrap or recommended profile
- installed packages later evolve under `zide-pm`
- default shell may initially be Bash, but shell choice should remain
  configurable userland policy rather than APK identity
- future naming should move from `dev-baseline` toward product language such as
  `recommended`, `terminal-baseline`, or similar once the onboarding/install
  model is mature enough

## Boundary

`zide-pm` consumes manifests and prefix archives. It does not parse `.deb`
payloads in the product path.

`zide-pm-admin` commands may still snapshot providers and materialize archives.
That work is backend/package-authority infrastructure, not product CLI surface.

The user-facing CLI should not expose Termux as the product. It can report
`provider=termux-main` as provenance while keeping the command model Zide-owned.

## Android First

The Android MVP installs an `android-prefix-archive` rooted at `usr/` into the
requested prefix.

That archive should currently be read as a bootstrap/recommended profile
artifact, not as a forever lock on package versions for the life of the APK.

For the current app:

```bash
zide-pm install dev-baseline --prefix /data/data/dev.zide.terminal/files/usr
```

The default manifest points at the current Android dev snapshot prerelease. Local
manifests are supported for development:

```bash
zide-pm install dev-baseline \
  --manifest ./android-dev-prefix.release.manifest.json \
  --prefix ./tmp/usr
```

Private GitHub release URLs use `ZIDE_PM_GITHUB_TOKEN`, `GITHUB_TOKEN`,
`GH_TOKEN`, or `gh auth token` when available.

Dev prefix archives include:

- `usr/bin/zide-pm`
- `usr/.zide-pm-install.json`

That install stamp lets the on-device shell run:

```bash
zide-pm doctor
zide-pm list-available
```

without requiring private GitHub release access from the phone. If manifest
loading fails, those read the installed state from `$PREFIX`.

## Not Yet

Do not claim these are done:

- arbitrary package install
- dependency mutation on-device
- upgrade/remove semantics
- product-clean provider policy
- iOS execution/install behavior
- stable long-term product naming for bootstrap/recommended profiles

Those need separate tickets after the first CLI shape is validated in the Zide
Android shell.
