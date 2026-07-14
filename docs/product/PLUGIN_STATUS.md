# Codex plugin status tracer

**Status:** Installable development slice

**Decision:** [DEC-0018](../decisions/DEC-0018-codex-plugin-compatibility-and-distribution.md)

**Issue:** [#51](https://github.com/dragondad22/codex-starter-kit/issues/51)

The Codex Starter Kit plugin currently provides one progressively disclosed, read-only
skill: `$starter-kit-status`. It routes explicit invocation and focused questions about a
repository's Starter Kit lifecycle to the standalone engine. It does not provide guided
create or verify yet, and it is never the conformance authority.

## Review and install for development

Review the local plugin manifest, status skill, compatibility contract, and marketplace
entry before installation. From a trusted clone, add this repository as a development
marketplace and install the plugin:

```text
codex plugin marketplace add <absolute-path-to-this-repository>
codex plugin add codex-starter-kit@codex-starter-kit-development
```

Restart Codex after installation or update so a new task loads the final cached file tree.
Confirm the installed identity and enabled state with:

```text
codex plugin list --json
```

The marketplace is a local development source, not a signed publication channel. Adding
it lets Codex read the marketplace metadata and copy the plugin into its local cache.
Installation adds model instructions to the enabled Codex context. It adds no connector,
MCP server, hook, browser access, scheduled task, telemetry path, repository permission,
or external-service authentication. It does not install or update Git, Go, the lifecycle
engine, a baseline/policy pack, or a managed repository. The plugin has no product license
fee; normal Codex use and organizational support may have separate cost.

Remove the plugin or marketplace through the corresponding `codex plugin remove` or
`codex plugin marketplace remove` command if the reviewed source or workspace policy no
longer permits it. Administrator and workspace controls remain authoritative.

## Invoke status

Use `$starter-kit-status` explicitly or ask a focused question such as “Is this repository
managed by Starter Kit?” The skill first requests the engine's non-mutating
`capabilities` envelope, then selects one workflow capability mode. Only a compatible,
externally verified engine with authorized read access may receive:

```text
starter-kit status --repository <absolute-path>
```

Both calls are executable-plus-argument-vector operations; repository content is never a
shell program. Status returns JSON containing `repository`, `lifecycle`, `problems`,
`recovery`, and `evidence`. The skill preserves them. Workflow mode and repository
lifecycle are separate: neither `full` nor `verification-only` means the repository
passed, and `managed_degraded`, `setup_incomplete`, and `unmanaged` remain explicit.

## Compatibility and current limitation

The plugin supports lifecycle protocol `1`, status schema `1`, and an engine that reports
the `status` operation. A filename or self-reported build identity is not provenance. The
source engine deliberately reports `provenance: unverified`; matching retained external
qualification evidence must verify the resolved artifact before the skill invokes a
repository operation.

No verified packaged engine is published in this development slice. Consequently, an
ordinary source build or PATH discovery currently selects `degraded-guidance`, gives exact
remediation and the direct-engine command for later use, and does not run status. This is
a truthful product limitation, not a managed success. A direct source-built engine remains
available for development and CI independently of the plugin, subject to the source-runtime
[support matrix](../architecture/SUPPORT_MATRIX.md).

Missing, incompatible, disabled, unverified, denied, malformed, or conflicting capability
facts never trigger installation, replacement, online fallback, authority changes, or a
partial managed result. Unsupported IDE marketplace behavior remains `needs-review`; web
and mobile cannot run this local engine workflow. The Codex CLI is the selected Phase 2
development surface, while desktop and additional surfaces remain qualification work in
#54.

Supported offline status requires the marketplace/plugin, verified engine, matching
qualification evidence, baseline compatibility inputs, and trust roots to be provisioned
in advance. A cache hit alone is not verified offline support.
