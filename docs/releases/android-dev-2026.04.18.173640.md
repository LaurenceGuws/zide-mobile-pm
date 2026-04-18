# android-dev-2026.04.18.173640

Automated Android **development** snapshot prerelease (not a product release).

- GitHub: https://github.com/LaurenceGuws/zide-mobile-pm/releases/tag/android-dev-2026.04.18.173640
- Consumer manifest URL:
  https://github.com/LaurenceGuws/zide-mobile-pm/releases/download/android-dev-2026.04.18.173640/android-dev-prefix.release.manifest.json

**MP-A9 / APX-B18:** Prefix manifests no longer advertise
`/data/data/zide.embed/files/usr` in `runtime_support_links` (not materializable
in the app sandbox). The dev usr-root bridge is
`/data/data/uk.laurencegouws.zide/.z=>/data/data/uk.laurencegouws.zide/files/usr`,
with a variable-length binary rewrite from `/data/data/com.termux/files/usr` to
that `.z` path.

`zide-pm` defaults to the consumer manifest URL above.
