# Foundation and Phase 1 Support Matrix

**Status:** Active development evidence
**Scope:** Repository foundation plus the unreleased Phase 1 create slice

The lifecycle-engine create slice exists, but product runtime support is not published
until #30 closes the complete Phase 1 native-equivalence contract. The plugin, verified
sensitive-data routes, and regulatory coverage do not exist. This matrix must not be
presented as a production, sensitive-data-route, or regulatory support claim.

| Environment | Foundation validation | Evidence |
|---|---|---|
| Ubuntu, current GitHub-hosted image | Required | Native matrix job |
| macOS, current GitHub-hosted image | Required | Native matrix job |
| Windows, current GitHub-hosted image | Required | Native matrix job |
| Python 3.12 | Pinned in CI | `actions/setup-python` by immutable commit |
| Go 1.26.5 | Pinned in CI | `actions/setup-go` by immutable commit |

Every native job runs the same semantic commands:

```text
python -m unittest discover -s tests -p "test_*.py"
python scripts/validate_docs.py
go test ./...
```

The validator uses only the Python standard library. GitHub issue forms and the label
manifest are JSON-compatible YAML so their structure parses identically without an
unpinned package installation.

## Currently verified invariants

- Required public, agent, decision, template, manifest, workflow, and support files exist.
- Local Markdown links remain inside the repository and resolve.
- D1–D15 have unique stable decision identities, targets, and substantive source anchors.
- Issue forms parse and reference known labels with unique body IDs.
- The label manifest has unique names, valid colors, and descriptions.
- Workflow actions are pinned, all three native runners are present, and no explicit
  platform shell is required.
- Engine-seam tests exercise an empty real Git repository through inspect, create/plan,
  apply, status, stale-precondition rejection, and explicit no-change behavior.
- Verification tests exercise explicit seed-control states, evidence-backed pass rules,
  deterministic clock input, semantic equivalence, machine evidence, human summary
  regeneration, and post-verification managed-contract validity.
- Hostile-input tests exercise traversal and absolute paths, portable reserved names,
  unsafe normalization, case-fold collisions, user-owned directory preservation,
  symlink roots/parents/artifacts and canonicalized ancestor aliases where supported, a
  native Windows directory junction, exact create ownership/provenance, sanitized Git execution that
  suppresses executable repository-local fsmonitor configuration, fixture-secret rejection
  and redaction across repository paths/create/apply/verification, and self-consistent
  malicious state/layout/route/ownership data.
- Recovery tests exercise exact applied/no-change replay, changed-input and human-content
  reconciliation, active and dead-stale lifecycle leases, lease-owned release, abandoned
  stage evidence, unknown staging-content preservation, same-plan committed-prefix resume,
  stale repository/Git preconditions, rollback, failed postconditions, and truthful
  `setup_incomplete` status.

## Deferred support decisions

DEC-0015 selects Go 1.26.5 for contributor and CI builds. Issue #30 will publish exact
minimum OS versions, CPU
architectures, filesystems, installer/package behavior, Codex client compatibility, and
external runtime requirements after #27–#29 close their verification, hostile-input, and
recovery obligations. Until then, the matrix records development evidence rather than a
released runtime promise.
