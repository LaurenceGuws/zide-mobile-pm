# android-dev-2026.04.18.175422

Automated Android **development** snapshot prerelease (not a product release).

- GitHub: https://github.com/LaurenceGuws/zide-mobile-pm/releases/tag/android-dev-2026.04.18.175422
- Consumer manifest URL:
  https://github.com/LaurenceGuws/zide-mobile-pm/releases/download/android-dev-2026.04.18.175422/android-dev-prefix.release.manifest.json

**MP-A10 / APX-B18:** `usr/bin/zide-pm` now bootstraps Android DNS for HTTPS
manifest and artifact fetches (`getprop net.dns*` with `8.8.8.8` fallback) so
in-prefix runs are not pinned to broken loopback resolvers in `resolv.conf`.

`zide-pm` defaults to the consumer manifest URL above.
