# android-dev-2026.04.18.162659

**Note (APX-B18 / MP-A9):** this snapshot advertised a
`/data/data/zide.embed/files/usr` `runtime_support_links` bridge that the Android
app sandbox cannot materialize. Later dev snapshots replace it with
`/data/data/uk.laurencegouws.zide/.z` and a variable-length binary rewrite.

Automated Android **development** snapshot prerelease (not a product release).

- GitHub: https://github.com/LaurenceGuws/zide-mobile-pm/releases/tag/android-dev-2026.04.18.162659
- Consumer manifest URL:
  https://github.com/LaurenceGuws/zide-mobile-pm/releases/download/android-dev-2026.04.18.162659/android-dev-prefix.release.manifest.json

Prefix archives add a same-width compiled-in bridge
`/data/data/zide.embed/files/usr` (rewritten from `/data/data/com.termux/files/usr`
in ELF payloads) plus a matching `runtime_support_links` entry to
`/data/data/uk.laurencegouws.zide/files/usr`. Audit-mode builds report
**hardcoded_termux_hits=0** for this package closure.

`zide-pm` defaults to the consumer manifest URL above.
