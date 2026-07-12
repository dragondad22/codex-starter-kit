# Codex Starter Kit — review and design discussion

**Status:** Discovery; product direction clarified, implementation decisions remain open  
**Reviewed:** 2026-07-11  
**Source project:** `/mnt/DATA/source/claude-starter-kit` at `0d9c5f7` (version 0.8.0)

## Why this document exists

We want a Codex-native development system that guides a person through an
AI-assisted development workflow, records decisions in durable human-facing
documents, and applies agreed standards and preferences without making the developer
manually coordinate all of that complexity.

“Easy to use” does **not** mean a lightweight product, a minimal process, or a
single-prompt application generator. It means a very complex system presenting a
coherent guided interface. Starting from an idea or an existing repository, it should
produce and maintain what a responsible developer needs: implementation, tests,
security controls, compliance evidence, technical documentation, decision records,
and communication suitable for non-developers. The developer should not have to
remember every obligation or wonder whether the AI chose an unsafe shortcut such as
committing credentials.

This is a critique and a decision workspace, not an implementation plan. Statements
marked **Finding** describe observed evidence. Statements marked **Proposal** are
starting positions to discuss. Only items moved to **Agreed** should constrain the
build.

### Standing capability rule

**Agreed (Chris, 2026-07-11):** Upgrading existing tools and installing additional
tools, skills, plugins, integrations, policy packs, or supporting services are valid
solution options. The user is not expected to know the available ecosystem in advance.
When a capability gap, reliability problem, or materially better workflow could be
addressed by an upgrade or installation, the system should identify and present those
options rather than silently constraining the design to what is already installed.

Each recommendation should explain the capability gained, why it is relevant, material
tradeoffs, trust/source, permissions or data access introduced, licensing/cost where
known, compatibility, and a no-install fallback when one reasonably exists. Discovery
and recommendation do not imply silent installation: changes that broaden authority,
send data externally, add executable dependencies, or alter shared environments remain
explicit and reviewable. Installed capabilities and versions become part of project
provenance and upgrade monitoring.

## Executive assessment

The Claude kit has an unusually thoughtful process model. Its strongest ideas are
worth preserving: brief-first inception, asynchronous interviews, explicit decision
provenance, a small always-loaded instruction file, progressive modules, audience-first
documentation, issue-backed work, and verification before release.

It should not be ported by renaming `CLAUDE.md` and `.claude/`. Its workflow engine is
mostly natural-language slash commands, and several important guarantees exist only as
instructions. A Codex version should keep the domain model while replacing the delivery
mechanism with a smaller Codex-native interface: `AGENTS.md` for durable repository
guidance, skills for user-invoked workflows, deterministic scripts for mechanical work,
and project configuration/hooks only where enforcement is both safe and useful.

The target is therefore a **deep module**: a small, understandable developer interface
with substantial policy, orchestration, validation, and evidence generation behind it.
Complexity is expected in the implementation. Complexity leaking through the interface
as dozens of choices, manual reminders, or disconnected commands is the failure mode.

The agreed universal baseline includes secrets protection, secure defaults,
threat-aware design, testing integrity, dependency hygiene, documentation currency,
decision provenance, change traceability, and truthful verification. A required control
without sufficient evidence cannot be reported as passing.

The main risks to resolve before building are:

1. The product’s required complexity currently leaks into the developer’s workflow.
2. The bootstrap path is not deterministic, transactional, or meaningfully tested.
3. The shipped security review can report success while skipping material checks.
4. Some shipped shell behavior contradicts the stated macOS portability guarantee.
5. A direct port would encode Claude-specific assumptions instead of using Codex’s
   actual instruction and extension surfaces.

## What is already strong

### Preserve the workflow concepts

- **Brief before questionnaire.** The user approves a plain-language understanding of
  the project before the system narrows the design space.
- **Asynchronous, resumable discovery.** Questions and answers live in files and survive
  context loss.
- **Decision provenance.** Answers can identify the documents derived from them, and
  durable decisions are distinguished from chat.
- **Progressive disclosure.** The root instruction file is kept small while detailed
  standards are loaded only when relevant.
- **Additive modules.** Database, UI, reports, deployment, and SLA material appear when
  a real trigger exists.
- **Audience-first writing.** User-facing material is separated from agent-facing
  standards and internal evidence.
- **A visible journey.** `docs/kit/WORKFLOW.md` directly answers “where am I?” and “what
  comes next?”
- **Self-validation.** The manifest, cross-reference lint, and bootstrap smoke test all
  pass in the source checkout.

### Preserve the separation of concerns

The source kit correctly separates shipped template content from kit-development
content using an allowlisted manifest. That is a useful module seam: the scaffold
interface is “install a declared kit version and selected capabilities,” while the
template layout and development records remain implementation details.

## Findings and proposals

### F1 — The Claude workflow interface does not map directly to Codex

**Severity:** Foundational  
**Finding:** The source product exposes ten workflows through
`.claude/commands/*.md`, loads `CLAUDE.md`, and stores permissions in
`.claude/settings.json`. Those are Claude Code interfaces. Codex discovers repository
instructions through `AGENTS.md` (including nested overrides), and its reusable
workflow interface is a skill. Codex also limits the combined project instruction
chain to 32 KiB by default, reinforcing the source kit’s instinct to keep root guidance
small.

Official references:

