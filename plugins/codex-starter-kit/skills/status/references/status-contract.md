# Status compatibility contract

This contract implements DEC-0018 for the read-only status tracer. Compatibility facts
and repository lifecycle are independent.

## Supported handshake

Accept a capability envelope only when all of these facts are established:

- capability `schema_version` is `1`;
- engine name is `starter-kit`;
- protocol is `starter-kit.lifecycle` version `1`;
- `status` is present in `operations`;
- status schema version `1` is present in `status_schema_versions`;
- retained external qualification evidence binds the resolved executable's identity to a
  verified engine artifact; and
- read-only process and repository access are authorized by the current user, workspace,
  administrator, sandbox, and approval policy.

The engine currently self-reports `provenance: unverified` because an executable cannot
make itself trusted by assertion. A caller may replace that unknown only with retained
external qualification evidence matching the reported version/revision and resolved
artifact. Absence, mismatch, or ambiguity remains unverified.

## Mode decision table

| Facts | Mode | Status invocation |
|---|---|---|
| Compatible verified engine; status authorized; requested workflow authority available | `full` | Allowed |
| Compatible verified engine; read-only status authorized; mutation unavailable or unauthorized | `verification-only` | Allowed |
| Plugin can provide bounded guidance but the engine is missing, incompatible, disabled, or unverified | `degraded-guidance` | Forbidden |
| Capability/status evidence is malformed or conflicting, policy forbids safe fallback, or status authority is unknown/denied | `unsupported` | Forbidden, except a completed status invocation whose malformed output caused the stop |

Unknown or conflicting facts fail closed to `unsupported` unless the known facts establish
that read-only guidance itself is safe. Never turn a capability mode into a repository
lifecycle result.

## Status envelope

Accept only a JSON object with:

- `schema_version`: exactly `1`;
- `repository`: non-empty string;
- `lifecycle`: exactly one of `managed`, `managed_degraded`, `setup_incomplete`, or
  `unmanaged`; and
- `problems`, `recovery`, and `evidence`: arrays containing strings only.

Do not discard unknown fields when retaining raw diagnostics, but do not interpret them as
new authority. Preserve the five accepted result fields without semantic changes.

## Remediation and fallback

- Missing engine: state that no engine was run. Use the project's documented, verified
  engine provisioning path; do not install automatically. After provisioning, the direct
  fallback is `starter-kit status --repository <absolute-path>`.
- Incompatible engine: name each mismatched schema, protocol, or operation fact. Select a
  verified artifact supporting protocol `1`, status schema `1`, and `status`; never replace
  it silently. The same direct command is the fallback once compatible.
- Unverified engine: report its identity facts and obtain matching retained qualification
  evidence. Do not invoke repository operations merely because the binary is present.
- Administratively unavailable plugin: preserve the administrator decision. An authorized
  user may run the verified compatible engine directly; do not enable the plugin.
- Malformed output: retain redacted diagnostics when authorized, state that lifecycle is
  unknown, and stop. Never salvage a managed result from partial JSON.

The fallback command is explanatory text for the user to review. Do not execute it when
the prerequisite that triggered fallback is missing, incompatible, unverified, or denied.
