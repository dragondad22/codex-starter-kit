# Codex Starter Kit

A Codex-native development system that turns an idea—or an existing repository—into
managed, executable, verifiable work.

The goal is not one-prompt application generation. The goal is to make rigorous software
development feel coherent: standards, testing, security, compliance, documentation,
decisions, GitHub work, stakeholder communication, releases, and upgrades are coordinated
behind a guided workflow and backed by evidence.

> **Project status: Phase 1 implementation.** The initial lifecycle-engine create seam is
> implemented for development, while verification, security hardening, recovery, runtime
> support, and the plugin remain incomplete. Do not use this repository as a production
> compliance control today.

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

The current development CLI implements `inspect`, `create`, `plan`, `apply`, `status`, and
seed `verify`.
See the [Phase 1 engine interface](docs/architecture/LIFECYCLE_ENGINE.md) for its JSON
contract, ownership model, and explicit limitations.

## Start here

- [Product requirements](docs/product/PRD.md)
- [Personas](docs/product/PERSONAS.md)
- [Architecture](docs/architecture/ARCHITECTURE.md)
- [Lifecycle state machines](docs/architecture/LIFECYCLES.md)
- [Policy-pack map](docs/architecture/POLICY_PACKS.md)
- [Current support matrix](docs/architecture/SUPPORT_MATRIX.md)
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
