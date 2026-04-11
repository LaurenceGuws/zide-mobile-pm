# Docs Index

Use this repo's docs as the authority for mobile package/artifact production.

Naming rules:

- `zide-pm` means the product-facing shell/package command.
- `zide-pm-admin` means the backend/admin tool.
- `zide-mobile-pm` means the repo/module/project identity.
- Android dev publishing must be described as snapshot prereleases, not product
  releases.

Start here:

- `README.md` — project scope and current commands
- `app_architecture/MOBILE_PACKAGE_MANAGER_BOUNDARY.md` — repo boundary
- `app_architecture/ZIDE_PM_CLI.md` — user-facing mobile package CLI
- `app_architecture/PROVIDER_MODEL.md` — provider/source boundary
- `app_architecture/ARTIFACT_CONTRACT.md` — manifest and artifact contract
- `docs/todo/implementation.md` — active implementation queue

Doc scope:

- `README.md` is the public project summary.
- `app_architecture/` is current technical authority.
- `docs/todo/` is the execution queue.
- `docs/research/` is for evidence and investigations.
- `docs/releases/` is for release notes once artifacts are published.
