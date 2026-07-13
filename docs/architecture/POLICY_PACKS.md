# Codex Starter Kit — Initial Policy-Pack Map

**Status:** Draft boundary proposal  
**Rule:** Packs add focused policy; the universal baseline cannot be silently weakened.

## Pack Design Rules

- A pack owns a coherent obligation set with stable control/artifact IDs.
- Packs do not own project-specific decisions or evidence.
- A pack may depend on another pack but circular dependencies are invalid.
- Controls declare applicability, evaluation, enforcement, exception policy, evidence,
  invalidation, earliest gate, and routing metadata.
- Packs are immutable after publication; fixes publish a new version.
- Regulatory packs encode verified control mappings and qualified-review requirements,
  not invented legal conclusions.

## Launch Packs

### `core-trust`

Always active. Owns project classification integrity, truthful status, secrets,
least-authority automation, change traceability, verification integrity, secure defaults,
dependency/provenance hygiene, documentation currency, decision provenance, artifact
ownership, coverage disclosure, recovery, licensing/IP hygiene, and breadcrumb integrity.

### `audience-context`

Always active. Owns the persona-registry contract, stable persona IDs, audience evidence,
human confirmation, authority/permission distinction, anti-stereotype rules, impact
analysis, and persona references in briefs, specs, issues, documentation, interfaces,
tests, evidence summaries, and release communication.

### `github-delivery`

Always active because GitHub is required. Owns issue taxonomy/templates/readiness,
Project fields/views/Horizon, synchronization, protected branch/ruleset outcomes, PR
traceability, review rules, merge completion memory, milestones-as-releases, and
break-glass tracking.

### `release-foundation`

Activated by publication, deployment, packaging, or explicit release management. Owns
change records, version transaction invariants, release evidence, audience communication,
rollback, immutable identifiers, and adapter requirements.

### `runtime-operations`

Activated by deployable services/apps/jobs. Owns environment separation, configuration,
observability, health, rollback, incident readiness, vulnerability response, and runtime
hardening outcomes.

### `public-interface`

Activated by APIs, CLIs, libraries, webhooks, schemas, and consumed file formats. Owns
contract/version compatibility, input validation, negative-path behavior, abuse controls,
consumer documentation, and contract testing.

### `identity-access`

Activated by users, roles, sessions, privileged actions, or machine identities. Owns
threat modeling, access-control matrices, authentication/session handling, negative
authorization tests, privileged auditability, and identity lifecycle.

### `data-governance`

Activated by stored/processed data beyond trivial public content. Owns inventory,
classification, minimization, lineage, quality, retention/deletion, encryption, access,
audit, backup/recovery, and data-subject/contractual behavior where triggered.

### `privacy-personal-data`

Activated by personal/sensitive data or relevant jurisdictions/contracts. Owns privacy
applicability, purpose/legal-basis records where required, notices/consent, rights,
transfers, DPIA/qualified review triggers, breach obligations, and evidence retention.

For v1, these pack boundaries provide applicability vocabulary and truthful
`needs-review`/`unsupported` results only. They do not imply that Codex, a connected tool,
or the development environment is an assured sensitive-data route. Detailed enforcement
and route assurance are Later work in
[issue #21](https://github.com/dragondad22/codex-starter-kit/issues/21).

### `user-experience-accessibility`

Activated by human-facing interfaces or documents. Consumes the `audience-context`
persona registry and owns audience-first text, accessibility requirements/evidence,
error-message safety, journey/usability validation, localization triggers, and human beta
guidance.

### `infrastructure`

Activated by IaC/cloud/network/environment changes. Owns plan review, state protection,
policy-as-code, least privilege, drift detection, environment promotion, destructive
change controls, and infrastructure rollback.

### `data-ml`

Activated by data pipelines, analytics, datasets, statistical/ML models. Owns lineage,
reproducibility, schema/data-quality tests, evaluation, fitness/bias triggers, dataset/model
versioning, and monitoring.

### `ai-systems`

Activated by model calls, retrieval, agents, or AI-generated decisions/content. Owns
prompt-injection/trust boundaries, tool authority, model/data provenance, evaluations,
human oversight, output handling, abuse/safety triggers, and model-change evidence.

### `payments-financial`

Activated by monetary or material financial effects. Owns authorization, idempotency,
reconciliation, calculation integrity, audit trail, refunds/failure handling, and
applicable provider/financial control triggers.

### `minors-vulnerable-users`

Activated by intended or foreseeable use. Owns audience/age safeguards, consent,
communication/content review, stricter data controls, escalation, and qualified policy
review.

### `high-impact-safety`

Activated by physical safety or material medical, employment, housing, credit, legal, or
similar decisions. Owns hazard/impact analysis, human review, independent validation,
explainability/appeal triggers, incident response, and strict release approval.

### `third-party-integrations`

Activated by external APIs, SaaS, connectors, plugins, marketplaces, or vendors. Owns
data-flow assessment, scoped credentials, vendor/provenance review, availability/failure,
revocation, and contractual/control inheritance.

### `team-governance`

Activated by multiple contributors/operators or organization policy. Owns CODEOWNERS,
independent review, separation of duties, branch/ruleset governance, onboarding/offboarding,
ownership, escalation, and approval identity.

### `distribution-supply-chain`

Activated by public/open-source/customer/package/app-store distribution. Owns licenses
and notices, SBOM/provenance, signing/attestation, release channels, vulnerability intake,
support/security policy, and publication communication.

## Later Regulatory Pack Strategy

Regulatory pack names must follow verified scope, not a marketing checklist. Build them
through a controlled process before making a support claim:

1. identify jurisdiction/industry/contract and authoritative source versions;
2. obtain qualified interpretation where needed;
3. map obligations to reusable technical/organizational controls;
4. retain citations, applicability questions, evidence, owner, review cadence, and
   limitations;
5. test representative applicable/not-applicable/uncertain projects;
6. sign and publish immutably;
7. monitor source changes and publish semantic updates.

Initial verified regulatory coverage is a Later release target requiring named expert
review; detection, a notice, or an unverified questionnaire is not regulatory support.

## Control ID Convention

Proposed form:

```text
<PACK>-<AREA>-<NNN>
CORE-SECRETS-001
GITHUB-READY-004
DATA-RETENTION-003
```

IDs are never reused. Superseded controls retain history and point to replacements.
