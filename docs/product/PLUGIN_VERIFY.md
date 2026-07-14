# Codex plugin guided verification

Shared installation, update, offline, troubleshooting, fallback, and removal implications
are documented in [Plugin installation and operations](PLUGIN_OPERATIONS.md).

**Status:** Implemented development workflow; qualified engine prerequisite unavailable

**Decision:** [DEC-0018](../decisions/DEC-0018-codex-plugin-compatibility-and-distribution.md)

**Issue:** [#53](https://github.com/dragondad22/codex-starter-kit/issues/53)

Plugin version `0.3.0` adds `$starter-kit-verify` beside the separately disclosed create
and status skills. It guides an explicit scope, lifecycle gate, actor, and authority through
immutable verification planning, exact-plan execution approval, evidence regeneration, and
truthful control/aggregate presentation. The standalone engine owns every evaluation and
evidence result; the model does not decide conformance.

## Installation and authority

Review the manifest, all focused skills/contracts, baseline projection, and scenarios
before adding/updating the development marketplace. The install commands and trust/cost/
removal implications are in the [status](PLUGIN_STATUS.md) and
[create](PLUGIN_CREATE.md) guides. Restart Codex after install/update and confirm plugin
version `0.3.0`, source, enabled state, and cache path with `codex plugin list --json`.

Plugin installation adds model instructions only. It does not regenerate evidence, run a
control, accept risk, alter a repository, install a scanner, update the engine/policy, add
network/data access, or weaken user/workspace/admin/sandbox/approval authority.

## Plan and effect boundary

Invoke `$starter-kit-verify` or ask specifically to verify Starter Kit controls,
conformance, coverage, evidence, or a lifecycle gate. Supply explicit non-secret values
for:

- verification scope;
- lifecycle gate;
- requesting actor; and
- authority for regenerating repository evidence.

The workflow cannot infer actor or authority from login, repository ownership, a risk
record, prior conversation, or the original request. After the non-mutating capability
handshake, a qualified workflow prepares:

```text
starter-kit verify-plan --repository <absolute-path> --scope <scope> --gate <gate> --actor <actor> --authority <authority>
```

The exact JSON plan and its ID are retained separately outside the repository. Review shows
repository/precondition/plan identities and every input. Executing verify may write machine
evidence and an operation event, regenerate `docs/evidence/CONFORMANCE.md`, update its
managed digest, use lifecycle locking, and roll back ordinary persistence failures. The
user must approve those effects for the exact plan before:

```text
starter-kit verify --plan <exact-private-plan.json> --plan-id <retained-plan-id>
```

Metadata entry and plan preparation are not execution approval. Repository drift requires
a new plan and approval. The workflow never repairs controls, retries changed input,
installs an evaluator, or accepts a risk exception by itself.

## Evidence and state presentation

The result preserves verification/evidence identities, ownership/source, scope/gate/
actor/authority, source revision/snapshot, engine/repository/policy versions, timestamp,
aggregate state, every control, coverage limitations, diagnostics, and local evidence/event
paths.

The six states remain exact:

- `pass`;
- `fail`;
- `not-applicable`;
- `not-configured`;
- `needs-review`; and
- `accepted-exception`.

An accepted exception retains its underlying non-pass state and risk evidence. It is not
pass. Aggregate pass is never presented when any control is non-pass, required evidence is
absent, coverage for the stated scope/gate is incomplete, or an evaluator failed. Empty,
partial, malformed, conflicting, or unknown output is `unsupported`, never a partial pass.

Diagnostics remain engine-redacted. Evidence references are shown without automatically
opening or transmitting their content. Further inspection requires separate user request,
repository-read authority, and a safe content route. Stale plans, lock/persistence/
rollback/evaluator failures, and Git-local attempt evidence remain explicit.

## Capability modes, fallback, and limits

`full` may verify when every relevant plugin/engine/baseline/policy/authority fact is
established. `verification-only` may perform only the specifically authorized local
evidence-regeneration workflow when create/apply mutation is unavailable; it cannot create,
repair, migrate, or upgrade. Unknown evidence-write authority is `unsupported`.

No verified packaged engine is currently published. Ordinary development source therefore
selects `degraded-guidance` and does not plan or execute verification through the plugin.
After independent verified provisioning, the direct commands above and direct CI engine
calls are the fallback. CI never depends on plugin routing or conversation.

Supported offline verification requires the plugin snapshot/cache, verified engine,
baseline/policy compatibility inputs, repository pins, and trust roots in advance. No
silent online fallback, install, enablement, or update is allowed. CLI is the selected
development surface; IDE marketplace support remains `needs-review`, web/mobile cannot
execute the local workflow, and cross-client/model/native qualification remains #54.
