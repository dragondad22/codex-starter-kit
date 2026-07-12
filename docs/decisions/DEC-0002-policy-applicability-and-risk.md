# DEC-0002 — Policy applicability and risk

**Status:** Accepted  
**Owner:** dragondad22  
**Date:** 2026-07-11  
**Source decision:** D2

## Context

Small projects must not be unsafe, but imposing every operational and regulatory control
on every repository creates noise and false confidence. Real work also needs an honest
way to govern controls that cannot immediately be satisfied.

## Decision

Every control independently declares applicability, evaluation, enforcement, and
exception policy. Universal trust controls always apply; context facts activate additional
policy. Risks use corrective exceptions, periodically reviewed residual risks, or
prohibited exceptions. Risk acceptance never changes an underlying result to pass.

## Consequences

The policy compiler operates on recorded facts and versioned rules. Self-approval cannot
satisfy independent review. Law, binding contracts, live-secret protection, truthful
coverage, and other prohibited controls cannot be waived by project preference.

## Source

[Discovery decision D2](../discovery/CODEX_STARTER_KIT_REVIEW.md#d2), including the approved
universal baseline and trigger families.
