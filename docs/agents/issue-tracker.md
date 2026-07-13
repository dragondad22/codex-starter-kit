# Issue Tracker — GitHub

Issues, PRDs, findings, epics, features, and executable tasks live in
[`dragondad22/codex-starter-kit` GitHub Issues](https://github.com/dragondad22/codex-starter-kit/issues).
The linked [Codex Starter Kit Project](https://github.com/users/dragondad22/projects/8),
with Status, Horizon, and Readiness fields, is operational authority.

## Field semantics

The three planning fields are independent. In particular, Project `Backlog` is not a
synonym for a Scrum product backlog and does not mean Horizon `Later`.

| Field | Question answered | Transition contract |
|---|---|---|
| Status | Where is this item in execution? | `Backlog` means tracked but not selected for immediate execution; `Next` means explicitly selected as the immediate queue; `In progress` means delivery has started; `Done` means completed. A Ready item may remain Backlog until selected. |
| Readiness | Could authorized work start now? | `Intake` has not been refined; `Needs refinement` has unresolved specification or authority; `Ready` is executable; `Blocked` has an identified unresolved dependency or control. |
| Horizon | Where does this feature sit in rolling product intent? | `Now`, `Next`, and `Later` apply to feature direction. Tasks normally inherit context from their parent and may leave Horizon blank. Horizon does not select work or assign release membership. |

`Next` is intentionally overloaded by GitHub's field option names: Status `Next` is an
execution queue, while Horizon `Next` is product direction. Always name the field when
the distinction is not obvious.

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

- Search for duplicates before creation.
- Preserve native parent/sub-issue hierarchy when decomposing work.
- Use the two-layer executable issue template for planned implementation.
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
