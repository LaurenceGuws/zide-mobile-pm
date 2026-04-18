# Product-candidate package roots (MP-A6)

## Passing fail-policy set (today)

Single Termux root package **`dash`** (closure is one `.deb` for aarch64 in the
current index) produces a prefix archive with **`hardcoded_termux_hits=0`** under
`-hardcoded-policy fail`, after fixed-width binary rewrites for dash’s embedded
Termux paths.

Generate the pinned input manifest:

```bash
go run ./cmd/zide-pm-admin android-dev-manifest \
  -out dist/product-candidate/android-input.manifest.json \
  -packages dash
```

Then:

```bash
go run ./cmd/zide-pm-admin android-product-candidate-probe \
  -manifest dist/product-candidate/android-input.manifest.json
# expect: mp_a6_product_candidate=pass
```

The full dev baseline (`bash,neovim,git,...`) remains **audit-only** until
remaining compiled-in paths are removed or owned; fail-policy hit counts drop
with shared rewrites but are not zero yet.

## Runtime symlinks

Archives built after this rewrite set advertise extra `runtime_support_links`
entries under `/data/data/uk.laurencegouws.zide/` (`ul`, `ub`, `b`, `u/bsh`).
See `ZIDE_MOBILE_ARTIFACT_CONSUMER.md` and emitted manifest metadata.
