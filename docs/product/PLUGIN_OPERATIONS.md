# Plugin installation and operations

This guide covers the Phase 2 development plugin. It is a skills-only adapter for guided
`create`, `status`, and `verify`; the standalone engine and managed repository remain
independent authorities. The tested live surface is Codex CLI `0.144.1` on Linux. That
observation is not a minimum-version promise. See the
[compatibility report](../evidence/phase2-plugin-compatibility.json) for the complete
tested envelope and explicit `needs-review` or unsupported surfaces.

## Before installation

Review the repository revision and `plugins/codex-starter-kit/.codex-plugin/plugin.json`.
The plugin contains Markdown skills, reference contracts, a bundled baseline projection,
and qualification fixtures. It adds no app/connector, MCP server, hook, browser extension,
scheduled task, telemetry service, or remote service.

Installation trusts the reviewed local repository as a marketplace source and allows
Codex to copy that plugin into the user's Codex cache/configuration. It does not install or
update the lifecycle engine, resolve a signed policy pack, migrate a repository, run a
control, or grant repository/network/data authority. The plugin itself has no separate
license or service charge; using an authenticated Codex model consumes the account's
normal model allowance or cost.

Codex model use may transmit the prompt and host-selected context under the chosen Codex
surface's data contract. The plugin adds no separate data path. Do not place secrets,
specially handled content, private evidence, or credentials into a prompt unless the
named route has handling authorization and product assurance. A special-data notice or
acknowledgment is not that authorization or assurance.

## Development installation

From a reviewed checkout, replace the placeholder with its absolute native path:

```text
codex plugin marketplace add <absolute-path-to-this-repository>
codex plugin add codex-starter-kit@codex-starter-kit-development
codex plugin list --json
```

Confirm plugin version `0.3.0`, the source path, `enabled: true`, and the marketplace
identity. Start a new Codex task/session after installation. Desktop or IDE hosts may need
a full application reload; current IDE marketplace behavior remains `needs-review`.
Administrator policy can deny installation or disable the plugin. Do not bypass or
silently change that policy.

## Update boundaries

A local development marketplace reads the reviewed checkout. A Git-backed qualification
marketplace must be refreshed explicitly with `codex plugin marketplace upgrade` and a
named marketplace. Refreshing a marketplace does not update an installed plugin cache.
Remove and add the plugin again only after reviewing the new source and identity, then
start a new task/session or reload the host before evaluating it.

The operations remain separate:

| Operation | Authorized effect | Effects it never implies |
|---|---|---|
| Plugin install/update | Codex plugin configuration and cache | Engine install/update, repository mutation/upgrade, baseline or policy resolution |
| Engine install/update | A separately reviewed engine binary and its local configuration | Plugin update, repository migration, policy change |
| Repository upgrade | Exact reviewed lifecycle-engine migration plan | Plugin/engine installation or policy download |
| Baseline/policy resolution | Verify and select a named immutable input | Plugin/engine update or repository effect |

Each operation requires its own trust source, identity, cost, authority, effect review,
and rollback/removal plan. Conversation or approval of one row grants none of the others.

## Capability and approval behavior

The versioned
[capability model](../../plugins/codex-starter-kit/contracts/capability-model-v1.json)
selects `full`, `degraded-guidance`, `verification-only`, or `unsupported` from evidence;
the mode is workflow availability, not repository conformance. Ordinary development use
currently remains `degraded-guidance` unless a verified compatible packaged engine and
all other requested facts are independently established.

Plan review, repository effects, network access, tool installation, data handling, and
authority changes are separate boundaries. The plugin may prepare or explain one without
approving another. `create` requires exact-plan repository-effect approval. `verify`
requires exact-plan evidence-write approval. `status` is read-only. A denial or
cancellation stops without manufacturing an engine result.

## Offline and restricted use

Supported offline engine operation requires the plugin snapshot/cache, verified engine,
baseline, compatibility metadata, and trust roots to be provisioned and verified before
disconnecting. The bundled baseline is an offline projection of DEC-0017; it is not a
signed policy pack or conformance evidence. Interactive Codex model access may still need
the Codex service network path and normal account usage. When that path is unavailable,
use the independently verified direct engine or CI; do not describe the conversational
plugin as offline-capable merely because its files are cached.

A restricted read-only workspace may support status and, only with explicit evidence-write
authority, verification-only behavior. It never permits create/apply. An administrator-
disabled plugin remains disabled. An offline first run with missing prerequisites stops
with exact provisioning guidance rather than fetching content silently.

## Troubleshooting and recovery

- **Missing engine:** install/select a trusted engine as a separate operation, verify its
  provenance and capability envelope, or use an already verified direct engine/CI path.
- **Incompatible engine:** select separately trusted compatible identities; never change
  repository pins merely to make the plugin proceed.
- **Malformed or conflicting engine output:** mode is `unsupported`; retain only
  authorized redacted diagnostics and stop until the contract is restored.
- **Interrupted operation:** inspect durable engine `status` and evidence before replaying
  the same immutable plan.
- **Cancellation:** inspect status; cancellation is not completion and grants no later
  effect authority.
- **Recoverable operation:** review the retained plan, status, evidence, and recovery
  action before explicit replay or reconciliation approval.
- **Stale task after install/update:** start a new task/session or reload the host, then
  confirm the installed identity with `codex plugin list --json`.
- **Plugin unavailable:** use the documented direct-engine/CI fallback. Do not simulate a
  lifecycle or conformance result conversationally.

## Fallback and removal

The direct verified engine and CI are the independent fallback. They use the same public
lifecycle contract and remain responsible for structured results and evidence. Direct use
does not make an unverified engine trusted or make missing policy pass.

Remove the development plugin and marketplace explicitly:

```text
codex plugin remove codex-starter-kit@codex-starter-kit-development
codex plugin marketplace remove codex-starter-kit-development
```

Removal deletes the plugin's Codex configuration/cache relationship and the configured
marketplace source. It does not remove an engine, change or delete a managed repository,
undo repository evidence, or remove independently provisioned policy/baseline material.
Start a new task/session or reload the host so stale skill context is not mistaken for an
installed plugin.

## Accessibility status

The experience uses structured text, Markdown, and native CLI interaction and does not add
a graphical interface. No dedicated keyboard, screen-reader, cognitive-load, or
alternative-presentation evaluation has been retained, so accessibility remains
`needs-review` rather than an inferred pass.
