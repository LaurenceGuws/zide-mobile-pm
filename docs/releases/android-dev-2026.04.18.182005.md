# android-dev-2026.04.18.182005

Automated Android **development** snapshot prerelease (not a product release).

- GitHub: https://github.com/LaurenceGuws/zide-mobile-pm/releases/tag/android-dev-2026.04.18.182005
- Consumer manifest URL:
  https://github.com/LaurenceGuws/zide-mobile-pm/releases/download/android-dev-2026.04.18.182005/android-dev-prefix.release.manifest.json

**MP-A11 hotfix:** removed unsafe variable-length blanket binary rewrite in
`androidprefix` so ELF payloads are no longer structurally corrupted. This
restores runnable shell binaries (`usr/bin/bash`) for Android staging while
keeping known fixed-width binary rewrites.

`zide-pm` defaults to the consumer manifest URL above.
