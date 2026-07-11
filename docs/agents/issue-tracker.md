# Issue Tracker — GitHub

Issues, PRDs, findings, epics, features, and executable tasks live in
[`dragondad22/codex-starter-kit` GitHub Issues](https://github.com/dragondad22/codex-starter-kit/issues).
The linked [Codex Starter Kit Project](https://github.com/users/dragondad22/projects/8),
with Status, Horizon, and Readiness fields, is operational authority.

Use `gh` or the connected GitHub adapter. Pass multiline bodies through files or safe
structured calls; do not interpolate issue content into shell commands.

## Pull requests as a request surface

**Yes.** External PRs from `CONTRIBUTOR`, `FIRST_TIME_CONTRIBUTOR`, or `NONE` author
associations enter the same triage queue as issues. PRs from owners, members, or
collaborators are treated as in-progress team work.

GitHub shares numbering across issues and PRs. Resolve an ambiguous `#N` before acting.

## Required behavior

- Search for duplicates before creation.
- Use the two-layer executable issue template for planned implementation.
- Do not implement until readiness passes.
- Keep issue and Project fields synchronized through the lifecycle.
- Use `Closes #N` from the completing PR.
- Preserve completion evidence and material deviations as work memory.
