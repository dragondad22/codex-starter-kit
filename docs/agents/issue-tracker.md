# Issue Tracker — GitHub

Issues, PRDs, findings, epics, features, and executable tasks live in
[`dragondad22/codex-starter-kit` GitHub Issues](https://github.com/dragondad22/codex-starter-kit/issues).
The linked [Codex Starter Kit Project](https://github.com/users/dragondad22/projects/8),
with Status, Horizon, Readiness, and Phase fields, is operational authority.

## Field semantics

The four planning fields are independent. In particular, Project `Backlog` is not a
synonym for a Scrum product backlog and does not mean Horizon `Later`.

| Field | Question answered | Transition contract |
|---|---|---|
| Status | Where is this item in execution? | `Backlog` means tracked but not selected for immediate execution; `Next` means explicitly selected as the immediate queue; `In progress` means delivery has started; `Done` means completed. A Ready item may remain Backlog until selected. |
| Readiness | Could authorized work start now? | `Intake` has not been refined; `Needs refinement` has unresolved specification or authority; `Ready` is executable; `Blocked` has an identified unresolved dependency or control. |
| Horizon | Where does this feature sit in rolling product intent? | `Now`, `Next`, and `Later` apply to feature direction. Tasks normally inherit context from their parent and may leave Horizon blank. Horizon does not select work or assign release membership. |
| Phase | Which ordered roadmap outcome contains this work? | `Phase 0` through `Phase 8` are assigned directly to roadmap features. Ordinary children derive Phase through their native parent and leave their own field blank. A directly assigned cross-cutting child records why. Phase is not execution state, Horizon, sprint, Milestone, or release completion. |

`Next` is intentionally overloaded by GitHub's field option names: Status `Next` is an
execution queue, while Horizon `Next` is product direction. Always name the field when
the distinction is not obvious.

Saved Project views are optional, human-owned navigation surfaces over these governed
facts. Individuals and teams may arrange them to fit their work. A `Phases` view, when
used, does not become another roadmap authority and its progress context does not prove a
phase or release complete; its absence or a different layout does not invalidate Phase.

Triage labels route a complete item to the intended executor; they do not supersede the
Project gate. For example, `ready-for-agent` plus Readiness `Blocked` means the brief is
complete and intended for an agent after its blockers resolve, not that work may start.

## Reconciliation checkpoints

Project cleanliness is part of completing work, not optional board administration. After
an issue starts, completes, reopens, gains or loses a dependency, or has a child change
execution state, reconcile this set before handing off:

1. the changed item;
2. its native parent, if any; and
3. every directly dependent issue named by its blocker relationship or executable brief.

Apply these standing rules:

- Starting execution requires Readiness `Ready` and moves Status to `In progress`.
- An incomplete parent moves to `In progress` once any child delivery starts or completes;
  it must not return to `Backlog` merely because no child is currently running.
- When every native child is complete, close the parent and move it to `Done`. If the
  parent's acceptance contract is not actually satisfied, create or attach the concrete
  outstanding child task before leaving the parent open; do not use an unexplained open
  epic as a placeholder for unknown work.
- Completing a blocker immediately re-evaluates each dependent. When no unresolved
  blockers remain, change Readiness from `Blocked` to `Ready`.
- Becoming Ready does not automatically move Status to `Next`. Use `Next` only for work
  deliberately selected as the immediate queue; otherwise a Ready item may stay Backlog.
- A dependent remains `Blocked` while any declared blocker is unresolved.
- Closing completed work moves Status to `Done`; reopening restores the state justified
  by its current readiness, selection, and execution facts.

At session orientation and before final handoff, audit the touched slice for these
invariants. If GitHub automation is absent or fails, make the corrections directly and
record material drift in the governing issue rather than leaving the Project stale.

Use `gh` or the connected GitHub adapter. Pass multiline bodies through files or safe
structured calls; do not interpolate issue content into shell commands.

## Pull requests as a request surface

**Yes.** External PRs from `CONTRIBUTOR`, `FIRST_TIME_CONTRIBUTOR`, or `NONE` author
associations enter the same triage queue as issues. PRs from owners, members, or
collaborators are treated as in-progress team work.

GitHub shares numbering across issues and PRs. Resolve an ambiguous `#N` before acting.

## Decomposition hierarchy

When an epic, feature, or other parent work item is decomposed, add each resulting issue
through GitHub's native parent/sub-issue relationship. A `#N` reference in the child body
is useful narrative context, but it is not a substitute for the native relationship: it
does not populate GitHub's hierarchy or sub-issue progress.

After publishing a decomposition:

1. verify through GitHub that the parent reports the intended children;
2. add every child to the operational Project and set its own Status and Readiness;
3. reconcile the parent's triage label and Project Readiness; and
4. preserve each child's dependency and execution state independently.

An approved, complete decomposition removes `needs-triage` from the parent. If no
parent-level refinement remains, the parent may be `Ready` as a delivery container even
when some children are `Blocked`; that does not make those children executable. Use each
child's triage label, Readiness, and dependencies to determine whether it can start.

## Required behavior

### Conversational capture

Agents share responsibility for noticing when a conversation produces information that
must survive it. At a natural checkpoint—a topic change, conclusion, newly discovered
scope, or handoff—search open and closed GitHub issues before suggesting capture when:

- new work falls outside the active issue;
- an answer must survive the conversation, materially changes planned work, requires
  named authority or evidence, or is likely to be referenced again;
- a product, architecture, policy, regulatory, or risk decision needs durable promotion;
  or
- a new risk, dependency, limitation, or verified defect requires ownership beyond the
  current mandate.

Route a duplicate or contained correction to the active issue instead of creating a new
item. Use the lifecycle-specific issue type for genuinely new work, and promote an
approved material decision into its authoritative record because an issue or comment is
not authority for that domain. Ordinary conversational clarification and exploratory
fragments stay in the active workflow rather than flooding the Project.

The prompt offers the owner a capture action; it does not silently create an issue,
change Project fields, or approve a decision. Discussion may continue while an idea is
being shaped, but material implementation stops until an applicable issue has Readiness
`Ready`. Once execution begins, reference the active `#N`; the completing PR uses
`Closes #N`.

### Lifecycle-specific issue templates

An issue is complete relative to its current lifecycle state; visibility does not imply
executability. Use the smallest template that preserves the work honestly:

- `Feature idea` records an Intake opportunity, audience, reason, and provenance.
- `Design-first initiative` records task-specific breadcrumbs, demand and design context,
  governing decisions, confirmed facts versus hypotheses, open questions, indicative
  surfaces, exclusions, and the promotion gate for later executable sub-issues.
- `Executable work item` is reserved for decision-complete implementation that can pass
  Readiness `Ready`.
- Question, research, and bug records use their subtype contracts and are promoted or
  decomposed rather than padded into task-shaped prose.

Ready task, question, and research forms retain
`<!-- starter-kit-executable-schema:v1 -->` and the exact canonical headings. The marker
is version identity, not authority. Every visible governing-reference line uses
`- STABLE-ID — relevance`; schema-v2 execution binds those IDs one-for-one to safe
repository paths and exact digests. Question/research forms add their sparse subtype
sections to the same parseable contract.

At selection/start, and again only after material change, refresh the current issue,
governing sources, native relationships, Project configuration, and related delivery
claims. The result is exactly one of `fresh`, `mechanical-drift-repaired`,
`contained-context-refreshed`, `needs-refinement`, `already-delivered`, or `blocked`.
Elapsed age never invalidates Ready work. Ordinary fresh work needs no pass comment.
Automation may replace Current context only when the task explicitly delegates that
narrow refresh and no other semantic section changed; otherwise return it to refinement.

Templates prescribe information shape, not conclusions. Do not invent architecture,
scope, tests, or decisions to fill a form. A design-first parent may carry feature-level
acceptance while deferring implementation tests to its eventual Ready sub-issues.

Issue bodies state desired outcomes and task-specific memory. They reference bootstrap,
policy, decisions, and normal Git flow instead of repeating them. Record an issue-level
workflow override only when it is exceptional, explicit, justified, and authoritative.

- Search for duplicates before creation.
- Preserve native parent/sub-issue hierarchy when decomposing work.
- Use the two-layer executable issue template only for planned implementation promoted to
  Ready; preserve earlier work through the Intake or design-first template.
- Reserve question issues for consequential uncertainties that need durable resolution;
  keep ordinary conversational clarification off the Project.
- Require bounded objectives, authority, depth or effort, stopping conditions, provenance,
  and durable outputs for research issues.
- Promote material question and research results into the correct authoritative record.
  Link that record back to the issue and identify it in the issue's closing comment.
- Keep Horizon as rolling feature intent. Use one native GitHub Milestone as the finite
  manifest for each named release and one aggregate release issue for readiness and
  publication; milestone percentage alone is not evidence that a release is ready.
- Do not implement until readiness passes.
- Keep issue and Project fields synchronized through the lifecycle.
- Treat dependency completion as a required reconciliation trigger: re-evaluate and
  promote fully unblocked dependents to Readiness `Ready`.
- Use `Closes #N` from the completing PR.
- Preserve completion evidence and material deviations as work memory.
