# Validation assertion and evidence-method research

**Status:** Completed bounded research; conclusions require later promotion before they
govern product or architecture

**Issue:** [#101](https://github.com/dragondad22/codex-starter-kit/issues/101)

**Parent:** [#83](https://github.com/dragondad22/codex-starter-kit/issues/83)

**Freshness:** 2026-07-28

**Intended use:** Input to actor/evaluator contracts, finding disposition,
validation-manifest schema design, and the validation-only decomposition gate in #108

## Objective and boundary

Determine the smallest useful taxonomy for mapping predetermined validation assertions to
evidence methods without allowing a test, tool, evaluator, or generated manifest to
redefine intent. The research covers ten representative assertion classes and method
selection rules. It does not approve a schema, module, evaluator kit, product, provider,
quality score, regulated claim, or implementation.

[DEC-0023](../decisions/DEC-0023-validation-manifest-authority-and-lifecycle.md)
governs the boundary: a validation manifest records exact authoritative sources, normalized
claims, and acceptable evidence constraints. Plans and receipts, not the manifest, record
chosen implementations, actors, attempts, and results.

## Method

The investigation used four steps:

1. Revalidated DEC-0023, the professional baseline, operating-profile and distinct-review
   decisions, current architecture, Work Manager, PRD, glossary, #83 seam inventory, and
   decision map at exact source revision `98f7cfe`.
2. Sampled the assertion classes already required by those records: acceptance,
   structure, lifecycle and authority, negative paths and recovery, security and privacy,
   accessibility, personas and user experience, architecture and maintainability,
   documentation and support, and process/provenance/evidence.
3. Compared that sample with current primary sources that distinguish assessment action,
   automation, human judgment, requirements mapping, provenance, and method limitations.
4. Tested the proposed taxonomy against ambiguous applicability, missing capability,
   incomplete coverage, stale evidence, conflicting results, subjective recommendations,
   and accepted-risk scenarios.

The stopping condition was met when every sampled class had at least one justified route,
known negative paths, and explicit limits. No external product, credential, installation,
paid service, sensitive data, or repository mutation was used.

## Sources

| Source | Version or snapshot | Use | Applicability and limit |
|---|---|---|---|
| Repository authority: DEC-0004, DEC-0017, DEC-0019, DEC-0020, DEC-0022, DEC-0023, PRD, glossary, architecture, Work Manager, #83 inventory and map | `origin/main` at `98f7cfe`, inspected 2026-07-28 | Governs authority, professional quality, explicit states, evaluator separation, evidence reuse, and current seams | Authoritative for this product; open issues and research prose are context, not product authority |
| [NIST SP 800-53A Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/a/r5/final) | January 2022; Release 5.2.0 dated 2025-08-27; retrieved 2026-07-28 | Distinguishes `examine`, `interview`, and `test` assessment actions applied to specifications, mechanisms, activities, and individuals | Security/privacy assessment source; its method distinctions are useful, but its control catalog is not imported as universal Starter Kit policy |
| [W3C ACT Rules Format 1.1](https://www.w3.org/TR/act-rules-format/) and [WAI evaluation guidance](https://www.w3.org/WAI/test-evaluate/) | W3C Recommendation dated 2026-02-05; guidance retrieved 2026-07-28 | Shows explicit applicability, assumptions, manual/semi-automated/automated modes, and cases where a passed rule still requires further testing | Accessibility-specific; demonstrates method limits rather than defining all product evidence |
| [OWASP ASVS](https://owasp.org/www-project-application-security-verification-standard/) | Stable version 5.0.0; retrieved 2026-07-28 | Demonstrates versioned, stable requirement references for security verification | Supplies security requirement candidates, not proof that one scan or test method is sufficient |
| [SLSA 1.2 provenance](https://slsa.dev/spec/v1.2/provenance), [source verification](https://slsa.dev/spec/v1.2/verifying-source), and [declared limits](https://slsa.dev/spec/v1.2/about) | Version 1.2, approved pages retrieved 2026-07-28 | Separates provenance/attestation, trust roots, expectations, and verification from product or code-quality claims | Supply-chain-specific; useful for provenance scope, not a general quality authority |

## Findings

### 1. The useful taxonomy is multi-axis, not a flat method list

`Automated test`, `static control`, `black-box behavior`, `specialist review`, and `human
attestation` are not mutually exclusive peers:

- **Automation** describes who or what executes a method.
- **Black-box** describes what information and surface the evaluator can observe.
- **Specialist** describes required capability or qualification.
- **Static** and **dynamic** distinguish examining state from exercising behavior.
- **Attestation** records an accountable statement about a fact, action, decision, or
  judgment.
- **Non-pass** is a result or capability disposition, not an evidence method.

A black-box evaluation can be automated, human, or hybrid. A static control can be
automated or manually examined. A specialist can conduct black-box behavior evaluation,
implementation-aware review, or validate an attestation. A manifest that stores only one
of these labels would lose material evidence constraints.

The smallest useful method profile therefore has five independent dimensions:

| Dimension | Minimum vocabulary | Question answered |
|---|---|---|
| Evidence action | `examine`, `exercise`, `review`, `attest` | What action can obtain evidence for this claim? |
| Observation boundary | `artifact`, `implementation-aware`, `public-black-box`, `process-record` | What may the evaluator observe? |
| Execution mode | `automated`, `human`, `hybrid` | How is the action performed? |
| Capability and separation | declared domain capability plus `self-check`, `distinct-context`, `independent`, or stronger governed qualification | Who may perform it, and with what separation? |
| Coverage | enumerated scope, exhaustive bounded set, generated cases, scenario/sample, environment matrix, or explicitly unknown | What population can the result support? |

These are research labels, not an approved schema. Existing governed terms such as
`distinct review pass`, `pass`, `fail`, and `accepted-exception` remain canonical.

### 2. Source class does not select the method

Acceptance criteria, policy, architecture, personas, risks, and assurance additions are
sources of authority, not evidence methods. Any one source can yield several kinds of
claim. For example, a security policy can require a syntactic configuration, runtime
isolation, a qualified design review, and an accountable risk decision. Each normalized
claim needs its own evidence route.

The derivation sequence is:

```text
authoritative source
  -> normalized assertion and applicability
  -> assertion object and observable property
  -> acceptable method profile and required coverage
  -> planned evaluator implementation
  -> attributable result and evidence
```

Tests and evaluators consume that sequence. They do not reverse it by turning their own
capabilities or checks into new intended behavior.

### 3. Representative assertion classes

The table classifies ten representative classes, not every future control.

| Assertion class | Representative accountable sources | Permissible evidence route | Capability, separation, and common failure risks |
|---|---|---|---|
| 1. Functional acceptance and deterministic semantics | Ready issue acceptance, specification, PRD, decision | Exercise through the supported public seam; automated examples, property/state-machine cases, and black-box scenarios may compose | A precise oracle is required. Passing examples can miss unenumerated behavior; a test that copies implementation behavior can encode the defect |
| 2. Structural, schema, and static invariants | Architecture contract, policy control, repository layout or schema | Automated or human examination of exact artifacts, parsed state, types, dependency graph, signatures, or digests | Strong for bounded mechanical properties. Pattern approximations create false positives/negatives; file presence does not prove semantic correctness |
| 3. Lifecycle, authority, and state transitions | Decisions, lifecycle contract, Work Manager, effective policy | Automated state-machine exercise plus examination of plans, mandates, receipts, and rejected transitions | Must test allowed and forbidden transitions. A valid plan or approval does not prove effect authority; fake-adapter evidence cannot establish live behavior |
| 4. Negative paths, recovery, and resilience | Acceptance negative paths, risk record, architecture transaction/recovery contract | Fault injection, interruption/replay tests, black-box recovery scenarios, and examination of residual state and receipts | Success-path evidence is insufficient. Deterministic faults may miss production conditions; destructive or live recovery claims need an approved safe target |
| 5. Security, privacy, and data handling | Effective policy, threat/risk record, security specification, handling declaration and authority | Claim-specific composition of static analysis, adversarial dynamic tests, configuration/provenance examination, and qualified specialist review | Scanners and signatures cover only declared rules. Broad absence claims need bounded threat/scope statements. Acknowledgment or provenance does not prove secure behavior or handling assurance |
| 6. Accessibility | Applicable accessibility standard/policy, persona needs, acceptance | Automated/static rules for mechanically decidable properties plus public-surface keyboard, assistive-technology, or knowledgeable human evaluation where required | W3C states that no tool alone determines site accessibility. Passed or inapplicable rule outcomes can still require further testing; environment and assistive-technology coverage must be explicit |
| 7. Persona, audience, UX, and communication fitness | Human-owned persona, journey, product specification, acceptance | Black-box scenario evaluation by representative users or a declared capable evaluator using explicit criteria; telemetry or user research only when separately authorized | Objective requirement violations must be separated from supported findings, subjective recommendations, and unresolved intent. A generic reviewer cannot invent persona preferences |
| 8. Architecture, maintainability, and professional design | Accepted architecture/decision, coding standards, professional baseline | Static structural controls for mechanical boundaries plus distinct implementation-aware review for coupling, coherence, operability, and tradeoffs | Tests can prove behavior but not by themselves prove that a module is maintainable or authority is well placed. Reviewer capability and source context bound the claim |
| 9. Documentation, setup, support, and release communication | Persona, specification, support/release contract, acceptance | Static link/schema/presence checks, black-box execution of documented procedures, and human review for correctness, clarity, and audience fit | Presence and grammar are weak proxies. A procedure that was not executed on the claimed platform cannot support an operational pass |
| 10. Process, review, provenance, and evidence occurrence | Effective policy, DEC-0020 review requirement, mandate, release or evidence contract | Examination of identity-bound records and attestations, corroborated with exact source/artifact digests and required postconditions | An attestation proves only the statement within issuer identity and trust assumptions. Review occurrence does not prove reviewer capability or change quality; provenance does not prove code quality |

No class has a permanently preferred single method. The table states permissible routes;
the exact normalized claim, risk, lifecycle gate, and governed assurance determine which
route is required.

### 4. Method selection rules

1. **Bind the claim before choosing the check.** Record exact source references,
   normalized claim, subject, scope, applicability, lifecycle gate, and required result
   semantics before selecting an evaluator.
2. **Prefer direct observation.** Examine stored state for artifact properties, exercise
   the supported surface for behavior, review semantic design against explicit criteria,
   and use attestation for accountable human/process facts.
3. **Use the highest public seam that matches the claim.** The lifecycle-engine seam is
   appropriate for product semantics. Internal tests are supporting evidence unless the
   assertion is specifically about the internal object.
4. **Keep method dimensions explicit.** Automation, black-box access, capability,
   separation, and coverage must not be inferred from a method name.
5. **Compose without compensation.** When a claim needs multiple methods, each obligation
   retains its own result. A static pass cannot compensate for a missing behavioral
   evaluation, and no blended score converts a required non-pass into success.
6. **Declare oracle and coverage.** Evidence identifies the expected result, environment,
   case-generation or sample rule, exclusions, tool or procedure version, and known
   false-positive/false-negative risks.
7. **Match capability to judgment.** A role label or model identity does not prove
   capability. Qualified human judgment remains required when governing policy, law,
   risk ownership, or a claim's semantics require it.
8. **Separate violation from advice.** A result is `fail` only when evidence shows an
   applicable governed assertion is false. A plausible improvement is a recommendation;
   ambiguous intent or conflicting interpretation is `needs-review`.
9. **Fail honestly when the route is unavailable.** Missing configuration,
   unavailable capability, unresolved applicability, and absent evaluation remain
   explicit. They never default to pass or `not-applicable`.
10. **Reuse only the proved scope.** DEC-0023 evidence reuse requires an unchanged
    assertion fingerprint, a method that permits reuse, current freshness, valid scope,
    and attribution in the new receipt. Human judgment and environment-sensitive
    black-box results normally need shorter freshness than immutable artifact
    examination.

### 5. Result and non-pass semantics

| Condition | Required treatment |
|---|---|
| Capable evaluation directly shows the applicable claim is true for the required scope | `pass` for that assertion and method obligation only |
| Capable evaluation directly shows the applicable claim is false | `fail` |
| Governed applicability facts show the assertion is irrelevant | `not-applicable` with the facts and rule |
| An applicable method, target, route, approval, or fixture is required but unconfigured | `not-configured` |
| Required capability is authoritatively unavailable | `unsupported` at the applicable capability/assessment boundary; the assertion cannot pass |
| Applicability, intent, evidence, source meaning, or evaluator result is ambiguous or conflicting | `needs-review` |
| Evidence was not collected | Unevaluated/missing required evidence; never infer `pass` or `not-applicable` |
| A qualified authority accepts a failed or incomplete gate | Preserve the underlying result and record a separate `accepted-exception`; residual-risk acceptance likewise never changes the result |
| Input schema, digest, or provenance is invalid | Reject the input rather than assigning an ordinary assertion or manifest disposition |

`Unsupported` is currently a manifest/workflow capability disposition, while the
Control Evaluator's documented assertion result list does not include it. This research
does not resolve that schema boundary. Any later design must still make the unavailable
capability and the unevaluated assertion visible without manufacturing a result.

### 6. Minimum traceability

DEC-0023 already fixes what belongs in the immutable manifest. A later plan or validation
receipt needs enough additional information to reconstruct the evaluation without
changing that identity:

- manifest ID, assertion ID and fingerprint, exact source revision, and target identity;
- selected method action, observation boundary, execution mode, coverage, and freshness;
- evaluator identity, declared capability, required separation/qualification, and
  deliberately withheld context;
- tool/procedure and version, configuration, environment, inputs, oracle, sample or case
  selection, and attempt identity;
- result, rationale, limitations, raw evidence routes/digests, and conflicting evidence;
- reuse source and justification when applicable; and
- accepted-exception or residual-risk linkage while retaining the underlying result.

These are traceability needs, not an approved receipt schema. Secrets and sensitive
evidence remain outside ordinary receipts under effective handling policy.

### 7. Representative trace examples

#### Managed-work dependency transition

Claim: completing the final blocker promotes a dependent to Readiness `Ready` without
selecting Status `Next`.

- Source: issue-tracker contract and Work Manager at exact source revision.
- Route: automated state-machine exercise through the lifecycle engine with an in-memory
  adapter; production GitHub support additionally requires exact live contract evidence.
- Negative path: an in-memory pass cannot be relabeled as live GitHub proof; missing
  Project identities are a non-pass.

#### Accessible user workflow

Claim: a supported user flow satisfies its applicable accessibility requirements.

- Source: exact accessibility policy/specification and referenced persona.
- Route: automated/static ACT-like rules for their exact mapped requirements plus
  knowledgeable black-box evaluation for remaining requirements and claimed environment.
- Negative path: a tool pass whose rules cover only part of the requirement cannot
  establish the aggregate claim.

#### Distinct review on exact source

Claim: a pull request received the required distinct capable review on its final head.

- Source: DEC-0020 and effective policy.
- Route: examine identity-bound PR/review records and exact head digest; attest declared
  reviewer context and capability; apply stronger qualified review when policy requires.
- Negative path: passing checks, a review on an old head, or an approval with no evidence
  of required capability cannot satisfy the assertion.

#### Persona-fit quality receipt

Claim: a concise receipt serves PER-OWNER without internal-language leakage while routing
all limitations and non-pass results.

- Source: PRD, PER-OWNER, evidence-presentation contract, and acceptance.
- Route: static completeness/link checks plus a black-box audience review against explicit
  comprehension and routing criteria.
- Negative path: style preference remains a recommendation unless the governed source
  makes it a requirement; unclear audience intent is `needs-review`.

#### Sensitive-data acknowledgment

Claim: the user acknowledged a special-data-handling notice.

- Source: special-data-handling workflow contract.
- Route: identity-bound process record or attestation can prove acknowledgment occurred.
- Negative path: that evidence cannot prove handling authorization, product assurance,
  safe transmission, or regulatory conformance.

## Conflicting evidence

No primary source conflicts with DEC-0023's authority boundary. The material tensions are
about claim scope:

- NIST treats examine, interview, and test as assessment methods. For this product,
  interview or attestation can clarify intent or locate evidence but cannot alone prove a
  technical behavior when a direct record or exercise is required.
- W3C ACT explicitly permits automated, manual, and hybrid checks and also shows that a
  passed or inapplicable rule can require further testing. Therefore tool outcomes cannot
  be promoted beyond their mapped requirement and coverage.
- SLSA provenance supplies attributable process evidence but explicitly does not establish
  code quality. A valid provenance or review attestation cannot stand in for functional,
  security, UX, or maintainability evidence.
- The repository uses `unsupported` for manifest/workflow capability and a smaller
  Control Evaluator result vocabulary. The later executable design must preserve both
  facts without flattening capability absence into `fail`, `not-configured`, or pass.

## Uncertainty

- The exact manifest and receipt schema, controlled vocabulary, and fingerprint inputs
  remain intentionally unresolved for #108 and later implementation work.
- #102 still owns actor capability, least-knowledge inputs, and formal separation levels.
  This record names the information needed for method selection but does not approve that
  actor contract.
- #103 still owns candidate-finding disposition. This research distinguishes violations,
  recommendations, and unresolved intent but does not define their state machine.
- Method freshness cannot be one global duration. Policy or a later schema must define it
  by assertion, environment volatility, evaluator qualification, and evidence type.
- Empirical false-positive/false-negative rates require specific tools, rules, datasets,
  and target populations. No vendor-independent rate is claimed here.
- Qualified legal, regulatory, safety, privacy, or risk decisions remain human-owned and
  may require authorities not present in an ordinary repository.

## Limitations

- The sample is intentionally capped at ten classes and is not a comprehensive policy or
  regulatory control catalog.
- No evaluator, model, product, test suite, live system, accessibility target, or
  regulated route was benchmarked or qualified.
- Repository examples demonstrate traceability semantics, not current validation-manifest
  implementation capability.
- External methods were generalized only where their semantics were useful. NIST, W3C,
  OWASP, and SLSA do not become Starter Kit product authority merely because they were
  researched.
- This record informs later decisions. It does not make its multi-axis labels canonical
  domain vocabulary or authorize implementation.

## Freshness

Repository authority was inspected at `origin/main` commit `98f7cfe` on 2026-07-28 after
#100 closed and DEC-0023 was promoted. The live Project showed #101 as `Ready / Backlog`
before owner selection and `Ready / In progress` after selection.

External sources were retrieved on 2026-07-28. Revalidate this research before promotion
if DEC-0023, the Control Evaluator state model, actor/evaluator contracts, applicable
professional-quality policy, W3C ACT, OWASP ASVS, NIST SP 800-53A, or SLSA changes.
