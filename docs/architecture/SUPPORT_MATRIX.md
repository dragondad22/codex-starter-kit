# Phase 1 Lifecycle-Engine Support Matrix

**Status:** Initial source-runtime support
**Scope:** Empty-repository `create` and seed `verify` through the standalone engine seam
**Evidence:** Native CI reports and aggregate semantic comparison retained by the
completing [Issue #30](../evidence/ISSUE-30.md) pull request

The Phase 1 source runtime is supported only for the behavior and environments below. It
is not a production compliance control, a packaged release, a plugin compatibility claim,
or support for retrofit, upgrade, policy distribution, release, sensitive-data routes, or
regulatory coverage. Explicit non-pass control states remain part of supported behavior.

## Tested runtime envelope

| Native target | Runner selector | Architecture authority | Filesystem boundary | Evidence source |
|---|---|---|---|---|
| Linux | GitHub-hosted `ubuntu-latest` | Exact `GOARCH` and runner architecture in the native report | Native filesystem backing the runner's temporary directory | `phase1-native-ubuntu-latest` artifact |
| macOS | GitHub-hosted `macos-latest` | Exact `GOARCH` and runner architecture in the native report | Native filesystem backing the runner's temporary directory | `phase1-native-macos-latest` artifact |
| Windows | GitHub-hosted `windows-latest` | Exact `GOARCH` and runner architecture in the native report | Native filesystem backing the runner's temporary directory | `phase1-native-windows-latest` artifact |

`latest` is a moving CI selector, not an OS-version claim. Every native report records
`ImageOS`, `ImageVersion`, `RUNNER_OS`, `RUNNER_ARCH`, Go version, Git version, and
`GOOS`/`GOARCH`; those resolved values are the exact tested versions for that run. The
completing pull request records its resolved image values before closure. Support outside
those tested architecture/image families is `needs-review`, not inferred from Go's broader
compilation targets.

The filesystem brand is deliberately not guessed. Support assumes only the behaviors
proven in the same native run: same-directory staged rename, observed case behavior,
portable LF-managed text, native path separators, and the reported symlink/junction and
permission capabilities. Network filesystems, removable media, unusual mount options,
and filesystems that do not provide those behaviors are unsupported until separately
tested.

## Required software and distribution

| Requirement | Supported statement |
|---|---|
| Go | Go 1.26.5 is the pinned contributor/CI build toolchain; `go.mod` requires Go 1.26 |
| Git | A native `git` executable is required at runtime; the exact tested version is recorded per native report |
| Python | Python 3.12 is required only for this repository's documentation/foundation validation, not for the lifecycle engine |
| Shell | No Bash, PowerShell, WSL, Git Bash, container, or universal shell is required by the engine |
| Package | Source build is supported; no signed installer, package-manager formula, or prebuilt binary is published yet |
| Codex | Direct engine use does not require a Codex client; plugin/client compatibility belongs to Phase 2 |
| Network | Create, inspect, plan, apply, status, and seed verify operate locally and do not require network access |

## Native proof contract

Every native matrix job runs the same commands and then retains a self-digested JSON
report:

```text
python -m unittest discover -s tests -p "test_*.py"
python scripts/validate_docs.py
go test ./...
go run ./cmd/phase1-evidence capture --output phase1-native-evidence.json
```

The evidence probe uses the public lifecycle seam to inspect an empty Git repository,
create and explicitly plan the same approved input, apply and replay it, report status,
apply an explicit no-change plan, and prepare/execute seed verification. It excludes
repository paths, timestamps, runner labels, and mechanism-specific capability facts from
the portable semantic snapshot. CI downloads all three reports and requires the same
semantic digest before its aggregate gate can pass:

```text
go run ./cmd/phase1-evidence compare --directory phase1-native-evidence
```

The compared semantics include schema/operation identity, stable planning, artifact paths,
ownership and provenance, applied/replay/no-change states, managed lifecycle status, LF
content semantics, every seed control state, aggregate state, and coverage limitations.
Platform mechanisms remain visible in a separate capability list and cannot silently alter
authority, ownership, evidence meaning, or conformance state.

## Verified native invariants

- The same engine-seam and security/recovery tests run natively on all three targets.
- Portable paths reject traversal, absolute forms, reserved names, unsafe normalization,
  trailing-dot/space aliases, and case-fold collisions on every host.
- Symlink fixtures run where native creation is granted; Windows adds a native directory-
  junction rejection fixture. Missing creation authority is reported, never converted to a
  pass.
- Structured Git execution removes hostile inherited overrides, disables interactive and
  repository-local executable configuration, and never interpolates content into a shell.
- Owned leases, exclusive evidence creation, staged same-directory rename, state-last
  commit, rollback, replay, and incomplete-status behavior are exercised on every runner.
- POSIX owner-only mode evidence applies only where POSIX mode bits are meaningful. It is
  `not-applicable` to Windows ACL assurance; no ACL-hardening claim is made.
- Supported platform differences may change capability details and diagnostics, but the
  aggregate semantic digest must remain identical.

## Known limitations

- `CORE-SECRETS-001` remains `not-configured`; no approved repository secret scanner is
  bundled.
- `CORE-RECOVERY-001` remains `needs-review` for an unversioned source build because the
  executing binary cannot bind itself to retained CI evidence. A future versioned release
  must supply that provenance before the control can pass.
- Multi-file local mutation uses staging, state-last commit, replay, and compensation; it
  is not claimed as one crash-atomic filesystem transaction.
- External effects are absent from Phase 1. Later adapters must define their own
  idempotency, evidence, and compensation.
- Windows ACL enforcement, code signing, installer behavior, package-manager behavior,
  minimum Git versions, non-hosted-runner OS versions, additional CPU architectures, and
  non-default filesystems remain `needs-review`.
- Native runner artifacts are retained for 30 days by CI. The durable issue evidence and
  completing pull request preserve the evidence identity, resolved environment summary,
  limitations, and downstream implications.
