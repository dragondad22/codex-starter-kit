# Versions, change records, and releases

The current Codex Starter Kit development product version is `0.3.0`.
`product-version.json` is the schema-versioned authoritative release identity shared by
the standalone CLI, engine capability report, plugin manifest, and plugin capability
contracts. This does not mean `0.3.0` has been published as a supported
GitHub Release. Protocol, schema, policy, baseline, and managed-repository format versions
remain separate compatibility identities.

## Recording a change

Every material pull request adds a schema-v1 JSON record under `changes/unreleased/`.
Externally visible records name their audiences; internal-only records explain why no
external communication is needed. The record remains the human-owned source. The root
`CHANGELOG.md` and audience summaries are generated views.

Use the source-built CLI:

```text
starter-kit changes validate --repository .
starter-kit changes render --repository .
starter-kit changes render --repository . --audience users
starter-kit changes render --repository . --release 0.4.0 --audience users
starter-kit changes check --repository .
```

The supported audiences are users, operators, developers, stakeholders, and security.
The all-audience changelog contains external records only; internal-only records remain
available for audit and release-membership disposition without leaking into release notes.

## Preparing a release

Preparation requires a selected release target, applicable aggregate authority, an
explicit next stable SemVer, an explicit date, and a schema-v1 admission record binding
the exact record IDs to the release Milestone, aggregate release issue, and approver:

```text
starter-kit release prepare --repository . --version 0.4.0 --date 2026-07-15 --admission changes/admissions/0.4.0.json
```

The command validates every pending record and same-release version surface before
mutation. It refuses malformed, unsafe, empty, version-skewed, existing-target,
invalid-date, and non-incrementing preparation. A successful local transaction:

1. archives the exact pending JSON bytes under `changes/releases/<version>/`;
2. records each SHA-256 digest and `state: prepared`, `published: false`;
3. updates `product-version.json`, the plugin manifest, and both plugin capability
   contracts together;
4. regenerates a fresh Unreleased view and the dated prepared release section; and
5. removes only the now-archived pending records.

Same-directory temporary files and backups support native Linux, macOS, and Windows
replacement. A durable `changes/release-transaction.json` journal is written before
mutation. After hard termination, preparation refuses to continue until
`starter-kit release recover --repository .` restores the original files and removes the
partial archive. Publication automation remains downstream work.

## Publishing is separate

Preparation does not commit, tag, push, create a GitHub Release, publish an artifact,
sign, deploy, approve scope, or close a Milestone. A later bounded publication adapter
must operate on the exact reviewed merge commit, require explicit human publication
approval, create and verify the tag and GitHub Release, retain provenance and evidence,
and report partial publication as degraded rather than successful.

The `1.0.0` Milestone and aggregate issue #45 remain the stable-release authority. A
minor or patch release along the way needs its own finite Milestone and aggregate release
issue under the same lifecycle contract.
