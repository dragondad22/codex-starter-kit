# Triage Labels

| Canonical role | GitHub label | Meaning |
|---|---|---|
| `needs-triage` | `needs-triage` | Maintainer evaluation required |
| `needs-info` | `needs-info` | Waiting for reporter information |
| `ready-for-agent` | `ready-for-agent` | Complete agent brief and intended agent executor once Project Readiness is `Ready` |
| `ready-for-human` | `ready-for-human` | Requires human implementation or judgment |
| `wontfix` | `wontfix` | Deliberately not actioned, with rationale |

These are triage routing states, not issue type, severity, priority, execution Status,
Readiness, or Horizon. Those dimensions remain separate. In particular,
`ready-for-agent` identifies who may execute a complete brief; it does not override
Project Readiness. An issue labeled `ready-for-agent` with Readiness `Blocked` must not
start until dependency reconciliation changes Readiness to `Ready`.
