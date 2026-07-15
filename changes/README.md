# Change records

Files under `unreleased/` are the authoritative, human-owned source for release
communication. `CHANGELOG.md`, audience-specific summaries, and prepared release manifests
are generated views. Do not edit generated release summaries independently.

Every material pull request adds one or more schema-v1 JSON records. An externally visible
record declares at least one audience. Internal-only work instead sets `internal_only` to
`true`, leaves `audiences` empty, and explains the exclusion in `internal_disposition`.

Supported categories are `added`, `changed`, `deprecated`, `removed`, `fixed`, and
`security`. Supported audiences are `users`, `operators`, `developers`, `stakeholders`,
and `security`. Record and component identifiers use lowercase letters, digits, and single
hyphens. Every record links at least one issue or pull request.

Use the checked-in product CLI through its stable public commands:

```text
starter-kit changes validate --repository .
starter-kit changes render --repository .
starter-kit changes render --repository . --audience operators
starter-kit changes render --repository . --release 0.4.0 --audience operators
starter-kit changes check --repository .
starter-kit release prepare --repository . --version 0.4.0 --date 2026-07-15 --admission changes/admissions/0.4.0.json
starter-kit release recover --repository .
```

`changes validate` checks records and version synchronization without writing. `changes
render` writes Markdown to standard output. `changes check` additionally requires the
checked-in `CHANGELOG.md` to equal the deterministic all-audience view.

`release prepare` is a local filesystem transaction. It requires a greater stable SemVer,
explicit date, and schema-v1 admission naming the Milestone, aggregate release issue,
approver, and exact included record IDs. It archives exact record bytes and the admission
with SHA-256 digests under
`releases/<version>/`, updates same-release version surfaces, regenerates the changelog,
and records `state: prepared` plus `published: false`. It never creates a commit, tag,
GitHub Release, package, signature, deployment, approval, or publication claim. Those
effects require a later approved release adapter operating on the exact merged candidate.
Preparation writes `changes/release-transaction.json` before mutation. If execution is
interrupted, the next preparation refuses to proceed until `release recover` restores the
original files and removes the partial archive.

Product release versions are independent from protocol, schema, policy-pack, professional
baseline, and managed-repository format versions. Do not synchronize those compatibility
identities merely because the product version changes.
