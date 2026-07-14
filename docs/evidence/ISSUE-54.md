# Issue #54 — Phase 2 plugin qualification and handoff

**Date:** 2026-07-14
**Change owner:** dragondad22
**Issue:** [#54](https://github.com/dragondad22/codex-starter-kit/issues/54)

## Delivered contract

The Phase 2 plugin retains version `0.3.0` and its three focused skills. Qualification adds
a versioned machine capability model, explicit approval-boundary contract, deterministic
workflow/operational scenarios, bounded live routing evidence, compatibility report,
quality receipt, and one operational handoff. Qualification artifacts do not add runtime
authority or turn the plugin into the conformance engine.

The capability model covers `create`, `status`, and `verify` across `full`,
`degraded-guidance`, `verification-only`, and `unsupported`. It names the evidence,
diagnostics, remediation, fallback, and transitions for each mode and retains first-run
offline, restricted workspace, administrator-disabled plugin, missing/incompatible
engine, malformed output, interruption, cancellation, and recovery outcomes.

## Live interaction evidence

With product-owner authorization, Codex CLI `0.144.1` installed plugin `0.3.0` from the
local development marketplace into an isolated temporary `CODEX_HOME`. The home referenced
the existing authentication file only for the run. Eight ephemeral read-only routing
sessions used a temporary empty Git repository and prohibited tool calls, engine calls,
and lifecycle effects.

Explicit and implicit create, status, and verify requests selected exactly their focused
skills. Generic grocery-list creation and arithmetic verification selected no Starter Kit
skill. Every result reported zero extra reference loads, no planned engine invocation, and
no planned lifecycle effect. The calls consumed 112,920 input tokens, of which 13,056 were
cached, and 515 output tokens under the active account's normal model usage.

The first preflight request was rejected before model execution because Codex CLI's output
schema subset does not accept `uniqueItems`; the schema was narrowed and all eight planned
checks then passed. The rejected setup attempt recorded no usage or effect. The temporary
Codex home, plugin cache, authentication link, sessions, and repository were removed.

## Qualification boundary

The selected live surface is Codex CLI on Linux. Desktop Codex and the VS Code IDE remain
`needs-review`; the IDE also retains the official-documentation conflict recorded by
DEC-0018. ChatGPT Work web, chat, and mobile are unsupported for Phase 2 local lifecycle
operation.

No verified packaged engine was installed on the qualification workstation. Live `full`
or `verification-only` engine execution is therefore not claimed. All four modes and every
required negative path are exercised by deterministic contract scenarios on native CI,
while ordinary development installation truthfully remains `degraded-guidance`. The
bundled professional baseline is a digest-bound offline projection, not a signed policy
pack or conformance evidence.

The machine [compatibility report](phase2-plugin-compatibility.json) retains exact tested
identities, freshness, native results, untested surfaces, and limitations. The machine
[quality receipt](phase2-plugin-quality-receipt.json) keeps functional, security,
interaction, accessibility, testing, documentation, compatibility, and evidence states
distinct. Accessibility remains `needs-review` because no dedicated assistive-technology
evaluation was retained.

## Verification evidence

Before publication, the complete local suite and documentation validator must pass. Local
Go remains unavailable and is not represented as pass. The completing pull request records
the exact final source revision and native Linux, macOS, and Windows CI run before the
aggregate state can move from `needs-review`.

## Downstream assumptions

- **GitHub executable work:** may consume profile identity and concise receipts, but must
  retain GitHub Project/issue authority and must not infer approvals from plugin chat.
- **Retrofit:** must add focused inspect/retrofit routing and its own immutable plan,
  conflict, history, and recovery qualification; create authority does not apply.
- **Policy distribution:** must replace the projection boundary with independently signed,
  immutable, online/offline resolution and provenance evidence.
- **Release:** must qualify the aggregate release contract, artifact identity, independent
  approvals, signing, publication effects, and audience communication separately.
- **Future plugin upgrades:** invalidate this report when skills, manifest, capability or
  approval contracts, distribution, surfaces, engine protocol, baseline, or native
  semantics change; installation never migrates a repository.
- **Engine packaging:** remains required before live `full` or `verification-only` support
  can be claimed without a separately verified local engine.
