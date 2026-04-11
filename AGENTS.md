# Agent Workflow (zide-mobile-pm)

This repo is artifact-authority infrastructure for Zide mobile products.

Default workflow:

1. Read `README.md`.
2. Read `docs/INDEX.md`.
3. Read the relevant architecture doc under `app_architecture/`.
4. Keep Android and iOS mechanics separate unless a shared artifact-contract
   concept is explicitly proven.
5. Update docs with any behavior or boundary change.
6. Run local validation before reporting completion:
   - `go test ./...`
   - `go run ./cmd/zide-mobile-pm validate examples/android-dev.manifest.json`

Rules:

- Do not put Zide app runtime code here.
- Do not put Android UI/lifecycle code here.
- Do not claim unmodified `com.termux` package payloads are product-correct for
  `dev.zide.terminal`.
- Do not invent a new mobile package manager when a pinned artifact contract is
  enough.
- Do not pretend iOS and Android have the same execution model.
- Generated artifacts belong under `dist/` and are not committed by default.
- Download/cache material belongs under `.cache/` and is not committed.

