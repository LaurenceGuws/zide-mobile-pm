# Product candidate (MP-A6 execution path)

This directory documents the **operator workflow** for Android prefix artifacts
that must pass **`hardcoded_termux_policy=fail`**. It is **not** the same as
the dev snapshot prerelease lane (`android-dev-snapshot-release`), which uses
**audit** policy.

## Layout (generated under `dist/`)

`dist/` is gitignored. Successful materialization writes:

| File | Purpose |
|------|---------|
| `dist/product-candidate/zide-android-prefix.tar.gz` | Prefix archive (`usr/` root) |
| `dist/product-candidate/android-prefix.manifest.json` | `android-prefix-archive` manifest document |
| `dist/product-candidate/prefix.audit.json` | Extraction / rewrite / hit audit |

## Commands (`zide-pm-admin`)

1. **Probe** (disposable temp tarball; persistent fail audit default):

   ```bash
   go run ./cmd/zide-pm-admin android-product-candidate-probe \
     -manifest dist/android-dev.manifest.json
   ```

2. **Materialize** (same fail policy; writes `dist/product-candidate/*` on success):

   ```bash
   go run ./cmd/zide-pm-admin android-product-candidate-materialize \
     -manifest dist/android-dev.manifest.json
   ```

3. **Ad-hoc** (full control): `android-prefix-archive -hardcoded-policy fail ...`

Authority: `app_architecture/ANDROID_PRODUCT_PROVIDER_DECISION.md`.

Consumer rules for emitted manifests: `app_architecture/ZIDE_MOBILE_ARTIFACT_CONSUMER.md`.
