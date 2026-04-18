# Android product provider decision (MP-A6)

Purpose: name the **decision space** and **interim authority** for Android
prefix artifacts so product work does not silently inherit `termux-main`
semantics.

This document is **groundwork** for closing MP-A6. It does not pick the final
product provider by itself; Architect/product still choose among the options
below.

## Channels (single vocabulary)

| Channel | Authoritative generator (this repo) | Provider id (metadata) | Typical `provider_role` | `hardcoded_termux_policy` |
|--------|--------------------------------------|-------------------------|-------------------------|---------------------------|
| **Dev snapshot** | `zide-pm-admin android-dev-snapshot-release` | `termux-main` | `android-dev-bootstrap` | **`audit`** (mandatory today) |
| **Product** (pending) | TBD: same toolchain with stricter inputs, or a new lane | TBD | TBD (not `android-dev-bootstrap` for a product channel) | **`fail`** at minimum for the chosen candidate |

Dev snapshots are **GitHub prereleases**. They are not product releases.

## Bash startup path (explicit)

For an `android-prefix-archive` with `metadata.archive_root=usr` staged under
the app-private prefix directory `$PREFIX` (the on-disk layout matches the
archive’s `usr/` tree):

- **Interactive shell:** `$PREFIX/bin/bash` is the supported baseline entry for
  terminal sessions (Termux-compatible SDK posture as recorded in
  `metadata.target_sdk`).
- **Startup files:** rely on relocated `etc/profile`, `etc/bash.bashrc`, and
  related paths **inside the staged prefix**, not on host paths.
- **Binary rewrites / support paths:** when `metadata.runtime_support_files` or
  `metadata.runtime_support_links` are non-empty, the **app consumer** must
  materialize those paths exactly as specified. That is part of honoring the
  artifact contract—not package-manager internals.

If a future product channel narrows the runtime surface, the manifest must still
state an explicit shell path policy compatible with the staged tree.

## Compiled-in `com.termux` assumptions

**Product-clean** means: `hardcoded_termux_policy=fail` on the
`android-prefix-archive` artifact **and** `hardcoded_termux_hits` consistent
with zero unknown compiled hits for the **approved** package set—or those hits
are **deliberately owned** (forked packages, controlled feed) and documented in
artifact `limitations` and audit output.

**Dev** may use `audit` only with the understanding that the audit file is a
**blocker list**, not ignorable debt.

Closing MP-A6 requires choosing one of:

1. **Narrowed `termux-main`:** keep provider id `termux-main` but change inputs
   (package set, patches, or mirror) until `android-prefix-archive` can ship
   with **`fail`** for the product candidate.
2. **Zide-owned Android feed:** new provider id (for example `zide-android-main`)
   with Zide-controlled payloads and metadata; same manifest kinds, different
   trust boundary.
3. **Controlled fork/mirror** of upstream: provider metadata names the mirror;
   payloads are still pinned by URL/hash in the manifest.

## Interim rule (until MP-A6 closes)

- **Dev consumer manifest** authority: latest `android-dev-*` prerelease from
  `android-dev-snapshot-release`, as referenced by `DefaultAndroidDevManifestURL`
  and `docs/releases/`.
- **Product consumer manifest**: **not published** from this repo until MP-A6
  acceptance criteria in `docs/todo/implementation.md` are met.

## Acceptance mapping

MP-A6 is **met** when:

- this decision is **executed** (one path above is in production for product),
  not only documented;
- `android-prefix-archive` for the product candidate validates with
  `-hardcoded-policy fail`;
- docs name **dev vs product** authoritative providers and roles without mixing
  audit artifacts into product claims.

## Executable evidence (product-candidate probe)

Run:

```bash
go run ./cmd/zide-pm-admin android-product-candidate-probe \
  -manifest dist/android-dev.manifest.json
```

This command builds prefix archive inputs with **`hardcoded_termux_policy=fail`**
into disposable temp paths. **Exit code 0** means no compiled-in `com.termux`
hits remain under fail policy for that manifest (product-candidate clean for
this probe). **Non-zero** means hits remain; an audit JSON is written to
**`dist/mp-a6-product-candidate.audit.json`** by default (`-audit-out` overrides)
so the blocker list is inspectable without starting from audit-mode tarballs.

The default dev package set is expected to **fail** this probe until MP-A6
narrows inputs or changes provider payloads. That failure is **evidence**, not a
tooling defect.

## Materialize path (product-candidate outputs)

When the probe **would pass**, operators (or automation) can materialize the
same fail-policy build into **`dist/product-candidate/`** using:

```bash
go run ./cmd/zide-pm-admin android-product-candidate-materialize \
  -manifest dist/android-dev.manifest.json
```

This is **`android-prefix-archive`** with **`hardcoded-policy=fail`** and default
paths:

- `dist/product-candidate/zide-android-prefix.tar.gz`
- `dist/product-candidate/android-prefix.manifest.json`
- `dist/product-candidate/prefix.audit.json`

On failure, **`prefix.audit.json`** is still written (and the command exits
non-zero) so the materialize step matches probe evidence. Operator notes live in
`docs/product-candidate/README.md`.
