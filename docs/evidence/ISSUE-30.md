# Issue #30 — Native semantic equivalence and Phase 1 runtime support

**Date:** 2026-07-13
**Issue:** [#30](https://github.com/dragondad22/codex-starter-kit/issues/30)
**Parent:** [#2](https://github.com/dragondad22/codex-starter-kit/issues/2)

## Delivered outcome

Every Linux, macOS, and Windows CI job now exercises the same complete Phase 1 lifecycle
scenario and retains a self-digested JSON report. The report separates portable product
semantics from native mechanisms and capability states. An aggregate job validates all
three documents, requires one report per `GOOS`, and fails when their semantic digests
differ. The final validation gate requires both the native matrix and semantic comparison.

The [support matrix](../architecture/SUPPORT_MATRIX.md) replaces the former foundation-only
statement with a bounded initial source-runtime claim. It names build/runtime requirements,
filesystem assumptions, evidence retention, supported operations, and explicit nonclaims.
The root README documents source build and direct CLI use without a mandatory platform
shell, WSL, container, Codex client, or network effect.

## Native evidence contract

| Evidence | Contents | Truth boundary |
|---|---|---|
| `phase1-native-ubuntu-latest` | Resolved image/runner/Go/Git facts, Linux capabilities, portable semantics, semantic and evidence digests | One completing-PR Linux run |
| `phase1-native-macos-latest` | Resolved image/runner/Go/Git facts, macOS capabilities, portable semantics, semantic and evidence digests | One completing-PR macOS run |
| `phase1-native-windows-latest` | Resolved image/runner/Go/Git facts, Windows capabilities, portable semantics, semantic and evidence digests | One completing-PR Windows run |
| `phase1-native-summary` | Owned/sourced/self-digested comparison with three validated evidence paths, resolved platform/architecture set, shared semantic digest, `equivalent: true` | Aggregate comparison, not a substitute for individual capability facts |

The completing pull request is the durable link to the exact 30-day CI artifacts. Run
[`29268327253`](https://github.com/dragondad22/codex-starter-kit/actions/runs/29268327253)
captured source revision `c29eaf3441c6adfaf5c849c262988a0c7d45d4b3`:

| Target | Resolved image | Architecture | Go / Git | Report evidence digest |
|---|---|---|---|---|
| Linux | `ubuntu24` `20260705.232.1` | runner `X64`; `linux/amd64` | Go 1.26.5 / Git 2.54.0 | `sha256:fe0736bc2ecd27d9507adc5edc90e7296b457cbfc0b889c54d57ad7a1a6212c8` |
| macOS | `macos26` `20260630.0213.1` | runner `ARM64`; `darwin/arm64` | Go 1.26.5 / Git 2.55.0 | `sha256:2aebcbd1783732c39aaa1789fe6e6c15d802772d279bbf95f26f33893a3cb54d` |
| Windows | `win25-vs2026` `20260628.158.1` | runner `X64`; `windows/amd64` | Go 1.26.5 / Git 2.54.0.windows.1 | `sha256:242fe43b9c6a7699e30047fed0b6d63f717a17b3eb7c0fda0204cbc7c1606d8a` |

All three reports produced semantic digest
`sha256:38d2405d313853059f4faae8424a0a302775f8e3ddc70fddb81f0d319b7329ad`.
The aggregate summary reported `equivalent: true` for `darwin/arm64`, `linux/amd64`, and
`windows/amd64`, with evidence digest
`sha256:dd3d8d84821010f355673d60de170caecec3936fd92def60f0c67970e0f0c81e`.

## Phase 1 roadmap coverage

| Phase 1 roadmap obligation | Delivered issue/evidence | Disposition and gap impact |
|---|---|---|
| Schemas for project facts, policy lock, layout, managed files, plans, results, routes, and evidence | #26, #27, #29; engine JSON contracts | Delivered for create-v1/seed verify; schema evolution remains upgrade work |
| Select engine language/package/signing approach | #25; DEC-0015 | Go 1.26.5 selected; source build delivered; signing and packaged distribution remain later release work |
| Implement `inspect`, `plan`, `apply`, `status`, and `verify` for local filesystem/Git | #26, #27 | Delivered through public engine and CLI seams; `retrofit`/`upgrade` remain later phases |
| Implement truthful seed controls for secrets, ownership, coverage, recovery, and routes | #27–#29 | Explicit states delivered; secrets are `not-configured`, recovery is `needs-review` for unversioned builds |
| Render minimal `AGENTS.md`, brief, decision index, and conformance summary | #26, #27 | Delivered with ownership/provenance and managed-contract validation |
| Seed/confirm persona registry and route governed artifacts through stable persona IDs | #26 | Seed owner persona and human-owned registry delivered; issue/spec/communication persona routing belongs to later producing modules |
| Prove idempotence, preconditions, conflicts, rollback, hostile paths, and native equivalence | #28–#30; native report/summary | Delivered for Phase 1 create/verify boundary; external effects and later lifecycle operations require their own proof |

## Governing-decision coverage

| Decision | Phase 1 disposition | Downstream impact |
|---|---|---|
| DEC-0001 first-release scope | Engine source-runtime boundary and sensitive-data nonclaims are explicit | Special-data routes remain issue #21/later assurance work |
| DEC-0002 applicability/risk | Policy lock and explicit control states exist without fabricated applicability | Full policy compilation, exceptions, and risk lifecycle remain policy phases |
| DEC-0003 three-layer distribution | Standalone engine and managed-repository layers are exercised | Codex plugin layer is Phase 2 and must consume the same seam |
| DEC-0004 state/document authority | Machine state, generated views, and human-owned records remain distinct | Retrofit/upgrade must preserve this ownership history |
| DEC-0005 GitHub executable work | Phase 1 delivery uses Issues/Project/PR gates; engine work manager is not claimed | Phase 3 implements managed-repository GitHub synchronization |
| DEC-0006 stage-specific enforcement | Missing scanner/release provenance remains explicit non-pass | Later release and policy gates must block at their own risk boundary |
| DEC-0007 policy distribution/layout | Layout and truthful unconfigured lock are seeded | Signed packs, registry, cache, and offline resolution remain Phase 5 |
| DEC-0008 Git/release contract | Structured local Git and protected issue-linked delivery are proven | Product release adapters/signing remain Phase 6 |
| DEC-0009 native platform support | One semantic digest is required across native Linux/macOS/Windows | New architectures/filesystems require new evidence before support expansion |
| DEC-0010 Claude migration | No runtime compatibility debt is introduced | Semantic migration remains Phase 4 |
| DEC-0011 governed breadcrumbs | Minimal generated `AGENTS.md` and stable routes validate | Later modules extend routes without loading unrelated context |
| DEC-0012 managed conformance | Explicit control states, evidence digests, coverage limits, and human summary exist | No aggregate pass is claimed while secrets/recovery remain non-pass |
| DEC-0013 question/research work | Repository governance exists; Phase 1 engine does not claim work-item rendering | Phase 3 work manager consumes established issue contracts |
| DEC-0014 finite release targets | Phase 1 closure is evidence-backed but is not a product release | A future named release requires milestone/aggregate-release gates |
| DEC-0015 lifecycle toolchain | Pinned Go toolchain and language-neutral observable JSON behavior are proven | Package/signing choices remain a separate versioned-release obligation |

## Relevant PRD-story coverage

Stories are included when Phase 1 creates their engine/state/evidence prerequisite or makes
a user-visible native support claim. Stories wholly owned by retrofit, policy distribution,
GitHub work management, release, or upgrade are dispositioned through the decision and
downstream matrices rather than misrepresented as Phase 1 delivery.

| Story | Phase 1 outcome | Disposition |
|---|---|---|
| 1 — guided brief/inception | Engine accepts only an explicitly approved brief/persona confirmation | Partial prerequisite; guided interaction belongs to plugin #3 |
| 2 — empty repository create | Managed repository is created through reviewed plan/apply | Delivered for the bounded create-v1 contract |
| 5 — explicit project facts | Approved brief/persona and seed project facts are structured | Partial; broader detected/applicability facts grow with later capabilities |
| 6 — special-data declaration/limits | Current absence of verified routes is disclosed, never a pass | Declaration/routes remain issue #21 and later assurance phase |
| 7 — universal controls | Stable seed control IDs and explicit states are evaluated | Partial; secret scanner is `not-configured` |
| 9 — explicit control evidence | Seed verification retains machine evidence and human summary | Delivered; aggregate remains `needs-review` |
| 12 — solo owner transparency | Seed owner persona is explicit | Separation-of-duties evaluation remains later policy/persona work |
| 16–18 — routing and layout | Minimal `AGENTS.md`, routes, persona registry, and logical layout are seeded | Delivered for Phase 1 artifacts |
| 29 — missing tooling blocks safely | Missing scanner and unbound release provenance remain non-pass | Delivered truth semantics; later gates consume them |
| 32–33 — deterministic CI and evidence views | Direct verify, native report comparison, machine evidence, and human summary exist | Delivered for Phase 1 scope |
| 36 — conflicts stop | Structured reconciliation preserves ambiguous/human work | Delivered for create-v1 |
| 37–38 — native Windows and equivalent macOS/Linux semantics | Same native suite and canonical digest are required | Delivered within the published runner/filesystem envelope |

## Cross-cutting Phase 1 definition of done

| Obligation | Evidence/disposition |
|---|---|
| Ready executable issues | #25–#30 delivered through the required issue/branch/PR flow |
| Engine-seam and adapter tests | Native suite plus `internal/nativeevidence` black-box scenario |
| Native CI evidence | Three reports and one aggregate summary retained by the completing PR |
| Threat/failure modes | ISSUE-28 hostile-input matrix and ISSUE-29 recovery matrix |
| User/developer/stakeholder documentation | README direct use, engine contract, support matrix, and issue evidence |
| Policy IDs/breadcrumbs | Six stable `CORE-*` results and route validation |
| Upgrade/migration implications | Explicitly deferred; ownership/schema/support assumptions recorded below |
| Human summary/machine evidence | Generated conformance summary, verification evidence, native reports/summary |
| No false pass | Secrets `not-configured`; recovery `needs-review`; unsupported capabilities explicit |

## Downstream assumptions

| Consumer | May rely on | Must not assume |
|---|---|---|
| Plugin/create flow | Public JSON seam, stable plans, structured reconciliation, native source runtime | Packaged engine availability, Codex-client compatibility, or broader authority |
| Retrofit/migration | Portable path policy, ownership classes, status states, immutable evidence | Authority to overwrite existing content or reuse create-only recovery blindly |
| Policy distribution | Versioned state/evidence semantics, local/offline engine operation | Configured scanner, signed packs, registry trust, or aggregate conformance pass |
| Release work | Native gate pattern, content identity, explicit coverage limits | Signed binary provenance, installer/package support, or release readiness |
| Future platform expansion | Native report schema and semantic comparison | Equivalence on untested architecture, OS version, filesystem, ACL, or reparse behavior |

## Verification commands

```text
go test ./...
go vet ./...
go test -race ./...
GOOS=windows GOARCH=amd64 go test -c ./engine -o <engine-test.exe>
GOOS=windows GOARCH=amd64 go test -c ./internal/nativeevidence -o <native-evidence-test.exe>
GOOS=windows GOARCH=amd64 go build -o <phase1-evidence.exe> ./cmd/phase1-evidence
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
go run ./cmd/phase1-evidence capture --output <evidence.json>
git diff --check
```

The native CI aggregate comparison and resolved environment evidence are recorded in the
completing pull request after all three hosted runners pass.
