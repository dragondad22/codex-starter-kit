# Codex Starter Kit

A Codex-native development system that turns an idea—or an existing repository—into
managed, executable, verifiable work.

The goal is not one-prompt application generation. The goal is to make rigorous software
development feel coherent: standards, testing, security, compliance, documentation,
decisions, GitHub work, stakeholder communication, releases, and upgrades are coordinated
behind a guided workflow and backed by evidence.

> **Project status: Phase 3 development slice.** Empty-repository create, seed
> verification, skills-only plugin status/create/verify workflows, and deterministic
> one-task GitHub transport are implemented for development. No approved live sandbox,
> verified packaged engine, or baseline policy pack is published, so full
> plugin execution remains unavailable; secret scanning, retrofit, upgrades, and broader
> policy/release capabilities are also incomplete. Do not use this repository as a
> production compliance control today.

The current development product version is `0.3.0`. It is not a claim that a supported
GitHub Release has been published. See the [release guide](docs/product/RELEASES.md) and
generated [changelog](CHANGELOG.md).

## What it will provide

- Guided creation of new projects and safe retrofit of existing repositories
- A standalone cross-platform lifecycle engine for deterministic enforcement
- A Codex plugin for progressively disclosed workflows
- Evidence-backed universal and context-triggered policy
- Executable GitHub issues that preserve complete implementation context
- A synchronized GitHub Project and Horizon-based feature roadmap
- Signed, versioned policy packs with offline and air-gapped operation
- Human-owned briefs, personas, decisions, risks, specifications, and communications
- Safe semantic upgrades for policy, templates, tooling, and managed repositories
- Native Linux, macOS, and Windows support

## Architecture at a glance

```text
Codex plugin
    ↓ guided workflows, routing, explanations, approvals
Standalone lifecycle engine
    ↓ create · retrofit · inspect · plan · apply · verify · status · upgrade
Managed repository
    ↓ pinned policy, decisions, evidence, provenance, GitHub state
```

The lifecycle engine is the primary test and enforcement seam. The plugin is the
preferred Codex experience; it is not the sole conformance authority.

The current development CLI implements a read-only `capabilities` handshake plus
`inspect`, `create`, `plan`, `apply`, `status`, seed `verify`, and credential-free
in-memory `manage-task`. A native Go GitHub adapter implements the same engine seam for
deterministic integration; live use remains `not-configured` pending #73.
See the [Phase 1 engine interface](docs/architecture/LIFECYCLE_ENGINE.md) for its JSON
contract, ownership model, and explicit limitations, and the
[Work Manager contract](docs/architecture/WORK_MANAGER.md) for the one-task route. The
[GitHub adapter contract](docs/architecture/GITHUB_ADAPTER.md) defines identity,
transport, permission, recovery, and current live-evidence limitations.

The repository also contains installable development status, create, and verify skills.
See the [plugin status guide](docs/product/PLUGIN_STATUS.md) and
[guided create guide](docs/product/PLUGIN_CREATE.md) and
[guided verify guide](docs/product/PLUGIN_VERIFY.md) before adding its local marketplace;
installation adds guided instructions but does not install or verify the engine/baseline.

## Build and use the Phase 1 engine

The supported distribution is currently a source build. Install native Git and Go 1.26.5,
clone the repository, and run this from the repository root:

```text
go build -o starter-kit ./cmd/starter-kit
```

Obtain Git and Go from publishers your environment trusts. Both are local executables with
access to the repository; Go may contact its configured toolchain/module source during a
build, while the built Phase 1 engine itself performs no network operation. Both are
available without a product license fee, but organizational support or distribution costs
may differ. Go 1.26.5 is the tested build version and native Git is a runtime requirement;
there is no supported prebuilt or no-Git fallback yet.

On Windows the output may be named `starter-kit.exe`. The executable and JSON contract are
the same on Linux, macOS, and Windows; no WSL, Git Bash, Bash, PowerShell, container, or
Codex client is required. A caller can also use `go run ./cmd/starter-kit` during source
development.

The minimal direct flow is:

