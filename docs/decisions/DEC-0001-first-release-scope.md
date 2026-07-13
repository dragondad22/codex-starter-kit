# DEC-0001 — First-release scope

**Status:** Accepted; regulated-project scope amended 2026-07-12
**Owner:** dragondad22  
**Date:** 2026-07-11  
**Source decision:** D1

## Context

The system must be useful beyond a single application shape or expert developer. Narrowing
launch scope to unregulated greenfield apps would undermine retrofit and policy design.

## Original decision

The first release supports applications, libraries, infrastructure, data projects, and
documentation repositories. It is solo-first but team-scalable, uses GitHub as the
required initial collaboration platform, and genuinely supports regulated projects.

The last clause is retained as decision history but is superseded by the amendment below.
It no longer establishes a first-release claim of comprehensive regulatory support or
verified handling of highly sensitive content.

## Decision

The project-type, solo/team, and GitHub scope remains accepted. For sensitive or regulated
contexts, the first release provides truthful detection and triggered governance, not a
general assurance claim. It records one project-level declaration—whether the project
intentionally contains or processes information requiring special confidentiality,
privacy, contractual, or regulatory handling—with values `No`, `Yes`, or `Unsure`.

`Yes` and `Unsure` require a concise data-handling notice, explicit acknowledgment, and
coverage states that distinguish content classification, handling authorization, and
product assurance. Acknowledgment records that the user saw the limitation; it does not
establish conformance, authorize a tool or transmission, or override law, contract,
organization policy, or qualified review.

Universal secret protection, least authority, no silent transmission or tool activation,
truthful unknown states, and deterministic lifecycle operation remain required. When a
verified sensitive-data route is absent, safe classification, planning, and remediation
may continue without exposing the content, but affected assurance determinations report
`needs-review` and operations requiring that route report `unsupported`. Detailed data
taxonomy, egress enforcement, provider/environment assurance, and verified regulatory
pack coverage are Later capabilities tracked by issue #21.

## Consequences

Project classification and policy compilation are core modules. The architecture cannot
assume a deployable application, self-approval, or absence of regulation. GitHub is
mandatory at launch; other tracker adapters are out of scope. A project may be managed
while sensitive-data assurance remains `needs-review` or an affected operation is
`unsupported`; neither state may be summarized as regulatory conformance.

## Source

[Discovery decision D1](../discovery/CODEX_STARTER_KIT_REVIEW.md#d1) and the first-release
operating contract recorded there. The 2026-07-12 amendment was approved through
[issue #20](https://github.com/dragondad22/codex-starter-kit/issues/20).
