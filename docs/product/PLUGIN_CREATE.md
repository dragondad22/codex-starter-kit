# Codex plugin guided create

**Status:** Implemented development workflow; full execution prerequisites unavailable

**Decision:** [DEC-0018](../decisions/DEC-0018-codex-plugin-compatibility-and-distribution.md)

**Issue:** [#52](https://github.com/dragondad22/codex-starter-kit/issues/52)

Plugin version `0.2.0` adds the focused `$starter-kit-create` skill. It guides a human-owned
brief and owner-persona confirmation through read-only inspection, immutable engine
planning, explicit effect review/approval, exact plan application, and truthful result or
recovery presentation. The standalone engine remains the only lifecycle authority.

## Install or update for development

Review the manifest, both skills, their referenced contracts, scenario cases, and the
repository marketplace before changing local Codex configuration. For a new trusted local
marketplace/install, use:

```text
codex plugin marketplace add <absolute-path-to-this-repository>
codex plugin add codex-starter-kit@codex-starter-kit-development
```

For an already configured development marketplace, use the supported marketplace/plugin
refresh flow and restart Codex before evaluating the final cached tree. Confirm the exact
installed version, source, enabled state, and path with `codex plugin list --json`.

Plugin install/update changes only guided model instructions in the Codex plugin cache. It
does not install/update the engine, Git, Go, baseline/policy material, a repository, or a
managed-repository schema. It adds no connector, app, MCP server, hook, browser, scheduled
task, telemetry route, network authority, repository authority, or content-handling
authorization. Administrator/workspace disablement and Codex sandbox/approval controls
remain authoritative. The plugin itself has no license fee; ordinary Codex use and support
may have separate cost.

## Inputs and special-data boundary

Invoke `$starter-kit-create` or ask specifically to initialize a new repository under
Starter Kit management. The workflow requires:

- an absolute target repository path;
- a human-supplied project brief and explicit approval of that exact brief;
- explicit confirmation of the seed project-owner persona; and
- a `No`, `Yes`, or `Unsure` special-data-handling declaration.

The workflow never drafts missing authority and marks it approved. `Yes` and `Unsure`
trigger the concise project notice: the current route is not assumed verified for specially
handled content, the user must not provide/transmit it until handling authorization and
route assurance exist, metadata-only planning/remediation can continue, and acknowledgment
records receipt only. It grants no handling authorization, classification, legal review,
product assurance, or conformance.

The create-v1 engine does not accept or persist this declaration. The skill identifies it
as a session-scoped workflow fact and coverage limitation. It never claims the created
repository recorded it.

## Capability, plan, and approval flow

Every workflow begins with the non-mutating `starter-kit capabilities` handshake. Full
create requires protocol `1`; `inspect`, `plan`, `apply`, and `status`; a verified engine;
a verified compatible professional baseline with locally available material; native
structured process/file operations; and separately authorized read/mutation access.

The plugin snapshot includes baseline `baseline:professional-engineering:v1` version
`1.0.0` as a digest-bound projection sourced from DEC-0017. Its manifest and content are
available offline under `baselines/professional-v1/`. The projection is guidance and
compatibility material, not a signed policy pack or conformance evidence; full use still
requires external qualification of the containing plugin snapshot and an exact digest
match.

When qualified, the guided flow is equivalent to these reviewed executable/argument
operations:

```text
starter-kit inspect --repository <absolute-path>
starter-kit plan --operation create --repository <absolute-path> --brief <approved-text> --approve-brief --confirm-owner-persona
starter-kit apply --plan <exact-private-plan.json> --plan-id <retained-plan-id>
starter-kit status --repository <absolute-path>
```

The skill retains the exact plan JSON outside the target repository with the plan ID kept
separately. Before any apply, it shows the repository/precondition/plan identities; every
path, ownership, source, and digest; whether the plan is no-change; result-evidence path;
conflicts; policy/control/special-data limits; local effects; and recovery behavior. The
user must approve applying that exact plan. Brief approval, persona confirmation, notice
acknowledgment, or a general original request is not effect approval.

Apply never regenerates or edits a plan. A stale precondition or changed content requires a
new inspection/plan and new approval. Existing human-owned content produces reconciliation,
not overwrite. Exact `applied`, `no_change`, `failed`, conflict, changed-file, recovery,
evidence, stage/cause, and recoverability facts remain distinct. Interrupted setup may
replay only the exact immutable plan when the engine permits it. Rollback failure is an
explicit non-recoverable result, never success.

## Current support, offline use, and fallback

This repository does not publish a verified packaged engine or signed baseline policy
pack. The bundled baseline projection supplies offline guidance material but a source
build, executable filename, marketplace cache, or plugin version cannot prove the engine
or containing snapshot identity. Consequently the ordinary development install selects
`degraded-guidance` and does not inspect, plan, or apply through the plugin. This is the
expected truthful result, not a broken managed-repository status.

The direct command sequence above is a fallback for a user to review only after verified
compatible engine/baseline provisioning and required authority exist. No workflow silently
downloads, replaces, enables, or upgrades those prerequisites. Supported offline create
requires the plugin snapshot/cache, verified engine, baseline material/identity,
compatibility metadata, and trust roots to be provisioned before going offline. A cache hit
alone is insufficient.

The Codex CLI remains the selected development surface. IDE marketplace behavior is
`needs-review` because official documentation conflicts; web/mobile cannot execute the
local engine workflow. Cross-model/client/native qualification remains #54. Guided verify
is a separate [implemented workflow](PLUGIN_VERIFY.md), and the create skill does not run
it automatically.
