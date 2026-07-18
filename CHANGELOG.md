# Changelog

All notable Codex Starter Kit changes are generated from structured change records.

<!-- source-digest: sha256:b38de120c5cd4c1b33d6d4fe7ba41be8709185aa17d1d9ad5d845cbc600230a5 -->

## [Unreleased]

### Added
- Create managed repositories through reviewable, immutable plans and an evidence-backed apply operation. (#26)
- Verify seed controls without converting fail, not-applicable, not-configured, needs-review, or accepted-exception states into a pass. (#27)
- Qualify equivalent Phase 1 lifecycle semantics on native Linux, macOS, and Windows runners. (#30)
- Add an installable Codex plugin status tracer that fails closed when engine compatibility or provenance is insufficient. (#51)
- Guide managed-repository creation through separate input, notice, plan, effect-approval, and recovery steps in the Codex plugin. (#52)
- Guide truthful verification through immutable plans, explicit evidence effects, redaction, and preserved control states in the Codex plugin. (#53)
- Manage and verify one task deterministically through a credential-free lifecycle request and in-memory adapter. (#71)
- Reconcile one managed task through a native, identity-bound GitHub REST and GraphQL adapter with explicit recovery evidence. (#72)
- Bootstrap and verify an isolated organization-owned GitHub contract sandbox through an approved, replay-safe lifecycle. (#73)
- Track one product version and generate audience-aware changelogs from validated, durable change records. (#78)
- Prompt owners to capture durable untracked work and decisions before material implementation. (#84)

### Changed
- Make managed-repository creation safe to replay, interrupt, recover, and reconcile without deleting ambiguous user content. (#29)
- Qualify plugin routing, capability modes, fallback behavior, and native development installation while preserving unpublished-engine limitations. (#54)

### Fixed
- Reconcile closed items, parents, and direct dependents through immutable lifecycle plans with before-and-after receipts. (#15)