- [Custom instructions with AGENTS.md](https://developers.openai.com/codex/guides/agents-md)
- [Build skills](https://developers.openai.com/codex/skills)
- [Advanced configuration](https://developers.openai.com/codex/config-advanced)
- [Build plugins](https://developers.openai.com/codex/plugins/build)

**Clarified requirement (Chris, 2026-07-11):** The product must support all three
lifecycle paths from day one:

1. Scaffold a new project from an idea or empty repository.
2. Inspect and retrofit an existing project into the standard structure and workflow.
3. Apply future kit corrections, standards changes, and capability upgrades safely to
   repositories already using it.

It is a template **and** an ongoing management system; “template first, functionality
later” is not acceptable.

**Revised proposal:** Treat the repository template as one adapter behind a lifecycle
module, not as the product itself. Use these seams:

| Concern | Codex-native home |
|---|---|
| Durable repository rules and command index | `AGENTS.md` |
| Area-specific rules | nested `AGENTS.md` only when that subtree genuinely differs |
| Guided workflows such as bootstrap, preflight, release | project skills |
| Deterministic transformations and validation | scripts called by skills |
| Trusted-repository settings | `.codex/config.toml`, kept minimal |
| Mechanical policy enforcement | hooks/rules, opt-in and narrowly scoped |
| Installable lifecycle/orchestration surface | plugin from day one, if current Codex plugin capabilities can own installation and upgrades |
| Scaffolded repository content | versioned template and policy manifests behind the lifecycle module |

Avoid a compatibility-shaped `.codex/commands/` clone. It would be a shallow module:
lots of prompt surface, little reliable behavior.

**Decision needed:** Which day-one distribution shape can provide one guided experience
while supporting create, retrofit, verify, and upgrade? The working direction is an
installable Codex plugin containing skills plus deterministic lifecycle tooling, with a
versioned template as an internal adapter. This still needs a capability spike against
current Codex plugin behavior before it becomes an approved architecture.

### F2 — Required complexity leaks through the interface

**Severity:** High  
**Finding:** Core installation ships 72 allowlisted files. The main instruction,
commands, standards, bootstrap material, and kit guide total about 2,950 lines. A text
sweep finds 131 uses of strong policy words such as “must,” “never,” and “required.”
The workflow also assumes a board, typed issues, feature specifications, ADRs,
compliance tracking, changelog discipline, multiple review modes, and periodic process
maintenance.

The problem is not that these responsibilities exist. The product is intended to own
them. The problem is that the developer currently has to understand and manually
navigate much of the machinery: which standards apply, which review command to run,
which artifact to create, whether a compliance trigger fired, and whether skipped
checks made a green result meaningless.

**Clarified requirement (Chris, 2026-07-11):** Standards, tests, security,
documentation, compliance, evidence, and non-developer communication are safe defaults,
not optional sophistication reserved for a “Governed” tier. Scalability is required.

**Revised proposal:** Replace scope-reducing profiles with **context-derived policy**.
The system asks or detects facts about the project—data handled, users, jurisdictions,
deployment model, team shape, risk, interfaces, and lifecycle—and compiles those facts
into an explicit obligation set. Universal engineering controls always apply; additional
controls activate when their evidence-backed triggers fire.

The developer interface should normally be stage-oriented rather than control-oriented:

```text
Understand → Plan → Build → Verify → Communicate → Release → Maintain
```

Within `Verify`, the system selects and orchestrates testing, security, compliance,
documentation, and evidence checks. Expert workflows can expose individual checks for
diagnosis, but the developer should not need to remember the full matrix for ordinary
work.

**Decision needed:** What is the universal baseline, and which controls may legitimately
be trigger-derived? “Small project” must not mean “unsafe project,” but not every project
needs the same regulatory evidence or operational machinery.

### F3 — Bootstrap is a prompt, not a dependable engine

**Severity:** High  
**Finding:** `.claude/commands/bootstrap.md` tells the agent to create an interview,
replace tokens throughout the repository, generate licenses and founding documents,
configure permissions, alter tracker state, and verify the result. There is no schema
for interview state, no deterministic generator, no transaction/rollback, and no
golden-output or idempotence test of the agent-authored workflow. The passing smoke
test substitutes dummy values for tokens; it does not exercise the actual inception
state machine or generated documents.

The scaffold script also skips existing files one by one. A rerun into an existing
repository can therefore create a hybrid installation without producing a conflict
plan or machine-readable result.

**Proposal:** Build a deep lifecycle module with a small interface used by both skills
and automation:

```text
kit create <repo>                  # idea/empty-repo inception path
kit retrofit <repo>                # inspect existing repo and propose conformance
kit inspect <repo>                 # read-only facts, risks, obligations, conflicts
kit plan <repo> <answers-file>     # exact writes, controls, modules, migrations
kit apply <plan-file>              # transactional or fail-before-write
kit verify <repo>                  # structural, behavioral, policy, evidence invariants
kit status <repo>                  # lifecycle state, kit version, drift, incomplete setup
kit upgrade <repo>                 # classify and apply upstream kit changes safely
```

Skills should conduct the human conversation and update a versioned answer document;
the module should own validation, rendering, file conflict policy, and idempotence.
Tests should cover fresh, existing, interrupted, rerun, and upgrade cases.

**Decision needed:** Is YAML/JSON acceptable as the machine-readable interview state
with Markdown rendered for the person, or must Markdown itself be the source of truth?

### F4 — Security review has false-green paths

**Severity:** Critical for the shipped claim  
**Finding:** `security-review.sh` says it returns nonzero when a real finding is
detected, but:

- an unconfigured dependency/SAST command is recorded as skipped and leaves success;
- targeted security tests are a TODO and leave success;
- secret/risky-pattern matches are logged but never change the exit status;
- only tracked unstaged/staged diffs are searched; untracked files are omitted;
- user-controlled `ID` and `SLUG` are placed in an artifact path without safe character
  validation, permitting path traversal;
- `SECURITY_SCAN_COMMAND` is interpolated into `bash -lc`, which intentionally enables
  arbitrary shell execution without a trust boundary or structured command form.

Evidence: source `template/core/ai/scripts/security-review.sh`, especially lines 17–19,
28–35, 55–66, 83–95, 97–126, and 128–145.

**Proposal:** Do not ship a generic “security passed” stub. Return one of `pass`, `fail`,
or `not-configured`; derive required checks from the universal baseline and triggered
obligations; validate artifact identifiers;
include untracked files when appropriate; and represent commands as checked argument
arrays or an explicitly trusted project script path. A heuristic finding should either
fail or produce an explicit `needs-review` status that cannot be summarized as pass.

**Decision needed:** Which security controls belong to the universal fail-closed
baseline, and which may remain a prominent incomplete-setup state until a detected
project trigger makes them mandatory?

### F5 — The portability claim is not actually covered

**Severity:** High  
**Finding:** The kit says shipped scripts are compatible with stock macOS/BSD tools.
`lib/redact.sh` uses GNU-style `sed -E -i` without a backup suffix and suppresses any
failure with `|| true`. BSD `sed` requires a suffix argument for `-i`. The CI matrix
does run macOS, but the smoke test does not execute the redaction path, so syntax checks
and smoke tests remain green.

Evidence: `template/core/ai/scripts/lib/redact.sh:21-27` and
`.github/workflows/kit-selftest.yml:27-47`.

**Proposal:** Prefer a small Python redactor with fixture tests for both redaction and
non-redaction cases, or use a portable temp-file-and-move implementation. Add executable
tests for every shipped script, not only `bash -n`.

### F6 — Supply-chain and CI hygiene need tightening

**Severity:** Medium  
**Finding:** CI installs the latest unconstrained `pyyaml` at runtime and uses GitHub
Actions by a mutable major tag (`actions/checkout@v4`). This is common but does not match
a hardened starter kit’s likely standards. The project does not currently run a secret
scanner, dependency review, provenance check, or static analysis over its own scripts.

**Proposal:** Pin Python dependencies with hashes or eliminate PyYAML from the validation
path; pin Actions by commit SHA with a dependency updater; add ShellCheck to CI; and add
a lightweight secret scan. Document which controls are universal versus
context-triggered so security does not become theater.

### F7 — Configuration and permissions should be least-authority

**Severity:** High  
**Finding:** The Claude settings allow broad command patterns such as all
`ai/scripts/*`. Bootstrap then offers to expand detected permissions. A direct Codex
translation risks committing permissive settings into every downstream repository.

**Proposal:** Ship no blanket approval policy. Let Codex’s sandbox and approval model
remain the outer control. Project configuration may declare safe defaults, but workflows
that mutate issue trackers, release state, credentials, remote repositories, or external
systems must remain explicit. Hooks should block a small set of mechanically dangerous
operations, not act as a substitute for review.

### F8 — Human-facing decisions and agent instructions are partially entangled

**Severity:** Medium  
**Finding:** The source kit has excellent human artifacts, but many decision semantics
live inside agent-facing standards and command prose. Conversely, `docs/plans/` is called
“structured discovery only” while also carrying the kit’s authoritative decision record.
That makes document purpose depend on context rather than interface.

**Proposal:** Use a simple artifact model:

- `docs/product/brief.md` — shared understanding of the product.
- `docs/decisions/` — approved product and architecture decisions, with status/history.
- `docs/specs/` — behavior to build, tied to user journeys and acceptance criteria.
- `docs/work/` only if no external tracker is selected.
- `.codex/skills/` and `AGENTS.md` — agent-facing workflow and rules.

Every generated document should declare its audience, authority, lifecycle, and source
inputs in a small front matter block. The person should not need to understand the
agent’s implementation structure to find an agreed decision.

### F9 — GitHub is treated as universal despite “stack-agnostic” positioning

**Severity:** Medium  
**Finding:** The source workflow requires one project board per repository, a Horizon
field, typed labels, branch-per-issue naming, and `gh`-driven setup. Other trackers are
mentioned, but GitHub behavior shapes the core documents and bootstrap flow.

**Superseded proposal (D5, 2026-07-11):** Work tracking is not optional for this product.
Every managed repository requires GitHub Issues and exactly one linked GitHub Project.
GitHub remains an adapter seam in the implementation so policy and lifecycle logic are
testable and a future platform adapter is possible, but the first-release product
contract does not offer `none`, file-only, or manual tracking modes.

### F10 — Upgrades and provenance are day-one core functionality

**Severity:** High  
**Finding:** The scaffold records a kit version and stages module copies, but skips an
existing marker and existing destination files. It does not retain a content manifest
with hashes, classify local modifications, or produce a three-way upgrade plan. “Never
regenerate an artifact that diverged by hand” is sound policy but currently depends on
agent judgment.

**Clarified requirement (Chris, 2026-07-11):** Repositories must receive later fixes and
standards improvements. Upgrade behavior cannot be postponed until after the initial
template stabilizes.

**Proposal:** Record, per managed file, the source kit version and content hash. Upgrade
should classify files as unchanged, user-modified, removed upstream, or conflicting;
automatically update only unchanged managed files and render a reviewable migration plan
for the rest. User-owned documents should never be silently regenerated. The first
release format must include migration/version semantics and upgrade fixtures, even if
only one version exists initially.

<a id="d11"></a>
### F11 — Breadcrumbs should be a governed context-loading interface

**Severity:** Foundational for context economy  
**Clarified requirement (Chris, 2026-07-11):** Agent-loaded documents should remain
small. They should contain links or references to additional standards, decisions, and
supporting material so the AI follows them only when relevant. This reduces routine
token use, but stale or frequently rewritten references can create correctness risk and
maintenance overhead.

**Assessment:** The breadcrumb model is sound. It is how the system can remain deep
without placing the entire implementation of its policy into every prompt. The mistake
would be treating breadcrumbs as informal Markdown links maintained by memory. They are
part of the instruction module's interface: a broken or misleading breadcrumb changes
what the agent can discover and therefore changes behavior.

**Proposal:** Use a layered context graph with stable identifiers and mechanically
verified edges:

1. **Orientation layer:** `AGENTS.md` contains only universal rules, the current workflow
   stage, and a compact routing table telling Codex what to load for a class of work.
2. **Workflow layer:** a selected skill owns the procedure for the current activity and
   names the policy topics/evidence it requires.
3. **Policy layer:** focused standards describe one concern and link to applicable
   decisions, templates, and verification controls.
4. **Evidence layer:** decisions, specifications, compliance records, test evidence, and
   reports are loaded only when a workflow or policy requires them.

Prefer stable semantic references over path-heavy prose:

```text
SEC-SECRETS        → secret handling and repository scanning policy
TEST-INTEGRITY     → testing integrity policy
DEC-0042           → approved decision record
CTRL-ACCESS-001    → access-control verification obligation
```

A small generated registry maps each ID to its current path, title, lifecycle state,
audience, and short “load when” description. Documents refer to the stable ID and may
include a human-clickable link generated from the registry. Moving or renaming a file
then changes one registry entry or regenerated links rather than requiring semantic
edits throughout the repository.

The lifecycle module should own these invariants:

- every referenced ID exists and resolves to one authoritative artifact;
- duplicate IDs, cycles that imply mandatory recursive loading, and orphaned required
  artifacts fail validation;
- superseded decisions resolve to their replacement while preserving history;
- links and routing descriptions are regenerated or linted during relevant changes;
- the root routing table stays within an explicit size/token budget;
- a policy change declares which routes, controls, templates, and generated summaries
  it affects;
- verification tests exercise routing—for representative tasks, the expected context is
  discoverable without loading unrelated policy.

Breadcrumbs should be **one-way by default**: orientation → workflow → policy → evidence.
Backlinks and “used by” lists should be generated for humans and impact analysis, not
copied into agent-loaded prose. This limits cycles and reduces update churn.

**Token tradeoff:** Maintaining the graph consumes tokens during changes to policy or
structure, but that is bounded, relevant work. It avoids paying the much larger recurring
cost of loading all standards on every task. Mechanical registry/link updates should be
performed by tooling so the AI spends reasoning tokens on semantic impact, not pathname
replacement.

**Agreed (Chris, 2026-07-11):** Stable IDs are required for governed policy, controls,
decisions, and specifications. Ordinary explanatory pages may use normal relative links
unless another artifact depends on them as an authoritative instruction. Breadcrumbs
and their registry are mechanically validated as part of the context-loading interface.

## Proposed product shape

This is a starting architecture, not yet a commitment. It now treats scaffolding,
retrofit, enforcement, and upgrades as one product from day one.

```text
codex-starter-kit/
├── AGENTS.md                    # concise rules for developing the kit itself
├── README.md
├── docs/
│   ├── discovery/              # this discussion and unresolved choices
│   ├── decisions/              # approved decisions for the kit
│   └── architecture.md
├── kit/                        # deterministic scaffold/inspect/apply/verify module
├── policy/                     # universal baseline + context-triggered obligations
│   ├── registry.yml            # stable IDs, locations, routing metadata
│   └── controls/               # focused standards and verifiable obligations
├── template/
│   ├── core/                   # smallest downstream repository surface
│   └── modules/                # tracker, ui, db, compliance, release, uat...
├── plugin/                     # installable Codex surface: skills + lifecycle tooling
├── skills/                     # source definitions for guided Codex workflows
└── tests/                      # golden, security, portability, upgrade, idempotence
```

The downstream repository should present a much smaller interface:

```text
AGENTS.md
.codex/skills/<selected workflows>/SKILL.md
docs/product/brief.md
docs/decisions/
docs/specs/                     # created and maintained as the delivery workflow requires
.starter-kit/state.json         # generated provenance; not human-authored
```

Detailed standards can live in versioned policy modules and skill references instead of
being duplicated into every repository. Each downstream repository still receives the
human-readable policy, decision, and evidence surfaces needed to prove what applies and
why. Project-specific deviations and accepted risks must be recorded locally.

<a id="d1"></a>
## First-release operating contract

**Agreed audience and scope (Chris, 2026-07-11):**

- The system supports all principal software-repository/project areas from the first
  release: applications, libraries, infrastructure, data projects, and documentation
  repositories. Project facts select relevant workflows and controls; the architecture
  must not assume that every repository produces a deployable application.
- The initial experience may optimize for one primary developer, while preserving
  ownership, approval, separation-of-duty, and communication concepts needed by teams.
- GitHub is the initially supported collaboration platform. Work tracking remains an
  adapter seam so core policy and evidence are not inseparable from GitHub-specific
  boards or fields.
- Regulated projects must be genuinely supported in the first release. Regulatory
  policy, applicability decisions, required evidence, exceptions, and verification are
  part of the launch contract—not merely future extension points.

This breadth makes project classification and policy compilation core functionality.
The system must distinguish “not applicable” from “not checked” and retain the evidence
behind either result.

<a id="d12"></a>
### Managed-repository conformance contract

**Agreed (Chris, 2026-07-11):** A repository managed by the kit guarantees:

1. Project context and applicable obligations are explicitly classified.
2. Every applicable control is identified by a stable ID and version.
3. Each control has an explicit state: `pass`, `fail`, `not-applicable`,
   `not-configured`, `needs-review`, or `accepted-exception`.
4. `pass` always points to current evidence; absence of evidence cannot produce green
   status.
5. `not-applicable` records the project facts and rule supporting that determination.
6. An accepted exception records the risk, rationale, approver, scope, expiration, and
   compensating controls.
7. Secrets protection and other universal safeguards fail closed.
8. Required tests, documentation, decisions, and stakeholder communications are derived
   from the change and project context.
9. Generated and user-owned files have explicit provenance, ownership, and safe upgrade
   behavior.
10. Every release produces a human-readable conformance summary and a machine-readable
    evidence manifest.
11. Retrofit reports existing violations without silently rewriting user work.
12. Upgrades classify changes, preserve local decisions, and never silently weaken
    controls.
13. The system reports coverage limits: what it checked, could not check, and requires
    human verification.
14. Conformance is reproducible: another authorized developer or CI run can evaluate
    the same repository against the same policy version.

Risk management is therefore a first-class product concern, not only a regulatory
feature. Real projects sometimes cannot satisfy a control immediately. The system must
make that risk visible, owned, time-bounded, and reviewable rather than allowing a false
pass or an undocumented permanent bypass. Solo projects may use the same person as owner
and approver, but the record must make that lack of separation explicit; team and
regulated contexts may require independent approval through triggered policy.

<a id="d2"></a>
## D2 workshop — universal and context-triggered controls

**Status:** **Agreed (Chris, 2026-07-11)**  
**Goal:** Define how the policy compiler decides what applies without making a small
project unsafe or forcing irrelevant regulatory and operational controls onto every
repository.

### Proposed classification model

Each governed control has four independent classifications:

1. **Applicability:** `universal` or `triggered`.
2. **Evaluation:** `automated`, `human-attested`, or `hybrid`.
3. **Enforcement:** `advisory`, `blocking`, or `release-blocking`.
4. **Exception policy:** `allowed`, `independent-approval-required`, or `prohibited`.

These dimensions must not be collapsed into a single severity. For example, a privacy
impact assessment may be triggered, human-attested, release-blocking, and require an
independent approver for any exception. A secret scan is universal, automated, blocking,
and cannot permit a known live credential to be committed.

### Proposed universal baseline

Universal controls are mostly **trust controls**: they make every later claim reliable.
They apply to applications, libraries, infrastructure, data, and documentation projects,
although their evidence can differ by project type.

| Area | Universal guarantee | Normal evidence | Initial exception policy |
|---|---|---|---|
| Classification | Project type, lifecycle, data, users, deployment, and regulatory facts are recorded and reviewed | Versioned project profile | Allowed temporarily as `needs-review`; never inferred as complete |
| Truthful status | No check is represented as passing without sufficient current evidence | Evidence manifest + evaluator result | Prohibited |
| Secrets | No known live credential, private key, or sensitive token enters managed source/history/artifacts | Pre-write scan, staged scan, history/release scan as appropriate | Prohibited for committing a known live secret |
| Least authority | AI, scripts, CI, and integrations receive only the authority needed for the operation | Configuration inspection + approval/action log | Allowed only with scope, expiry, and rationale; legal/admin limits still apply |
| Change traceability | Material changes identify intent, owner, affected artifacts, verification, and resulting decisions | Work item/commit/PR/evidence links | Allowed only for declared emergency flow with retrospective deadline |
| Verification integrity | Required verification actually runs against the relevant change; skipped, flaky, stale, or altered tests are visible | Command provenance, results, test-change analysis | Prohibited from being reported as pass; risk acceptance may permit release only where policy allows |
| Secure defaults | Generated configuration and examples avoid known-insecure defaults and fail safely when required values are absent | Static rules + rendered-output tests | Known insecure production defaults prohibited |
| Dependency/provenance hygiene | External code, actions, tools, models, standards, and generated inputs have identifiable sources and versions | Lockfiles/manifests/SBOM or equivalent provenance record | Allowed only where pinning is impossible and risk is recorded |
| Documentation currency | User-visible behavior, operating instructions, applicable controls, and decisions change with the implementation they describe | Change-impact mapping + link/route validation | May be time-bounded for non-safety-critical prose; false instructions cannot ship knowingly |
| Decision provenance | Material product, architecture, policy, and accepted-risk decisions are durable, attributable, and supersedable | Decision record with stable ID | Emergency retrospective allowed; silent permanent decisions prohibited |
| Artifact ownership | Generated, managed, and user-owned files are distinguishable; automation does not silently overwrite user work | Managed-file manifest + hashes + plan | Prohibited |
| Coverage disclosure | Results state what was checked, not checked, not applicable, and requires human review | Human summary + machine manifest | Prohibited |
| Recovery | Mutating lifecycle operations have a preview, conflict policy, and recoverable/transactional behavior | Plan, backup/transaction record, rollback test | Destructive unrecoverable mutation prohibited without explicit exceptional authority |
| Licensing/IP hygiene | Repository content and dependencies have identifiable licensing/provenance; generated content does not knowingly reproduce restricted material | License inventory + source record | Applicable law and license terms cannot be waived by the kit |
| Breadcrumb integrity | Governed references resolve, routing stays within budget, and relevant context remains discoverable | Registry lint + routing tests | Prohibited from claiming conformance with broken required routes |

“Universal” does not mean every row runs the same tool. A documentation repository may
prove verification through link checks, rendering tests, editorial review, and provenance;
an application may require unit, integration, security, and end-to-end tests. The
universal requirement is that the project defines and truthfully executes the relevant
verification contract.

### Proposed trigger families

Triggered controls add obligations; they never remove the universal baseline.

| Trigger family | Example facts that activate it | Example additional controls |
|---|---|---|
| Deployable runtime | Service, application, job, container, package execution | Environment separation, observability, rollback, vulnerability response, runtime hardening |
| Public or consumed interface | API, CLI, library surface, file format, webhook | Compatibility/versioning, contract tests, input validation, abuse limits, consumer documentation |
| Authentication/authorization | Accounts, roles, sessions, privileged operations | Threat model, access-control matrix, negative authorization tests, session/credential handling |
| Sensitive or personal data | PII, credentials, health, financial, precise location, confidential business data | Data inventory, classification, minimization, encryption, retention/deletion, access/audit evidence |
| Regulated context | Jurisdiction, industry, contract, customer requirement, certification target | Applicable control pack, responsibility mapping, evidence retention, independent review, required notices/records |
| User interface | Web, mobile, desktop, terminal UI used by people | Accessibility, audience/persona review, error-message safety, usability and journey validation |
| Infrastructure/change to environments | IaC, cloud resources, networks, secrets stores, production settings | Plan review, policy-as-code, drift detection, least privilege, state protection, rollback |
| Data pipeline/modeling | Collection, ETL, analytics, reporting, ML datasets | Lineage, data-quality tests, reproducibility, bias/fitness checks where applicable, retention |
| AI-enabled behavior | Model calls, agents, retrieval, generated decisions/content | Prompt-injection boundaries, tool authority, evaluation, provenance, human oversight, model/data risk |
| Payments/financial effects | Charges, payouts, billing, financial calculations | Integrity/reconciliation, authorization, idempotency, audit trail, applicable payment/financial controls |
| Minors or vulnerable users | Intended or reasonably foreseeable use | Age/audience safeguards, consent, content/safety review, stricter data and communication controls |
| Safety-critical or high-impact decisions | Physical safety, medical, employment, housing, credit, legal or similar impact | Hazard analysis, human review, independent validation, stricter release and incident controls |
| Third-party integration | External API, SaaS, connector, plugin, marketplace package | Data-flow review, scoped credentials, vendor/provenance assessment, failure/revocation handling |
| Team/contributor scale | Multiple contributors, external contributors, production operators | Ownership, review independence, branch protection, separation of duties, onboarding/offboarding |
| Distribution/publication | Open source, package registry, app store, customer delivery, published docs | License/notices, release signing/provenance, support/security policy, audience-facing release communication |

The project profile stores facts, not legal conclusions. A versioned rule set maps those
facts to policy packs and records why each pack applied. Where jurisdiction or regulatory
interpretation is uncertain, the state is `needs-review`; the AI must not invent legal
certainty.

### Proposed exception rules

Risk acceptance changes the disposition of a failed control; it does not rewrite the
evaluation result. The underlying control remains failed or incomplete, alongside a
separate accepted-risk record.

An exception cannot:

- turn absent evidence into `pass`;
- waive applicable law, a binding contract, license terms, or a platform requirement;
- authorize committing a known live secret;
- conceal scope or coverage limitations;
- silently weaken policy during an upgrade;
- be permanent by default.

Every exception requires an owner, rationale, affected assets/controls, impact and
likelihood, compensating controls, approval rule, creation date, expiration/review date,
and closure criteria. Expired exceptions become blocking unless renewed through the same
review process. Emergency exceptions are short-lived and automatically create follow-up
work and a retrospective obligation.

### D2 decision

**Approved (Chris, 2026-07-11):**

1. The proposed universal baseline is accepted as the initial policy skeleton. Concrete
   policy-pack design may propose additions, but may not silently remove or weaken it.
2. A **corrective exception** is time-limited and exists for a condition expected to be
   remediated. An **accepted residual risk** may describe risk inherent in a chosen
   architecture and use periodic review rather than a forced remediation date. Its
   acceptance still expires unless reviewed on schedule. A **prohibited exception**
   cannot authorize release or produce conformance.
3. Self-approval does not satisfy a control requiring independent approval. A solo
   developer may record the risk and continue work where the applicable policy permits,
   but the repository cannot claim the corresponding regulatory or high-impact
   conformance until an eligible independent reviewer approves it.
4. Documentation or verification debt may permit a release only when the affected
   control allows exceptions, missing evidence is disclosed, the release is not
   represented as conformant for that control, no legal/contractual/safety/platform
   obligation forbids it, and the exception has compensating controls plus a deadline.

<a id="d3"></a>
## D3 workshop — day-one distribution and lifecycle architecture

**Status:** **Agreed (Chris, 2026-07-11)**  
**Goal:** Provide one guided create/retrofit/verify/upgrade experience from day one
without making conformance depend on a single Codex client surface.

### Capability-spike findings

Current official Codex documentation establishes that:

- plugins are stable distributable packages for shared workflows and may bundle skills,
  MCP/app configuration, lifecycle hooks, and other plugin resources;
- skills are available across the desktop app, CLI, and IDE extension, may contain
  executable scripts/references/assets, and use progressive disclosure;
- marketplaces support personal, repository, and workspace distribution;
- plugin changes/installations are normally picked up at a restart or new-task boundary;
- plugin sharing and availability can be restricted by workspace administration.

Official references:

- [Build plugins](https://developers.openai.com/codex/plugins/build)
- [Build skills](https://developers.openai.com/codex/skills)
- [Advanced configuration](https://developers.openai.com/codex/config-advanced)

The initial local capability check on 2026-07-11 found `codex-cli 0.104.0` with no
`codex plugin` command. After upgrading, the terminal resolves to `codex-cli 0.144.1`
and exposes `plugin add/list/marketplace/remove`; the VS Code extension bundles
`0.144.0-alpha.4` and also exposes plugin management. The immediate limitation is
resolved, but the observed version gap remains evidence of a real compatibility
condition: the conformance engine cannot exist only as a plugin or assume every
installed Codex client has current plugin management.

### Proposed three-layer distribution model

```text
Codex plugin / project skill
        │ guided conversation, routing, approvals
        ▼
Standalone lifecycle engine
        │ deterministic create/retrofit/plan/apply/verify/upgrade
        ▼
Managed repository
        policy state, human records, generated views, evidence, provenance
```

#### Layer 1 — Codex experience adapter

The installable plugin is the preferred user experience. It contains focused skills,
policy references needed for reasoning, lifecycle hooks where they are safe, assets, and
the metadata necessary for personal/workspace distribution. It translates user intent
into lifecycle-engine operations, explains plans, obtains approvals, and presents
evidence.

It is an **adapter**, not the sole enforcement authority. Skills can be invoked
implicitly or explicitly, and agent behavior is probabilistic; therefore a skill alone
cannot prove conformance. Plugin absence or administrative disablement must be detectable
and must not cause a managed repository to appear conformant.

#### Layer 2 — Standalone lifecycle engine

A versioned, cross-platform executable/module owns schemas, policy compilation,
rendering, file provenance, migrations, conflict handling, evidence manifests, and
verification. It provides the agreed lifecycle interface:

```text
create · retrofit · inspect · plan · apply · verify · status · upgrade
```

The same engine is callable by Codex skills, developers, CI, and future integrations.
This makes conformance reproducible outside an AI conversation and lets older or
restricted Codex environments use a documented fallback path.

The engine distribution must provide checksums/signatures, a supported-platform matrix,
version pinning, an offline verification mode, and a bootstrap installer that never
executes unverified remote content. The implementation language/package mechanism is a
separate decision; D3 establishes that the engine cannot depend on the plugin runtime.

#### Layer 3 — Managed repository contract

The repository contains the minimum durable local state required to understand and
reproduce its conformance: pinned engine/policy versions, structured project state,
managed-file provenance, human-owned decisions and risks, generated views, routing
registry, and evidence manifests. It does not depend on a particular developer’s global
Codex installation to explain what applies.

### Installation and upgrade paths

| Situation | Preferred path | Required fallback |
|---|---|---|
| New/empty project | Install plugin, run guided `create` | Install verified engine, run `create`, then add project skill/plugin when available |
| Existing unmanaged project | Plugin-guided `retrofit` with read-only inspection first | Engine `inspect` + `plan`; no writes before approval |
| Managed project opened without plugin | Plugin/skill availability check explains install | Engine `status` and `verify` remain fully functional |
| Kit update | Update plugin/engine catalog, generate repository upgrade plan | Pinned old version remains reproducible; verified manual engine update path |
| CI | Pinned lifecycle engine verifies policy/evidence | CI must not depend on an interactive plugin or AI judgment |
| Restricted/offline environment | Pre-approved mirrored plugin/engine/policy pack | Offline verification against pinned artifacts; no silent online fallback |

Plugin updates and repository upgrades are separate operations. Updating the plugin
changes the available user experience and bundled catalog; it must not silently migrate
repositories. `upgrade` reads the repository’s pinned state, proposes version-by-version
migrations, shows policy changes and weakened/strengthened controls, and applies only
after the required approval.

### MCP, hooks, and configuration

- **MCP/app:** not required for the core lifecycle. Add it only when a real second adapter
  is needed—for example, a rich conformance dashboard or remote policy/evidence service.
  GitHub access can use the available connector/CLI adapter without making core policy
  dependent on it.
- **Hooks:** useful for narrow fail-fast controls such as secret detection or blocking
  known destructive lifecycle bypasses. Hooks supplement CI and engine verification;
  they are not the only enforcement layer and must not perform surprising writes.
- **Project configuration:** keep minimal and version-aware. Never weaken a user or
  administrator sandbox/approval policy as part of installation.

### Compatibility and support policy

The plugin declares a tested capability baseline, not merely a guessed Codex version.
On startup it performs a non-mutating capability check and selects one of:

- `full`: plugin, skills, engine, and required integrations are available;
- `degraded-guidance`: engine works, but a plugin/client capability is unavailable;
- `verification-only`: repository can be checked but not safely migrated in this
  environment;
- `unsupported`: required guarantees cannot be met, with exact remediation shown.

No degraded state may be summarized as full conformance if it prevented a required
control from being evaluated.

### D3 decision

**Approved (Chris, 2026-07-11):** Adopt the three-layer model: **plugin as preferred
Codex adapter, standalone engine as the deterministic authority, managed repository as
the durable contract**. Ship all three from the first usable release. Do not require MCP
for the initial architecture, and do not make repository upgrades implicit in plugin
updates.

Remaining implementation decisions—engine language/package format, signed distribution,
minimum supported Codex clients, and marketplace/publication channels—belong in the
architecture and delivery design after this distribution model is approved.

<a id="d4"></a>
## D4 decision — authoritative project state and human documents

**Status:** **Agreed (Chris, 2026-07-11)**

Structured, machine-readable state is authoritative for inception answers, detected
project facts, applicability inputs, policy versions, lifecycle state, and generated-file
provenance. Human-friendly Markdown views are rendered from that state for review,
discussion, and communication.

This does not make all Markdown disposable or machine-owned. The artifact model has
three explicit ownership classes:

1. **Machine state:** structured, schema-versioned, validated, and updated transactionally.
   It drives policy compilation and lifecycle operations.
2. **Generated views:** reproducible Markdown summaries carrying a generated marker,
   source-state version/hash, and instructions for proposing changes. They are not a
   competing source of truth and are never silently treated as input after hand edits.
3. **Human-owned records:** approved briefs, decisions, risk acceptances, specifications,
   and stakeholder documents intended to preserve human meaning and history. They use
   stable IDs and structured metadata where needed, but their prose remains authoritative
   for the decision or communication they record.

The guided workflow converts proposed human changes into validated state updates, shows
the resulting policy and document impact, obtains required approval, then regenerates
affected views. For asynchronous/file-based work, the system may emit an editable
proposal document or patch, but it must parse and validate that proposal before updating
authoritative state. It must never rely on bidirectional free-form synchronization.

Every projection records enough provenance to detect staleness. Conflicting edits stop
the operation and produce a reviewable reconciliation plan; neither side silently wins.

<a id="d5"></a>
## D5 decision — GitHub issues as executable memory

**Status:** **Agreed (Chris, 2026-07-11)**

Every managed repository requires GitHub Issues and exactly one linked GitHub Project.
The project is a live operational surface, not optional reporting. Its state must remain
synchronized with issue and delivery activity, and it includes the feature roadmap.

### Issue purpose and audiences

An issue is both:

1. a concise, human-readable description of the work and why it matters; and
2. a self-contained execution brief that any authorized AI agent or developer can pick
   up without the originating conversation, reinvestigation, or new product decisions.

Issues are durable memories of planned and completed work. The issue, linked decisions
and specifications, PR, evidence manifest, and completion record together explain what
was intended, what governed it, what changed, and how completion was verified.

### Required two-layer issue interface

Every executable issue has two layers in one body:

- **Human summary:** approximately 5–8 plain-language lines covering what, why now, user
  or project impact, and an observable **Done when** statement. It excludes internal
  implementation noise and is useful to non-developers.
- **AI execution brief:** a collapsed detail section containing everything needed to
  execute safely: task-specific bootstrap reading, current context, governing decisions
  and policy/control references, precise scope, explicit out-of-scope boundaries, key
  files in dependency order, implementation constraints, acceptance criteria, required
  tests, documentation/communication impact, risk/compliance impact, evidence required,
  and the completion gate.

The AI writes issues from a versioned template and validates them before marking them
ready. Humans may edit them, but edits pass the same validation and preserve history.

### Ready-for-execution contract

An issue may be executed autonomously only when its readiness state is `Ready` and all
of the following hold:

- scope and out-of-scope boundaries are unambiguous;
- no product, architecture, policy, regulatory, or risk decision needed by the work is
  unresolved;
- acceptance criteria are independently verifiable;
- required tests and evidence are specified at the appropriate level;
- applicable controls and obligations have stable IDs and resolved breadcrumb routes;
- governing decisions/specifications exist and are approved;
- relevant paths and repository facts have been checked recently enough;
- dependencies, sequencing, and required external/human actions are explicit;
- the issue records the policy/catalog version against which it was prepared.

Before starting work, the lifecycle engine performs a lightweight readiness refresh.
Changed decisions, policy versions, paths, dependencies, or repository facts move the
issue to `Needs refinement`; they do not authorize the implementing AI to invent a new
decision. Refinement updates the issue and records the reason. Purely mechanical path
changes may be regenerated from stable references.

### Work-item hierarchy and taxonomy

- `type:epic` — a parent outcome grouping related features/tasks through native
  sub-issues.
- `type:feature` — a user- or stakeholder-visible capability, optionally under an epic.
- `type:task` — one independently executable implementation slice.
- `type:bug` — a verified defect or quality finding.

Exactly one `type:*` label applies. Project-specific `area:*` labels identify ownership.
Planned work uses `priority:*`; observed defects use `severity:*`. These remain separate
because scheduling intent and observed impact are different facts. Milestones represent
releases only. The label taxonomy has one versioned manifest and is reconciled
idempotently.

Feature and epic issues may begin as lightweight intake records. They become executable
only when promoted and expanded into the full two-layer brief or decomposed into Ready
sub-issues. A large “executable” issue that still requires several independent decisions
or could be safely split is not Ready.

### Completion memory

Closing an issue requires a linked PR/change (`Closes #N` where applicable), final
verification and conformance evidence, documentation/communication updates, and a brief
completion record noting deviations from the original plan. The original intent is not
erased to make history look cleaner; material scope changes update the issue with an
auditable explanation or create follow-up issues.

### Required GitHub Project

Each repository has exactly one linked GitHub Project with, at minimum:

| Field | Values / purpose |
|---|---|
| Status | `Backlog`, `Next`, `In progress`, `Done` — execution lifecycle |
| Horizon | `Now`, `Next`, `Later`, blank — feature roadmap intent |
| Readiness | `Intake`, `Needs refinement`, `Ready`, `Blocked` — execution readiness |

Required saved views:

- **Current work:** Status `Next` or `In progress`.
- **Backlog:** future work not currently authorized for execution.
- **Roadmap:** open `type:feature` issues grouped by Horizon and manually ranked within
  each group.
- **Needs refinement:** promoted work that is not yet executable.

Horizon and Status are deliberately independent. `Horizon: Now` says a feature is part
of committed product direction; it may remain `Status: Backlog` until scheduled.
Fine-grained ordering is the manual order within a Horizon group; dates or releases are
recorded only when they are real commitments.

The roadmap is always a view over live issues, never a hand-maintained status document.
Any stakeholder-facing roadmap document is generated from GitHub and records its source
time/version.

### Synchronization contract

The lifecycle system and GitHub automation keep project state synchronized:

- new tracked issues are added to the linked project;
- promoted, refined, started, merged, closed, reopened, and blocked work receives the
  corresponding project-field transition;
- starting implementation moves Status to `In progress` only after readiness passes;
- merge/closure moves Status to `Done`; reopening restores an appropriate active state;
- parent completion is reconciled from sub-issue state;
- label, field, view, automation, and project-membership drift are detected;
- reconciliation previews potentially destructive changes and snapshots field values
  before option migrations;
- session orientation reports material drift without derailing unrelated emergency work;
- CI or scheduled reconciliation catches drift even when no interactive Codex session
  runs.

GitHub is an external system and can be unavailable. Local lifecycle operations may
continue only within the applicable policy: mutations are queued with provenance, the
repository reports degraded synchronization, and work cannot claim full conformance
until required GitHub state is reconciled.

<a id="d6"></a>
## D6 decision — unavailable security tooling and evidence

**Status:** **Agreed (Chris, 2026-07-11)**

Missing tooling, unavailable integrations, skipped checks, and insufficient evidence
never produce `pass`. The system uses stage-specific enforcement: it blocks at the
earliest lifecycle stage where continuing would create or externalize the risk that the
control exists to prevent.

| Stage | Default behavior when a required control cannot be evaluated |
|---|---|
| Understand / classify / plan | Continue read-only work; record the gap and present install, upgrade, configuration, and manual-evidence options |
| Local build / remediation | Continue only when the unavailable control is not required to make the attempted mutation safe; constrain authority and mark the work unverified |
| Commit / persist generated artifacts | Block when the missing control protects source/history/artifact integrity, such as secret detection, provenance, or destructive-write safety |
| External mutation | Block when required authorization, data-handling, approval, or external-effect controls are unavailable |
| PR ready / merge | Block when required change verification, review, policy, or evidence is incomplete unless D2 explicitly permits an approved corrective exception |
| Release / deployment / publication | Block on every unmet release-blocking control, prohibited exception, expired risk acceptance, or applicable legal/contractual/platform requirement |
| Conformance claim | Never claim the affected control or overall required scope as conformant; report exact coverage and state |

Each control declares its earliest blocking stage and exception policy. A later gate may
be stricter but cannot weaken an earlier one. For example, known live credentials are
blocked before commit; an unavailable production penetration test may allow local
implementation but block the applicable release.

The system should try to restore capability before stopping: detect stale tools, list
supported upgrades or installations, offer a trusted manual evidence path when the
control permits one, and explain the no-install fallback. It may not install software,
broaden permissions, transmit data, or substitute weaker evidence silently.

Exceptions follow D2. An accepted exception changes whether work may cross a specific
gate; it does not change the underlying control result to `pass` or conceal missing
coverage.

<a id="d7"></a>
## D7 workshop — standards, templates, and predictable repository growth

**Status:** **Agreed (Chris, 2026-07-11)**  
**Goal:** Ensure every developer and AI uses the same versioned standards and templates,
online or offline, while allowing repositories of every project type to grow into an
appropriate but predictable structure.

### Requirements clarified by Chris (2026-07-11)

- Projects begin from a core directory structure and expand predictably as capabilities
  and obligations appear.
- The kit cannot predeclare every directory every future project will need; it must
  provide placement rules and guide the AI when new areas are introduced.
- All team members and their AIs follow the same standards and templates.
- Canonical material may be centralized, but offline operation is required.
- Updates must be safe and understandable.
- The design and lifecycle tooling must work across operating systems.

### Options considered

| Model | Strength | Failure mode |
|---|---|---|
| Copy every standard into every repository | Simple offline access; everything visible | Duplication, noisy upgrades, local drift, difficult fleet-wide fixes |
| Cloud-only standards service | Immediate centralized updates; small repositories | No offline/air-gapped operation; historical builds can become irreproducible; service availability becomes a conformance dependency |
| Plugin-only bundled standards | Convenient Codex access and team distribution | Client/plugin versions drift; CI and non-Codex verification lack an independent authority |
| Versioned policy packs with repository lock + local cache | Reproducible, centrally maintainable, offline-capable, CI-friendly | Requires a resolver, signed distribution, cache management, and explicit upgrade machinery |

**Recommendation:** Adopt the fourth model, with the plugin bundling a verified baseline
pack so a first run can work offline.

### Versioned policy-pack model

Standards, controls, templates, schemas, routing metadata, and verification definitions
are published as immutable, signed/versioned **policy packs**. A pack is content-addressed
and contains a manifest with stable artifact/control IDs, semantic version, compatibility
requirements, dependencies, checksums, and migration notes.

```text
Policy registry / signed release
            │ resolve + verify
            ▼
Immutable local policy cache
            │ exact versions/digests
            ▼
Repository policy.lock + effective-policy index
```

The managed repository commits a small lockfile such as:

```yaml
schema: 1
engine: 1.2.0
packs:
  - id: core-engineering
    version: 1.4.2
    digest: sha256:...
  - id: github-delivery
    version: 1.1.0
    digest: sha256:...
  - id: regulated-health-data
    version: 2.0.1
    digest: sha256:...
```

The exact shape is deferred, but the semantics are not: IDs, versions, and content
digests are pinned; resolution is deterministic; a mutable “latest” reference never
defines conformance.

### Online, offline, and archival behavior

- **Normal online use:** resolve pinned packs from a trusted registry/release source,
  verify signature/digest, and store them in an immutable per-user or shared cache.
- **First-run offline use:** the plugin/engine distribution carries the universal
  baseline pack and any explicitly bundled packs.
- **Previously resolved offline use:** use only cached content matching the lockfile
  digest; never substitute another version silently.
- **Air-gapped or regulated use:** mirror the signed registry or vendor the exact packs
  into a repository/approved artifact store.
- **Historical proof:** every release evidence manifest records pack digests. Regulated
  retention policy may require archiving the exact pack bundle with release evidence,
  not merely retaining a registry URL.
- **Missing pack:** report `not-configured`/unavailable and apply D6. Do not claim
  conformance from a description or remembered policy.

Vendoring is a deployment option, not a fork. Vendored packs remain immutable and
digest-verified. Project-specific policy belongs in the project layer described below,
not as edits inside the vendor directory.

### Policy layering and team consistency

Effective policy is compiled in a deterministic precedence order:

```text
universal baseline
  + project-type packs
  + context/regulatory packs
  + organization policy
  + repository-specific additions/decisions
  + recorded exceptions and residual risks
= effective policy
```

Higher layers may add controls or strengthen enforcement. They cannot silently weaken a
lower-layer requirement. A relaxation must use the D2 exception/risk mechanism, and a
prohibited control cannot be relaxed at all. Applicable law, contract, platform, or
organization policy may also forbid project-level exceptions.

The repository lockfile and effective-policy digest are the team synchronization point.
Codex plugins, developers, CI, and scheduled reconciliation resolve the same graph.
Personal AI preferences may affect presentation or local ergonomics but cannot alter the
effective repository policy or evidence result.

### Context loading

Policy packs participate in the agreed breadcrumb system. The repository carries a
small generated effective-policy index: stable ID, title, applicability, enforcement,
source pack/version, and “load when” description. The AI initially sees only routing
metadata. The resolver supplies the exact focused standard/template from the locked pack
when a workflow requires it.

This avoids copying thousands of lines into every repository while ensuring every AI
loads identical content for the same stable ID. Project-specific decisions, risks, and
human-facing documents remain local and are linked from the same index.

### Predictable directory growth through logical roles

A universal physical directory tree would fit some project types badly. Instead, the
system defines stable **logical directory roles** and maps them to concrete project
paths:

```text
ROLE-SOURCE
ROLE-TEST-UNIT
ROLE-TEST-INTEGRATION
ROLE-DOC-USER
ROLE-DOC-DECISION
ROLE-DOC-SPEC
ROLE-EVIDENCE
ROLE-INFRA
ROLE-MIGRATION
```

The project’s structured layout map records the actual paths for roles that apply. For
example, `ROLE-SOURCE` might map to `src/`, `packages/*/src/`, `app/`, or remain absent
for a documentation-only repository. Standards and issue templates reference the role
and resolve it through the map instead of hard-coding one path.

The core scaffold creates only universal management and human-record locations. A new
project capability triggers a **layout rule** that:

1. inspects existing conventions and relevant tool/framework expectations;
2. proposes the applicable module and logical roles;
3. reuses a valid existing path when possible;
4. otherwise proposes the canonical path for that project type/toolchain;
5. shows moves, conflicts, policy effects, and breadcrumb changes in a plan;
6. creates or remaps paths only after approval;
7. updates the layout map, managed-file provenance, effective-policy index, and issue
   references transactionally;
8. verifies the resulting structure and records the decision when multiple valid homes
   existed.

This makes growth predictable without pretending every stack has the same shape. Nested
`AGENTS.md` files are created only when a subtree genuinely needs different durable
instructions; they are not required merely because a directory exists.

### Cross-platform rule

Policy content and state use platform-neutral UTF-8 formats and repository-relative POSIX
paths as logical identifiers. The lifecycle engine performs native path conversion,
atomic writes, locking, permissions checks, and process execution through a cross-platform
implementation. Policy controls declare capabilities/commands structurally rather than
embedding shell command strings. Shell, PowerShell, or platform-specific adapters may be
selected when required, but no universal workflow depends on Bash, GNU utilities, or
manual OS-specific document variants.

### Upgrade contract

Policy updates are explicit repository changes:

1. discover trusted newer versions and compatibility requirements;
2. verify signatures and fetch migration metadata;
3. compute semantic differences—added/removed/strengthened/weakened controls, changed
   templates, layout migrations, evidence invalidation, and new human decisions;
4. produce a human-readable and machine-readable upgrade plan;
5. obtain the required approval;
6. update lockfile/state and apply migrations transactionally;
7. regenerate affected views/routes/issues where safe;
8. re-evaluate conformance and retain the prior pack/evidence history.

Fleet or organization policy may require a minimum pack version, but repositories still
receive a visible migration plan. Urgent revoked/vulnerable policy or tooling versions
surface as blocking upgrade obligations rather than silently rewriting projects.

### D7 decision

**Approved (Chris, 2026-07-11):** Adopt **signed, immutable policy packs + repository
lockfile + verified local cache**, with optional vendoring/mirroring for regulated or
air-gapped environments. Use logical directory roles and rule-driven layout expansion.
Keep project decisions and proof local; keep reusable standards/templates centrally
versioned; resolve both through the stable breadcrumb registry.

<a id="d8"></a>
## D8 workshop — Git and release workflow

**Status:** **Agreed (Chris, 2026-07-11)**  
**Goal:** Make every change traceable, reviewable, reproducible, and communicable without
forcing one release mechanism onto applications, libraries, infrastructure, data, and
documentation projects.

### Recommendation: universal delivery contract, selected release adapter

The core policy should standardize the guarantees around change. Project facts and
organization policy select the mechanics for versioning, release cadence, artifact
publication, deployment, and merge queues.

### Universal Git contract

1. **Git and GitHub are required.** The repository has a declared default branch and
   linked GitHub Project under D5.
2. **The default branch is protected.** Normal work never commits directly to it.
   GitHub rulesets/branch protection require pull requests and applicable status checks.
3. **One Ready issue produces one primary work branch and PR.** The branch carries the
   issue identity (recommended `<type>/<issue-number>-<slug>`). Closely coupled follow-up
   issues may be linked, but a PR that implements unrelated issues fails scope review.
4. **Every PR closes or explicitly advances tracked work.** It links the issue, governing
   decisions/specifications, conformance/evidence result, documentation impact, and any
   deviations or follow-up work.
5. **No merge while required gates are unresolved.** Readiness, tests, security,
   compliance, documentation, review, synchronization, and D2 risk rules determine the
   required checks.
6. **Review independence is policy-derived.** Solo low-risk projects may permit
   self-review with the lack of separation recorded. Team, protected-area, regulated,
   high-impact, and production changes can require CODEOWNERS or independent approval.
7. **History rewriting is constrained.** Force-push and deletion are prohibited on
   protected branches/tags. Work-branch rebasing is allowed when policy permits and does
   not erase required review/evidence history.
8. **Emergency changes use a governed break-glass path.** They remain issue- and
   evidence-linked, use narrowly scoped authority, and automatically create retrospective
   and remediation work. Emergency does not mean unrecorded.
9. **Repository automation is least-authority and pinned.** Actions, reusable workflows,
   apps, and release credentials follow D2/D7 provenance and permission controls.
10. **Merge changes project state.** Successful merge/closure synchronizes the issue and
    Project; failed automation leaves an explicit reconciliation obligation.

### Merge strategy

Use **squash merge as the default** because the Ready issue and PR are the durable unit of
intent, review, and evidence. The surviving commit title follows Conventional Commits
(`feat`, `fix`, `docs`, `refactor`, `test`, `chore`, etc.), and the body retains issue and
decision references.

Allow a repository or organization to select **merge queue / merge commit** when it needs
batched integration, preserved commit topology, or stricter protected-branch throughput.
The invariant is that the final history retains a stable issue/PR/evidence link. Do not
require every intermediate work-branch commit to be polished when squash merge is used;
that adds ceremony without improving the durable record.

### Change communication

Every user-, operator-, integrator-, or stakeholder-visible change records an unreleased
change entry in the same PR. Internal-only changes declare why no external entry is
needed. The entry is audience-tagged so release tooling can produce different views:

- user/customer release notes;
- operator/deployment notes;
- developer/API/package changelog;
- non-developer stakeholder summary;
- compliance/security disclosure where policy permits and requires it.

The source may be a managed changelog, Changesets-like fragments, or another selected
adapter. Hand-edited duplicate release summaries are avoided; published views are
generated from one change record plus linked issues/decisions.

### Version and release adapters

Not every repository ships the same thing. Bootstrap/project classification selects one
or more adapters:

| Project output | Recommended version/release behavior |
|---|---|
| Versioned application/service | SemVer by default; immutable release identifier tied to source SHA and deployment evidence |
| Library/package | Ecosystem-native SemVer, compatibility analysis, package provenance/signing, consumer changelog |
| Monorepo | Fixed or independent package versions selected explicitly; release tooling manages affected packages |
| Infrastructure | Version modules and immutable plans/artifacts; environment promotion is distinct from source version |
| Data pipeline/model | Version code, schemas/contracts, datasets/models where applicable, and reproducibility metadata |
| Documentation | Version with the product or publish immutable build revisions; link validation and audience review gate publication |
| Non-released/internal repository | No invented SemVer requirement; use immutable commits/evidence until a publication/deployment trigger appears |

Version files, package manifests, tags, changelog headings, and release metadata update
through one transactional release operation. Manual independent version edits are
prohibited. Milestones represent releases only; epics and roadmap intent remain Issues
and Project fields.

### Release contract

A release/publish/deploy operation:

1. identifies included merged issues and change records;
2. resolves the version/release adapter and proposes the next identifier;
3. verifies the exact source and effective policy version;
4. runs all release-blocking controls against the release candidate;
5. produces the human conformance summary and machine evidence manifest from D12;
6. updates versions/change records transactionally;
7. creates an immutable GitHub Release/tag when the adapter uses them;
8. builds/publishes/deploys through least-authority automation;
9. records artifact digests, provenance, environment, approvals, and rollback path;
10. synchronizes issues, milestones, project state, and stakeholder communications.

Signed tags, artifact signing, SBOMs, attestations, deployment approvals, and independent
release approval are triggered by distribution, supply-chain, organization, regulatory,
or risk facts. They may also be adopted universally by an organization policy pack.

### Upgrade and tool selection

The lifecycle system should detect the project ecosystem and present maintained tools
that can implement the selected adapter—for example release-fragment/versioning,
dependency update, merge queue, signing, provenance, or changelog tooling. Per the
standing capability rule, it explains trust, permissions, maintenance, compatibility,
cost, and fallback before installation. The policy describes required outcomes; it does
not hard-code an obsolete tool forever.

### D8 decision

**Approved (Chris, 2026-07-11):** Adopt the universal Git contract above. Default to
**Ready issue → issue-named branch → PR → required gates → squash merge**, Conventional
Commit PR/final-commit titles, one source of change records, and transactional releases.
Select versioning, publication, deployment, signing, and merge-queue adapters from
project context and policy rather than making every repository pretend to be the same
kind of product.

<a id="d9"></a>
## D9 decision — supported operating systems

**Status:** **Agreed (Chris, 2026-07-11)**

The first release supports Linux, macOS, and Windows as native environments. WSL, Git
Bash, containers, and similar compatibility layers may be supported adapters, but Windows
users are not required to install a Unix compatibility environment for the universal
workflow.

Policy/state formats, repository-relative logical paths, schemas, evidence, and generated
documents are platform-neutral. The lifecycle engine owns native path handling, atomic
writes, file locking, process invocation, permissions/capability detection, and line-ending
normalization. Universal controls and workflows do not depend on Bash, PowerShell, GNU
utilities, or OS-specific manual instructions; a platform adapter may use them when the
selected project toolchain requires it.

The project publishes and tests a support matrix covering OS versions, CPU architectures,
Codex surfaces, filesystem constraints, and required external runtimes. Exact minimum
versions and architectures are an implementation/release decision backed by CI evidence,
not assumed in this discovery record. An unsupported environment reports that state and
available upgrade/install/container alternatives under the standing capability rule.

<a id="d10"></a>
## D10 decision — Claude starter kit compatibility

**Status:** **Agreed (Chris, 2026-07-11)**

The Codex starter kit imports the Claude kit’s useful concepts, project knowledge, and
durable records; it does not promise file-for-file or command-for-command compatibility
with `CLAUDE.md`, `.claude/`, Claude permission syntax, or Claude-specific slash commands.

Retrofit inspection recognizes known Claude starter-kit versions and produces a semantic
migration plan that:

- identifies existing interview answers, briefs, decisions, specifications, standards,
  modules, issue references, project configuration, and kit-version state;
- preserves human-owned prose, attribution, stable history, GitHub Issues/Project state,
  and evidence wherever meaning remains valid;
- maps Claude concepts to the agreed Codex lifecycle, policy-pack, breadcrumb, structured
  state, and skill interfaces;
- classifies each source artifact as `adopt`, `transform`, `retain-as-history`,
  `supersede`, `conflict`, or `unsupported`;
- shows lossy or ambiguous mappings for human resolution rather than inventing intent;
- validates the resulting repository against the current universal and triggered policy;
- leaves a durable migration record linking old artifacts to their replacements.

Claude and Codex support may coexist temporarily when a project still uses both tools,
but duplicated instructions must have a declared authority and generated synchronization
strategy. The migration cannot maintain two independently edited policy sources. Removing
obsolete Claude runtime files is a planned, reviewable migration step, never a silent
cleanup.

## Proposed delivery sequence

1. Agree on audience, universal baseline, trigger-derived obligations, tracker stance,
   and source-of-truth format.
2. Validate the day-one plugin/lifecycle distribution shape against Codex capabilities.
3. Specify the artifact model and the create, retrofit, and upgrade state machines.
4. Prototype `inspect → plan → apply → verify` against empty, existing, partially
   conforming, and previously managed repositories.
5. Build one vertical workflow: brief → approved answers → policy compilation → scaffold
   or retrofit → verify → evidence summary.
6. Add Codex skills around the lifecycle module and test their expected file/state
   transitions.
7. Ship upgrade semantics and migration fixtures with the first public version.

## Derived build specification

The approved discovery decisions are synthesized into the following implementation
documents:

- `docs/product/PRD.md` — problem, solution, user stories, requirements, success, and
  testing decisions.
- `docs/product/PERSONAS.md` — shared human audience reference for motivations,
  perspectives, constraints, authority, risk, and communication.
- `docs/architecture/ARCHITECTURE.md` — modules, external seam, adapters, repository and
  policy contracts, transaction/security/version models.
- `docs/architecture/LIFECYCLES.md` — create, retrofit, work, control, release, upgrade,
  and risk state machines.
- `docs/architecture/POLICY_PACKS.md` — initial universal/domain pack boundaries and
  regulatory-pack construction rules.
- `docs/roadmap/IMPLEMENTATION_ROADMAP.md` — tracer-bullet milestones and prerequisite
  investigation issues.

These are derived artifacts while this discovery record remains the discussion and source
history. Approved D1–D12 are promoted into `docs/decisions/INDEX.md` and its linked
records, which are the normal authority surface. If a record and its source D-item
conflict, stop and reconcile them through an explicit superseding decision.

## Decisions to work through

| ID | Question | Starting recommendation | Status |
|---|---|---|---|
| [D1](#d1) | Who and what must the first release serve? | All principal project areas; solo-first but team-scalable; GitHub initially; regulated projects fully supported | **Agreed (2026-07-11)** |
| [D2](#d2) | What is universal versus trigger-derived? | Agreed universal trust baseline, context-triggered policy families, and corrective/residual/prohibited exception model | **Agreed (2026-07-11)** |
| [D3](#d3) | Distribution shape? | Plugin adapter + standalone deterministic lifecycle engine + durable managed-repository contract, all from day one | **Agreed (2026-07-11)** |
| [D4](#d4) | Interview and project-state source of truth? | Structured authoritative state; generated Markdown views; separately governed human-owned records | **Agreed (2026-07-11)** |
| [D5](#d5) | Work tracking and roadmap contract? | GitHub Issues + one synchronized GitHub Project required; two-layer executable-memory issues; Horizon roadmap; readiness and reconciliation contracts | **Agreed (2026-07-11)** |
| [D6](#d6) | Security when tooling or evidence is absent? | No false pass; restore capability where possible; stage-specific blocking at the earliest risk boundary; D2 exceptions only | **Agreed (2026-07-11)** |
| [D7](#d7) | Where do standards live and how does structure grow? | Signed immutable policy packs, repository lockfile, verified offline cache/vendor option, logical directory roles, rule-driven expansion | **Agreed (2026-07-11)** |
| [D8](#d8) | How opinionated is Git/release workflow? | Universal protected, issue-linked PR contract; squash default; context-selected version/release/signing/deployment adapters | **Agreed (2026-07-11)** |
| [D9](#d9) | Supported platforms at v1? | Native Linux, macOS, and Windows; WSL optional; published/tested support matrix; no universal shell dependency | **Agreed (2026-07-11)** |
| [D10](#d10) | Compatibility with Claude kit? | Semantic import with preserved human/GitHub history and explicit mapping; no file-for-file runtime compatibility | **Agreed (2026-07-11)** |
| [D11](#d11) | Breadcrumb identity scope? | Stable IDs for governed policy, controls, decisions, and specifications; normal links elsewhere | **Agreed (2026-07-11)** |
| [D12](#d12) | What does “managed by the kit” guarantee? | Versioned, evidence-backed conformance with explicit coverage, applicability, and risk states; no false-green results | **Agreed (2026-07-11)** |

## Suggested first discussion

Start with D1, D2, and D4, while running the D3 capability spike. They determine the
policy compiler’s inputs, what the lifecycle engine must guarantee, and which documents
are authoritative. D3 is no longer a template-versus-functionality choice: create,
retrofit, and upgrade are all required from day one. Once these are agreed, the accepted
choices can be promoted into decision records before implementation begins.

## Verification notes

The following source checks passed on 2026-07-11:

```text
python3 scripts/validate-manifest.py
python3 scripts/lint-dead-refs.py
bash scripts/bootstrap-smoke.sh
shellcheck -S warning scripts/*.sh template/core/ai/scripts/*.sh template/core/ai/scripts/lib/*.sh
```

The official Codex manual fetch helper was attempted but rejected its response because
the expected integrity header was absent. Codex-specific surface claims above were
therefore checked against the linked official Codex documentation pages instead.
