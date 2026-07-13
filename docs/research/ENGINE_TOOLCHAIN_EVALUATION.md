# Lifecycle-engine toolchain evaluation

**Status:** Approved and promoted to DEC-0015  
**Issue:** [#25](https://github.com/dragondad22/codex-starter-kit/issues/25)  
**Freshness:** 2026-07-12  
**Decision:** [DEC-0015](../decisions/DEC-0015-lifecycle-engine-toolchain.md)

## Objective and stopping conditions

Select the smallest credible implementation toolchain for the Phase 1 lifecycle-engine
walking skeleton. The result must support native Linux, macOS, and Windows behavior,
deterministic local filesystem/Git operations, offline use, provenance, safe packaging,
and a future implementation migration without changing the managed-repository contract.

The evaluation stops after comparing three credible candidates against the approved
contract and identifying one recommendation, one fallback, explicit limitations, and
invalidation triggers. It does not prototype the engine, choose every future library,
guarantee platform code-signing reputation, or establish the 1.0.0 release target.

## Governing constraints

- The lifecycle engine is the deterministic authority and highest test seam.
- Runtime use cannot require Codex, a shell compatibility layer, or an installed language
  toolchain.
- Universal behavior must run natively on Linux, macOS, and Windows.
- Plans are structured data, process execution uses executable/argument/environment
  fields, and repository content is untrusted.
- Machine state is schema-versioned and remains the durable contract if the engine is
  reimplemented.
- CI, developers, and the plugin call the same interface.
- Offline verification must remain possible with pinned inputs and imported trust data.

## Candidates

### Go — recommended

Go produces native binaries for the required operating-system and architecture families,
including Darwin, Linux, and Windows on amd64 and arm64. The official toolchain exposes
these targets through `GOOS`/`GOARCH`, while release evidence can still build and test on
each native runner rather than treating cross-compilation as behavioral proof. The Go 1
compatibility policy provides long-lived source compatibility, with explicit exceptions
for security fixes, unspecified behavior, and operating-system interfaces.

Phase 1 can remain standard-library-only: JSON, hashing, Ed25519 primitives, filesystem,
path, process, templates, and testing are available without a runtime dependency graph.
If dependencies become justified, Go modules record hashes in `go.sum`; `go mod verify`
checks cached module content, and `go mod vendor` supports offline builds with a version
manifest.

Tradeoffs: Go does not statically prevent ownership or aliasing errors, its garbage-
collected runtime increases binary size relative to some Rust builds, and OS-specific
filesystem safety still requires careful design and native tests. The toolchain must be
installed for contributors and CI, though users run only the released binary.

### Rust — supported fallback

Rust's tier-1 targets include native host tools for required macOS, Linux, and Windows
targets, and its ownership model is attractive for memory and concurrency safety. Cargo
locks dependency resolution and can vendor registry and Git dependencies for offline use.

Rust is not the first choice because Phase 1 is dominated by orchestration, schemas,
filesystem transactions, and subprocess boundaries rather than unsafe memory or a large
concurrent core. Its compiler and ownership complexity would increase implementation and
review cost before those benefits are material. Windows builds may also introduce target
toolchain/linker choices that expand the support matrix. Rust becomes the preferred
fallback if measured safety, performance, embedding, or distribution requirements make
Go unsuitable.

### Python — foundation-only, not the product runtime

Python is already used for dependency-free foundation validation and remains appropriate
there. It is not selected for the engine runtime because the standard `zipapp` format
still requires a suitable Python interpreter on the target, Windows behavior depends on
file association, and native extensions require per-platform unpacked binaries. Freezing
tools could hide that runtime at the cost of a new packaging and provenance dependency.
That conflicts with the first usable release's direct native-binary goal.

## Decision matrix

| Criterion | Go | Rust | Python |
|---|---|---|---|
| Native Linux/macOS/Windows runtime without installed language | Strong | Strong | Weak without a freezer |
| Standard-library coverage for the Phase 1 seam | Strong | Moderate; serialization normally adds crates | Strong, but interpreter remains |
| Offline dependency path | `go.sum`, verification, vendor | lockfile, frozen mode, vendor | wheelhouse/vendor plus interpreter packaging |
| Dependency/provenance surface at Phase 1 | Lowest with stdlib-only rule | Higher if serialization/CLI crates are used | Low source graph; higher frozen-runtime graph |
| Filesystem/process safety | Explicit validation and tests required | Explicit OS validation plus stronger memory safety | Explicit validation and tests required |
| Contributor complexity | Lowest of native-binary candidates | Highest | Lowest locally, highest at distribution boundary |
| Long-term source compatibility | Documented Go 1 promise | Stable ecosystem with explicit MSRV/edition choices | Version support and packaging must be managed |
| Reimplementation fallback | Language-neutral CLI/schema seam | Same | Same |

No score is presented as proof. Go wins because it satisfies every mandatory constraint
with the fewest Phase 1 dependencies and the smallest contributor/distribution contract.

## Recommended implementation contract

Approved implementation contract:

1. Implement the standalone engine and CLI in Go, initially pinned to Go 1.26.5 for
   contributor and CI builds. Review each supported Go release and pin patch upgrades;
   released users do not need Go installed.
2. Build and test separate native `starter-kit` binaries on the supported Linux, macOS,
   and Windows runners. Cross-compilation may be an additional check, never the native
   semantic-equivalence evidence.
3. Keep Phase 1 standard-library-only. A third-party module requires an issue-backed
   trust/provenance review, a committed `go.sum`, offline strategy, license review, and
   explicit reason the standard library is insufficient.
4. Use JSON for versioned machine state and operation input/results, Markdown only for
   human views, and structured `os/exec` arguments for Git. Do not expose Go types as the
   durable repository or plugin contract.
5. Publish SHA-256 manifests and GitHub artifact attestations for released binaries.
   Retain attestation bundles and trusted-root material so imported releases can be
   verified offline. Treat platform-native code signing/notarization and reputation as a
   release-adapter requirement, not something an attestation silently satisfies.
6. Preserve a language-neutral engine contract and black-box fixture suite. A future
   implementation may replace Go by reproducing the same schema versions, observable
   operations, evidence semantics, and migration behavior.

## Trust, authority, data, cost, compatibility, and fallback

- **Trust:** Go compiler inputs and release artifacts are pinned; Phase 1 accepts no
  third-party runtime modules by default. GitHub attestations establish build provenance,
  not artifact safety.
- **Authority:** Installing Go changes contributor/CI tooling only. Engine execution does
  not gain network, external mutation, content access, or elevated filesystem authority
  from the language choice.
- **Data:** Builds may contact the Go module proxy/checksum service unless the dependency
  graph is already cached or vendored. Phase 1's standard-library-only build avoids module
  downloads after the pinned toolchain is available.
- **Cost:** Go and Rust are open-source toolchains; operational cost is native CI time,
  release storage, signing/reputation services where selected, and maintenance effort.
  No paid service is authorized by this decision.
- **Compatibility:** Source compatibility follows the Go 1 policy, but binaries are built
  per supported OS/architecture and OS behavior remains governed by the support matrix.
- **Fallback:** Rust is the preferred reimplementation candidate. The CLI/schema seam and
  black-box tests prevent the Go module graph from becoming the product contract.

## Downstream impact

| Consumer | Dependency on this selection |
|---|---|
| #26 create walking skeleton | Establishes module/CLI layout, JSON contracts, native binary, and stdlib-only baseline |
| #27 truthful verification | Uses standard hashing/JSON/time abstractions and engine-seam fixtures |
| #28 hostile path defense | Relies on Go path/filesystem/process APIs plus native adversarial tests; language is not treated as the defense |
| #29 transactions and recovery | Requires platform-specific atomic replacement/locking adapters behind the engine seam |
| #30 native equivalence | Builds and executes on each native runner and publishes exact runtime support |
| Plugin and CI | Invoke the binary/JSON seam; do not import Go implementation packages as product authority |
| Retrofit, policy, release, and upgrade | Reuse schemas and operation semantics; third-party dependencies require separate approval |

## Deferred and excluded choices

- Exact CLI syntax, package layout, schemas, and library interfaces belong to #26.
- Exact minimum OS versions, architectures, filesystems, installers, and code-signing
  reputation claims close in #30 with evidence.
- GitHub API, policy registry/signature verification, plugin, release, and upgrade
  dependencies remain owned by their later vertical slices.
- Platform-native signing/notarization, package managers, auto-update, FIPS claims,
  sandboxing, and dynamic plugins are not authorized here.
- This decision does not assign work to the 1.0.0 Milestone or authorize a release.

## Invalidation triggers

Return this decision to review if any of the following becomes true:

- native Phase 1 semantics cannot be implemented without unsafe or unsupported Go/OS
  behavior;
- a required target cannot run the released binary within the approved support contract;
- binary size, startup, memory, performance, or embedding fails an approved measurable
  requirement;
- a binding cryptographic, FIPS, platform-signing, sandbox, or distribution requirement
  cannot be met credibly;
- the standard-library-only boundary causes more security or maintenance risk than a
  reviewed dependency;
- plugin/CI consumers require an in-process ABI rather than the language-neutral seam; or
- a prototype or native test shows Rust or another option materially reduces total risk
  or cost.

## Primary sources

- [Go build targets and toolchain requirements](https://go.dev/doc/install/source)
- [Go modules, hashes, verification, and vendoring](https://go.dev/ref/mod)
- [Go 1 source-compatibility policy](https://go.dev/doc/go1compat)
- [Go release policy and current supported releases](https://go.dev/doc/devel/release)
- [Rust platform support tiers](https://doc.rust-lang.org/stable/rustc/platform-support.html)
- [Cargo dependency resolution and lock behavior](https://doc.rust-lang.org/cargo/reference/resolver.html)
- [Cargo vendoring](https://doc.rust-lang.org/cargo/commands/cargo-vendor.html)
- [Python `zipapp` runtime and native-extension caveats](https://docs.python.org/3/library/zipapp.html)
- [GitHub artifact-attestation scope](https://docs.github.com/en/actions/concepts/security/artifact-attestations)
- [GitHub offline attestation verification](https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/verify-attestations-offline)