```text
git init <repository>
starter-kit inspect --repository <repository>
starter-kit capabilities
starter-kit version
starter-kit create --repository <repository> --brief <approved-text> --approve-brief --confirm-owner-persona
starter-kit apply --plan <create-plan.json> --plan-id <sha256-plan-id>
starter-kit status --repository <repository>
starter-kit verify-plan --repository <repository> --scope repository --gate development --actor <actor> --authority <authority>
starter-kit verify --plan <verify-plan.json> --plan-id <sha256-plan-id>
starter-kit manage-task --input <managed-task-v1.json>
```

`create` and `verify-plan` emit reviewable JSON plans on standard output. Retain each exact
document and separately retain its `plan_id` using the native process/file facilities of
your environment, review it, then pass it to `apply` or `verify`. Results are JSON on
standard output; structured apply/reconciliation failures are JSON on standard error.
The engine never installs Git or Go, changes sandbox authority, or performs network effects.
`manage-task` strictly reads one JSON envelope containing desired intent plus normalized
in-memory capability and observation, writes self-digested credential-free state below
the selected repository, and returns the complete inspect/plan/apply/verify/status result.
It is a deterministic contract route, not live GitHub qualification.

See the [support matrix](docs/architecture/SUPPORT_MATRIX.md) for the exact tested envelope,
runtime requirements, capability gaps, and evidence model.

## Start here

- [Product requirements](docs/product/PRD.md)
- [Personas](docs/product/PERSONAS.md)
- [Architecture](docs/architecture/ARCHITECTURE.md)
- [Lifecycle state machines](docs/architecture/LIFECYCLES.md)
- [Policy-pack map](docs/architecture/POLICY_PACKS.md)
- [Current support matrix](docs/architecture/SUPPORT_MATRIX.md)
- [Plugin status tracer](docs/product/PLUGIN_STATUS.md)
- [Plugin guided create](docs/product/PLUGIN_CREATE.md)
- [Plugin guided verify](docs/product/PLUGIN_VERIFY.md)
- [Versions, change records, and releases](docs/product/RELEASES.md)
- [Generated changelog](CHANGELOG.md)
- [Implementation roadmap](docs/roadmap/IMPLEMENTATION_ROADMAP.md)
- [Discovery and approved decisions](docs/discovery/CODEX_STARTER_KIT_REVIEW.md)
- [Durable decision index](docs/decisions/INDEX.md)
- [Documentation index](docs/README.md)

## How work is organized

All work is tracked in [GitHub Issues](https://github.com/dragondad22/codex-starter-kit/issues)
and the repository's single [Codex Starter Kit Project](https://github.com/users/dragondad22/projects/8).

- Feature ideas enter as `type:feature` issues in `Status: Backlog`.
- `Horizon: Now / Next / Later` expresses rolling roadmap intent, not release membership.
- A named release uses one finite GitHub Milestone plus an aggregate release issue for
  scope, gates, evidence, approval, and publication.
- An implementation issue becomes `Ready` only when an AI or developer can execute it
  without inventing missing decisions.
- When conversation surfaces durable untracked work or decisions, agents search existing
  issues and offer the appropriate capture action before material implementation.
- Normal delivery is Ready issue → issue-named branch → pull request → required gates →
  squash merge.

The live GitHub Project is the roadmap authority. Documents do not duplicate project
status or feature ordering.

## Contributing

This project welcomes issues and external pull requests. Start with
[CONTRIBUTING.md](CONTRIBUTING.md). Please open or identify an issue before substantial
implementation; external PRs enter the same triage workflow as issues.

If you discover a vulnerability, do not open a public issue. Follow
[SECURITY.md](SECURITY.md).

## Trust and limitations

The design follows a strict rule: no evidence means no pass. A managed result must state
what was checked, what failed, what was not applicable, what could not be checked, and
what needs qualified human review. The first release will detect projects that need
special data handling, apply universal safeguards, and report limitations truthfully; it
will not claim that Codex, connected tools, or a development environment are certified or
verified for highly sensitive content. It will not replace legal, security, privacy,
accessibility, safety, or domain experts where qualified judgment is required.

## License

[MIT](LICENSE) © 2026 dragondad22
