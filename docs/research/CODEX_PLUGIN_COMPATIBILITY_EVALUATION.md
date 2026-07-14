# Codex plugin compatibility and distribution evaluation

**Status:** Approved and promoted to DEC-0018  
**Issue:** [#50](https://github.com/dragondad22/codex-starter-kit/issues/50)  
**Freshness:** 2026-07-13  
**Decision:** [DEC-0018](../decisions/DEC-0018-codex-plugin-compatibility-and-distribution.md)

## Objective and stopping conditions

Select the narrowest supportable Phase 2 distribution and compatibility contract for
guided `create`, `status`, and `verify` without making the plugin the conformance
authority. The result must define the development and publication paths, capability
checks, engine/plugin version relationship, offline behavior, fallback, and truthful
limitations needed by issues #51–#54.

The evaluation stops after current official documentation and locally observable Codex
surfaces are sufficient to disposition every Phase 2 roadmap question, define all four
approved capability modes, identify one recommended distribution path, and preserve
explicit uncertainty. It does not implement the plugin, select the public signing
identity, publish to the universal directory, implement repository upgrade, or establish
support for an untested client or native environment.

## Method and provenance

The evaluation used the fresh Codex manual fetched through the official OpenAI
documentation helper, then used the official OpenAI documentation MCP to search and
fetch the exact pages needed to resolve current claims. Local checks inspected only
version and read-only plugin/marketplace commands; they did not install a plugin, add a
marketplace, change workspace policy, or exercise an unapproved external service.

Sources were retrieved on 2026-07-13. Official documentation is primary evidence for
documented product contracts. Local observations establish only this environment and do
not become general support claims.

## Governing constraints

- The plugin is a guided adapter; the standalone engine owns deterministic lifecycle
  behavior and the managed repository owns durable state.
- Plugin installation or update never implies engine, policy, or repository upgrade.
- Direct engine and CI use remain available without Codex or an interactive AI.
- No degraded state may appear as full workflow capability or conformance.
- Network, installation, repository mutation, content handling, and authority changes
  remain separately visible and approved.
- Universal behavior must remain native on Linux, macOS, and Windows without a universal
  Bash, PowerShell, WSL, or interpolated shell-command dependency.
- First-run offline use requires pre-provisioned, verified inputs; it cannot silently
  fetch or execute remote content.

## Current official capability findings

### Plugin and skill shape

The official build contract requires `.codex-plugin/plugin.json`. A plugin may point to
skills, apps/connectors, MCP servers, hooks, and presentation assets, but none of those
optional capabilities is required. Skills use progressive disclosure and are the
documented reusable workflow surface.

The Phase 2 plugin should therefore begin as a **skills-only plugin**. Guided lifecycle
skills call the existing standalone engine through its public executable/JSON contract.
The slice does not need an MCP server, connector/app, browser extension, scheduled task,
or hook. Omitting them reduces trust, authentication, data-flow, network, administration,
and review surface without reducing the required create/status/verify behavior.

### Development distribution

Official documentation supports local, repository, personal, Git-backed, and npm-backed
marketplace entries. The CLI can add, list, refresh, and remove marketplace sources, and
plugin packages are installed into a local cache. Git-backed marketplace entries can pin
a ref or SHA. npm packages introduce an npm client, registry authentication/configuration,
and semver resolution even though lifecycle scripts are not run.

The recommended development path is:

1. keep the plugin and repository marketplace entry in this repository;
2. validate and exercise the local/repository marketplace in development;
3. qualify a Git-backed marketplace snapshot pinned to an immutable source identity for
   team or release-candidate testing; and
4. avoid npm distribution until a separately justified packaging need outweighs its
   additional dependency and resolution surface.

Local and Git marketplace installation is a distribution convenience, not artifact
provenance. A supported workflow still verifies the engine, baseline pack, plugin source,
and compatibility facts it relies on.

A personal marketplace is permitted as an individual developer convenience but is not a
qualification or support surface. Workspace sharing and workspace-installed distribution
are deferred until an authorized workspace, plan, administration policy, and exact
surface behavior can be qualified; Phase 2 team testing uses the pinned Git marketplace
instead. Plugin or marketplace changes require the documented refresh plus a new task,
session, or desktop restart before evaluation. Refreshing a marketplace snapshot does not
authorize repository migration or prove that an installed cached plugin changed. An
administrator-disabled or unavailable plugin selects the safe narrower capability mode;
it is never bypassed by copying or silently enabling the bundle.

### Public publication

The official submission portal accepts skills-only plugins. Public submission requires
an OpenAI Platform organization role with Apps Management write access, a verified
developer or business identity, public listing/support/legal materials, starter prompts,
and exactly five positive plus three negative review cases. Submission begins review;
publication remains a later explicit action after approval.

Phase 2 should produce a publication-ready skills-only package and reusable review cases,
but it should **not publish publicly**. Public identity, legal URLs, exact release
artifacts, signing/provenance, and release approval are Phase 6 authorities. Development
and qualification use the repository/Git marketplace path until those gates exist.

### Surface conflict and disposition

Current official pages conflict about IDE plugin availability:

- the general Plugins page says the CLI and IDE extension can browse and install plugins
  for a Codex environment; while
- the enterprise Plugin controls page says plugin availability on web, desktop, and CLI
  does not make plugins available in the IDE extension.

The conflict prevents a blanket IDE distribution claim. It may reflect different
workspace, administration, rollout, or documentation states, but the source material does
not establish which interpretation is universally correct.

Phase 2 therefore uses **capability evidence, not a client-version threshold**:

| Surface | Phase 2 disposition |
|---|---|
| Codex CLI | Required development surface once its plugin/marketplace commands and installed skill behavior pass the qualification suite |
| ChatGPT desktop Codex | Candidate supported surface; requires manual installation, restart/new-task, routing, approval, and engine-invocation evidence on each claimed OS |
| IDE extension | `needs-review`; installed skills may be observable, but marketplace browsing/installation and administration claims require direct qualification because official pages conflict |
| ChatGPT Work web | Distribution may exist, but the local engine path is unsupported unless a separately verified host/execution route is present |
| Chat/mobile | Unsupported for the Phase 2 local lifecycle workflow |

A version may be recorded as evidence for a passing run, but no version alone proves the
required capabilities, workspace policy, plugin enablement, local engine access, sandbox,
approval behavior, or offline inputs.

## Local observation snapshot

Observed on 2026-07-13:

| Fact | Observation | Claim boundary |
|---|---|---|
| Codex CLI | `codex-cli 0.144.1` | Provides `plugin add/list/marketplace/remove` in this environment only |
| CLI marketplace | `openai-curated` resolves locally; installed plugins can be listed with status, version, and path | Read-only observation; no marketplace or plugin was added for this evaluation |
| VS Code | `1.128.0`, commit `fc3def6774c76082adf699d366f31a557ce5573f`, x64 | Editor host identity only |
| OpenAI extension | `openai.chatgpt@26.707.41301` active; an older `26.707.31428` directory is also present | The active IDE session exposes installed plugin-contributed skills, but this does not prove general IDE marketplace support |
| Operating system | Ubuntu-family Linux, x86_64, kernel `6.17.0-35-generic` | Local research observation, not the native support matrix |
| Git | `2.43.0` | Local observation only |
| Go | Not present on this host after the IDE reload | Direct source build is unavailable here; requiring users to install Go remains a 1.0 release blocker |

The observed CLI version is the same version recorded in the 2026-07-11 discovery spike,
but that repetition still does not establish a minimum supported version.

## Approved compatibility contract

### Capability handshake

Every focused workflow begins with a non-mutating handshake. The handshake records facts
separately rather than collapsing them into a guessed version check:

1. host surface and observable client/plugin capabilities;
2. plugin identity, version, enabled state, and source identity;
3. engine executable resolution from an explicit configured or managed location;
4. engine identity, provenance status, supported protocol/schema range, and available
   operations;
5. repository lifecycle/schema/pin facts when a managed repository exists;
6. baseline pack identity, digest, availability, and compatibility;
7. filesystem, process, sandbox, approval, network, and offline facts relevant to the
   requested workflow; and
8. the requested operation's authority and any content-handling boundary.

The handshake must not mutate the repository, install or update a tool, contact a network
unless separately approved, parse human prose as authority, or convert an unknown fact to
success. Sensitive diagnostics remain redacted.

### Capability modes

The modes describe **workflow capability**, not repository conformance. A repository may
be fully operable through the plugin while verification truthfully returns a non-pass
control state.

| Mode | Required evidence | Permitted behavior |
|---|---|---|
| `full` | Installed/enabled compatible plugin and focused skill; verified compatible engine and baseline; requested operation present; required local permissions and approvals available | Guide and execute the requested supported workflow through the engine, preserving every result and evidence state |
| `degraded-guidance` | The plugin can explain the workflow, but a plugin/client feature, engine identity, baseline, or other required execution capability is missing, incompatible, disabled, or unverified | Read-only guidance, capability report, exact remediation, and direct-engine fallback only; no lifecycle effect |
| `verification-only` | A verified compatible engine can safely perform `status`/`verify`, but mutation is unavailable or unauthorized in the current host, policy, repository, or offline state | Read-only inspection, status, verification, and evidence presentation; no create/apply or migration effect |
| `unsupported` | The handshake cannot establish required guarantees, the environment conflicts with policy, safe fallback is unavailable, or the requested operation itself is not implemented | Stop with the exact failed/unknown guarantees, retained diagnostics where authorized, and bounded remediation |

Modes are evaluated per requested workflow. `verification-only` is not a weaker form of
conformance, and `degraded-guidance` is not permission to simulate an engine result.
Unknown, malformed, or conflicting capability evidence selects `unsupported` unless the
known facts positively establish the narrower `degraded-guidance` boundary.

### Version and upgrade relationships

- Plugin, engine, managed-repository schema, baseline/policy packs, and templates retain
  independent semantic versions and identities.
- The plugin declares supported engine protocol/schema ranges and required operation
  capabilities. It does not import Go packages or treat an executable filename as proof.
- A managed repository retains its own engine/policy/schema pins. Opening it with a newer
  plugin does not change those pins or files.
- An incompatible but verified engine produces an explicit fallback/remediation result;
  the plugin never silently replaces it.
- Plugin update changes guided workflow availability only. Repository migration remains
  the engine's later `upgrade` operation with a reviewed plan and separate approval.
- Exact tested client/plugin/engine/OS identities belong in evidence snapshots, not in a
  timeless claim that a version number guarantees capability.

## Offline and fallback contract

Supported offline use begins only after the plugin marketplace snapshot or cached plugin,
verified engine binary, baseline pack, compatibility metadata, and required trusted roots
are already present. The workflow validates local identities and performs no silent online
fallback. A cache hit without verified identity is not supported offline operation.

If the plugin is missing, disabled, incompatible, or unavailable on a surface, the direct
engine remains the required fallback. If the engine is missing or unverified, the plugin
may explain installation and verification steps but cannot execute lifecycle operations.
CI continues to call the engine directly and never depends on plugin routing or
conversation.

## Trust, authority, data, cost, compatibility, and fallback

- **Trust:** Skills are instructions and must be reviewed as executable workflow input.
  Marketplace location and cache presence do not prove publisher identity or safety.
  Engine and baseline identities require their own verified manifests/evidence.
- **Authority:** Plugin installation adds no repository, network, external-service,
  content-handling, sandbox, approval, or upgrade authority. The active host and user or
  administrator retain those controls.
- **Data:** The recommended skills-only Phase 2 plugin declares no connector, app, MCP
  server, telemetry service, or browser capability. Local engine invocation must not send
  repository content elsewhere. Future additions require separate data-flow review.
- **Cost:** Local/repository marketplace development has no documented OpenAI submission
  fee, but Codex/ChatGPT plan eligibility, workspace administration, native test
  infrastructure, signing, support, and publication operations may carry cost. No paid
  service or public submission is authorized by this decision.
- **Compatibility:** Support is a tested tuple of capabilities, plugin/engine/protocol
  identities, repository/baseline state, host policy, and native environment. Version
  strings are retained evidence, not a substitute.
- **Fallback:** Direct `starter-kit` engine commands are always documented. Restricted or
  older Codex environments remain usable to the extent the independently supported engine
  contract permits.

## Downstream implementation impact

| Issue | Required use of this contract |
|---|---|
| #51 status tracer | Build the minimal skills-only plugin and repository marketplace; implement the handshake and status path without optional integrations |
| #52 guided create | Require `full` create capability, explicit plan/effect approvals, verified baseline availability, and direct-engine fallback |
| #53 guided verify | Permit `full` or `verification-only` verification while preserving all non-pass evidence states |
| #54 qualification | Exercise each mode on every claimed surface/native tuple; retain exact identities, the IDE documentation conflict, and unsupported dispositions |
| #47 operating profiles | May change engagement and evidence presentation only; cannot weaken this capability, authority, or no-false-pass contract |
| Phase 5 upgrade | Supplies verified pack/engine resolution and repository upgrade behavior that Phase 2 must not simulate |
| Phase 6 release | Owns signed/attested binaries, public publication identity/materials, exact release qualification, and publication approval |

## Deferred and excluded choices

- Exact skill prose, manifest fields, plugin directory layout, process adapter, and
  evaluation fixtures belong to #51–#54.
- Exact public publisher identity, legal URLs, support process, pricing commitments,
  signing keys, native installers, and universal-directory publication belong to approved
  release work.
- IDE marketplace support remains `needs-review` until official documentation converges
  or direct qualification proves the exact claimed behavior.
- Apps/connectors, MCP servers, hooks, browser extensions, scheduled tasks, analytics,
  and remote policy/evidence services are excluded until a separately approved use case
  justifies their authority and data surface.
- This decision does not broaden sensitive-data handling, implement retrofit or
  upgrade, or turn a development plugin into a supported 1.0 artifact.

## Invalidation triggers

Return the contract to review when official plugin structure or distribution semantics
change; a required supported surface cannot run the skills-only workflow; capability
probing cannot distinguish a required mode safely; the engine protocol cannot support
the handshake without a breaking change; offline operation requires an undisclosed
network or package-manager dependency; public review requires materially different
packaging; or native qualification reveals semantic drift.

## Primary sources

- [Build plugins](https://learn.chatgpt.com/docs/build-plugins)
- [Build skills](https://learn.chatgpt.com/docs/build-skills)
- [Plugins](https://learn.chatgpt.com/docs/plugins)
- [Plugin controls](https://learn.chatgpt.com/docs/enterprise/apps-and-connectors)
- [Skill controls](https://learn.chatgpt.com/docs/enterprise/skills)
- [Submit plugins](https://learn.chatgpt.com/docs/submit-plugins)
- [Agent approvals and security](https://learn.chatgpt.com/docs/agent-approvals-security)
- [Managed configuration](https://learn.chatgpt.com/docs/enterprise/managed-configuration)
