# Codex Starter Kit

A Codex-native development system that turns an idea—or an existing repository—into
managed, executable, verifiable work.

The goal is not one-prompt application generation. The goal is to make rigorous software
development feel coherent: standards, testing, security, compliance, documentation,
decisions, GitHub work, stakeholder communication, releases, and upgrades are coordinated
behind a guided workflow and backed by evidence.

> **Project status: design and foundation.** The product architecture and requirements
> are documented, but the lifecycle engine and plugin are not implemented yet. Do not use
> this repository as a production compliance control today.

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

## Start here

- [Product requirements](docs/product/PRD.md)
- [Personas](docs/product/PERSONAS.md)
- [Architecture](docs/architecture/ARCHITECTURE.md)
- [Lifecycle state machines](docs/architecture/LIFECYCLES.md)
- [Policy-pack map](docs/architecture/POLICY_PACKS.md)
- [Implementation roadmap](docs/roadmap/IMPLEMENTATION_ROADMAP.md)
- [Discovery and approved decisions](docs/discovery/CODEX_STARTER_KIT_REVIEW.md)
- [Documentation index](docs/README.md)

## How work is organized

All work is tracked in GitHub Issues and the repository's single GitHub Project.

- Feature ideas enter as `type:feature` issues in `Status: Backlog`.
- `Horizon: Now / Next / Later` expresses roadmap intent.
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
what needs qualified human review. This system will support regulatory workflows, but it
will not replace legal, security, accessibility, safety, or domain experts where qualified
judgment is required.

## License

[MIT](LICENSE) © 2026 dragondad22
