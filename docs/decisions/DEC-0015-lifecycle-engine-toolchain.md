# DEC-0015 — Go lifecycle-engine toolchain

**Status:** Accepted
**Owner:** dragondad22
**Date:** 2026-07-12
**Source decision:** D15

## Context

The lifecycle engine must be a deterministic standalone authority on native Linux,
macOS, and Windows. Users should not need Codex, a compatibility shell, or an installed
language runtime. Contributors need a tractable build, test, dependency, provenance,
offline, signing, compatibility, and migration contract before the Phase 1 schemas and
engine implementation begin.

Go, Rust, and Python were compared against the accepted product contract. Rust offers a
stronger compile-time memory model but adds implementation and toolchain complexity before
Phase 1 has material unsafe-memory or high-concurrency requirements. Python remains useful
for foundation validation but its standard single-file distribution still requires a
suitable interpreter; freezing would add a separate packaging/provenance system.

## Decision

Implement the standalone lifecycle engine and CLI in Go, initially pinned to Go 1.26.5
for contributor and CI builds. Build and test separate native `starter-kit` binaries on
Linux, macOS, and Windows; cross-compilation is not native behavior evidence.

Keep Phase 1 standard-library-only. Any third-party module requires an issue-backed trust,
authority, data, license, provenance, compatibility, cost, and offline review. Use JSON
for versioned machine operations/state, Markdown for human views, and structured process
arguments for Git. Go packages and types are implementation details: the durable public
contract is the language-neutral lifecycle operation, schema, evidence, and fixture seam.

Release binaries with SHA-256 manifests and GitHub artifact attestations. Retain
attestation bundles and trusted-root material for offline verification. Platform-native
code signing, notarization, and reputation are separate release-adapter obligations and
must not be implied by provenance attestations.

## Consequences

Released users run a native binary without installing Go. Contributors and CI need the
pinned toolchain, while the standard-library-only boundary avoids module downloads after
that toolchain is available. Native filesystem, locking, atomic replacement, process, and
Git behavior still require adversarial tests; the language choice is not itself a safety
control.

Rust is the preferred reimplementation fallback. Return this decision to review if native
semantics, approved platform targets, measurable size/startup/memory/performance, embedding,
cryptographic/FIPS, signing, sandbox, or distribution requirements cannot be met credibly,
or if evidence shows another option materially reduces total risk or maintenance cost.

The decision selects no exact CLI syntax, package layout, schema, installer, minimum OS,
release target, dynamic plugin model, or external-adapter dependency. Those remain owned
by the applicable vertical slices and release decision.

## Source

[Discovery decision D15](../discovery/CODEX_STARTER_KIT_REVIEW.md#d15), approved through
[issue #25](https://github.com/dragondad22/codex-starter-kit/issues/25), with the full
[toolchain evaluation](../research/ENGINE_TOOLCHAIN_EVALUATION.md).
