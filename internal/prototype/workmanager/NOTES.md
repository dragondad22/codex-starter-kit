# Work Manager reconciliation prototype

**Status:** Throwaway logic prototype for GitHub issue #64. Do not treat this package or
its JSON as a production compatibility contract.

## Question

What lifecycle-engine-facing interface and versioned data contract let the Work Manager
own desired issue, hierarchy, readiness, Project, review, and completion state while an
in-memory double and production GitHub adapter own transport?

Run the interactive harness from the repository root:

```text
go run ./cmd/work-manager-prototype
```

The pure module is `internal/prototype/workmanager`; the command only reads keys and
renders the complete state after every action. State is in memory and contains no real
credentials, network calls, filesystem operations, or GitHub mutations.

## Representative scenarios

1. Press `p`, `a`, `p`. The seeded closed-issue drift from #15 is corrected once; the
   completed receipt remains, and the second plan contains only remaining semantic
   differences.
2. Continue applying and planning until the `future:sandbox-matrix` create is first,
   then press `l`, `u`, `p`. A lost create response becomes `ambiguous`; marker lookup
   discovers one object and re-planning does not create a duplicate.
3. Press `p`, `a`, `p`, `r`, `h`, `p`, `r`, `h`, `p`. The completed effect stays
   observed while the rate-limited remainder becomes credential-free desired intent.
   The second bounded attempt exhausts the retry allowance and blocks another plan until
   `t`, `h`, `p` observes a reset and repeats the handshake.
4. Press `p`, `m`, `a`. A changed option ID invalidates the retained plan before its
   effect. Press `g`, `h`, `p` to
   show that only a new Work Manager input revision—not the adapter—accepts migrated IDs.
5. Press `o`, `p`, `n`, `p`, `h`, `p`. Offline state queues intent rather than raw HTTP;
   reconnect cannot apply until identity, capability, and preconditions are refreshed.
6. Press `c`, `p`. Completing blocker #64 promotes the sandbox ticket from `Blocked` to
   `Ready`, but leaves Status `Backlog` until it is explicitly selected. Parent #4 stays
   `In progress` while children remain.
7. Press `p`, `s`, `a`. A changed governed source revision invalidates the retained plan
   before any effect.

The seed also keeps #16's promoted decision record separate from its closed issue and
models #46's Phase as inherited parent context rather than copying the field to the task.
Review requirements are separate data from checks, implementation, outcome approval, and
completion.

## Answer to lift into production

Use four versioned boundaries behind the lifecycle-engine `plan`/`apply` seam:

- `DesiredIntent`: credential-free, source-bound Work Manager policy and desired state,
  including stable managed IDs, native relationships, lifecycle fields, review roles,
  promotion/completion evidence routes, and immutable target identities.
- `Capability` plus `Observation`: adapter-reported actor/mode, permissions, budgets,
  limitations, immutable GitHub IDs, configuration revision, and normalized current
  state. Partial or stale observations are not success.
- `Plan`: immutable preconditions and a semantic effect delta bound to desired-source,
  observation, repository, Project, field, and option identities.
- `EffectReceipt`: one explicit `applied`, `ambiguous`, denied, rate-limited, or other
  non-pass outcome with attempt and recovery evidence. Completed receipts and refreshed
  observations survive partial failure.

The Work Manager derives readiness, parent, dependency, phase, review, and completion
policy before planning. The adapter may observe and execute effects, but it must not
select credentials, broaden authority, decide policy, or translate missing evidence into
pass. Creates reconcile by a stable non-secret managed marker; updates reconcile by
immutable IDs and semantic comparison. Queues retain desired intent, hashes, and expiry,
never credentials or transport requests.

## Delete or absorb

Delete the terminal command and this entire prototype package after the production Work
Manager contract and in-memory adapter absorb the prototype-supported boundary. Do not preserve the
prototype's seed IDs, terse error strings, synthetic revision scheme, or absence of real
schema validation/tests. Production work still needs schema design, durable state and
evidence storage, expiry, pagination, GraphQL partial-error handling, permission manifests,
bounded retry scheduling, and sandbox-qualified adapter behavior.
