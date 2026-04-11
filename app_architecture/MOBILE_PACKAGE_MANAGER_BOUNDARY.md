# Mobile Package Manager Boundary

Purpose: define why `zide-mobile-pm` exists and where it stops.

## Product Role

`zide-mobile-pm` is the mobile package authority and repo identity for the Zide
project family.

It is used by Zide, but it should remain portable enough to support other
Zide-family mobile consumers that need the same artifact discipline.

## Ownership Split

`zide-mobile-pm` owns:

- the `zide-pm` package CLI
- the `zide-pm-admin` backend/admin tool
- artifact manifests
- provider metadata and trust boundaries
- package/bootstrap metadata
- package index snapshots
- checksums
- artifact cache layout
- host-side materialization tools
- release/channel policy for mobile artifacts

`zide` owns:

- runtime integration
- terminal/editor UX
- Android lifecycle/input/insets
- native rendering
- PTY/session semantics
- user-facing bootstrap progress and errors

## Android First

Android currently needs:

- a Bash-capable app-private prefix
- Neovim and terminal-dev tools for implementation testing
- package metadata that is not silently tied to `com.termux`
- exact artifact versions and checksums

Android may use:

- providers such as `termux-main`
- prefix archives
- package index snapshots
- package recipes or binary payloads
- SDK 28 execution posture for terminal userland compatibility

Android must not claim:

- unmodified `com.termux` package payload roots are product-correct for
  `dev.zide.terminal`
- host-side relocation hacks are a final package-manager contract

## Providers

Providers are build-time artifact sources, not Zide runtime package managers.

The current Android provider is `termux-main`. It is allowed as a development
and bootstrap source because its package index and payloads can be pinned and
audited. It is not automatically a product provider for `dev.zide.terminal`.

Provider outputs must become Zide artifact contracts before Zide consumes them.
That keeps Android-specific mechanics out of the Zide runtime and leaves room
for a future Zide-owned Android feed or completely different iOS artifact
sources.

## Package CLI

`zide-pm` is the product-facing command intended for Zide mobile shells.

It consumes artifact contracts and reports provider provenance. It must not make
Termux, apt, or any future provider look like the Zide product surface.

`zide-pm-admin` is not a product shell command. It is host-side/admin tooling
for validation, provider snapshotting, archive generation, and snapshot
publishing.

The current MVP is `dev-baseline` only. Arbitrary package install, upgrade, and
remove semantics are later work.

## iOS Later

iOS should be treated as a separate execution model.

Likely iOS artifact types may include:

- bundled read-only tools
- editor/IDE resource bundles
- LSP or syntax asset bundles
- platform-approved helper artifacts

iOS should not inherit Android assumptions such as:

- apt/dpkg as a baseline
- writable executable userland prefixes
- arbitrary downloaded binary execution

The shared boundary is the manifest and trust contract, not the package
mechanics.

## Stop Line

If code needs to run inside the Zide app at runtime, it probably does not belong
here unless it is generated artifact metadata.

If code builds, verifies, signs, snapshots, or publishes mobile artifacts, it
belongs here unless it is still a short-lived probe.
