# Agent Workflow (zide-mobile-pm)

This repo is the mobile package authority for Zide-family products.

Naming and ownership are strict:

- `zide-pm` is the product-facing package CLI.
- `zide-pm-admin` is the backend/admin tool.
- `zide-mobile-pm` is the repo/module/project identity.
- Android dev publishing is a snapshot prerelease lane, not an official product
  release lane.
- Do not blur product CLI semantics with provider/admin tooling.

Default workflow:

1. Read `README.md`.
2. Read `docs/INDEX.md`.
3. Read the relevant architecture doc under `app_architecture/`.
4. Keep Android and iOS mechanics separate unless a shared artifact-contract
   concept is explicitly proven.
5. Update docs with any behavior or boundary change.
6. Run local validation before reporting completion:
   - `go test ./...`
   - `go run ./cmd/zide-pm-admin validate examples/android-dev.manifest.json`

Rules:

- Do not put Zide app runtime code here.
- Do not put Android UI/lifecycle code here.
- Do not claim unmodified `com.termux` package payloads are product-correct for
  `uk.laurencegouws.zide`.
- Do not treat provider internals as the Zide product surface.
- Do not describe dev snapshot prereleases as product/official releases.
- Do not let docs or commands imply that `zide-pm-admin` is the user-facing
  shell command; that role belongs to `zide-pm`.
- Do not pretend iOS and Android have the same execution model.
- Generated artifacts belong under `dist/` and are not committed by default.
- Download/cache material belongs under `.cache/` and is not committed.
