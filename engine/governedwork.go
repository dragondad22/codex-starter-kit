package engine

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const executableIssueSchemaMarker = "<!-- starter-kit-executable-schema:v1 -->"

const (
	readinessDecisionAssertion = "No unresolved product, architecture, policy, regulatory, or risk decision is hidden in this task."
	readinessContextAssertion  = "An authorized implementer can execute this without the originating conversation."
)

// ExecutableIssueContract is the normalized human-owned two-layer Ready issue.
type ExecutableIssueContract struct {
	SchemaVersion       int                  `json:"schema_version"`
	Parent              string               `json:"parent,omitempty"`
	HumanSummary        string               `json:"human_summary"`
	CurrentContext      string               `json:"current_context"`
	GoverningReferences string               `json:"governing_references"`
	Scope               string               `json:"scope"`
	OutOfScope          string               `json:"out_of_scope"`
	Acceptance          string               `json:"acceptance"`
	Verification        string               `json:"verification"`
	Dependencies        string               `json:"dependencies,omitempty"`
	ReadinessAssertions []string             `json:"readiness_assertions"`
	Subtype             *WorkSubtypeContract `json:"subtype,omitempty"`
}

// GovernedSourceBinding binds one stable authoritative repository path to exact content.
type GovernedSourceBinding struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

// GovernedWorkContract enables the #74-qualified Ready path without reinterpreting v1 evidence.
type GovernedWorkContract struct {
	SchemaVersion             int                     `json:"schema_version"`
	Issue                     ExecutableIssueContract `json:"issue"`
	Sources                   []GovernedSourceBinding `json:"sources"`
	RefreshableContextDigests []string                `json:"refreshable_context_digests,omitempty"`
}

// GovernedWorkContractDigest binds the issue, authoritative sources, and refresh authority together.
func GovernedWorkContractDigest(contract GovernedWorkContract) string {
	return digestJSON(contract)
}

// WorkSubtypeContract carries subtype-specific Ready facts without padding every issue type.
type WorkSubtypeContract struct {
	Question *QuestionWorkContract `json:"question,omitempty"`
	Research *ResearchWorkContract `json:"research,omitempty"`
}

// QuestionWorkContract governs one consequential uncertainty and its promotion route.
type QuestionWorkContract struct {
	Question              string `json:"question"`
	Impact                string `json:"impact"`
	Relationship          string `json:"relationship"`
	AnswerAuthority       string `json:"answer_authority"`
	EvidenceNeeds         string `json:"evidence_needs"`
	ResolutionCriteria    string `json:"resolution_criteria"`
	PromotionDestination  string `json:"promotion_destination"`
	NoPromotionResolution string `json:"no_promotion_resolution,omitempty"`
}

// ResearchWorkContract governs bounded evidence-producing work and its durable output.
type ResearchWorkContract struct {
	Objective          string `json:"objective"`
	IntendedUse        string `json:"intended_use"`
	Scope              string `json:"scope"`
	Exclusions         string `json:"exclusions"`
	Provenance         string `json:"provenance"`
	DepthOrEffort      string `json:"depth_or_effort"`
	Authority          string `json:"authority"`
	StoppingConditions string `json:"stopping_conditions"`
	Output             string `json:"output"`
	Freshness          string `json:"freshness"`
	ReviewNeeds        string `json:"review_needs"`
}

// WorkFreshnessDisposition is one deterministic pre-work qualification result.
type WorkFreshnessDisposition string

const (
	WorkFreshnessFresh                     WorkFreshnessDisposition = "fresh"
	WorkFreshnessMechanicalDriftRepaired   WorkFreshnessDisposition = "mechanical-drift-repaired"
	WorkFreshnessContainedContextRefreshed WorkFreshnessDisposition = "contained-context-refreshed"
	WorkFreshnessNeedsRefinement           WorkFreshnessDisposition = "needs-refinement"
	WorkFreshnessAlreadyDelivered          WorkFreshnessDisposition = "already-delivered"
	WorkFreshnessBlocked                   WorkFreshnessDisposition = "blocked"
)

// WorkDeliveryObservation is normalized evidence that the exact outcome already exists.
type WorkDeliveryObservation struct {
	State              string   `json:"state"`
	SourceRevision     string   `json:"source_revision,omitempty"`
	ContractDigest     string   `json:"contract_digest,omitempty"`
	RepositoryRevision string   `json:"repository_revision,omitempty"`
	ResidualScope      string   `json:"residual_scope,omitempty"`
	Evidence           []string `json:"evidence,omitempty"`
}

// WorkDeliveryClaim is the exact governed outcome identity retained in a delivery PR body.
type WorkDeliveryClaim struct {
	SchemaVersion      int                     `json:"schema_version"`
	ManagedID          string                  `json:"managed_id"`
	SourceRevision     string                  `json:"source_revision"`
	ContractDigest     string                  `json:"contract_digest"`
	ImplementedSources []GovernedSourceBinding `json:"implemented_sources"`
}

// WorkPromotionLink is the exact issue-side backlink required by DEC-0013.
type WorkPromotionLink struct {
	SchemaVersion int    `json:"schema_version"`
	ManagedID     string `json:"managed_id"`
	Path          string `json:"path"`
}

// WorkPromotedRecordBacklink is the authoritative record-side half of DEC-0013 reciprocity.
type WorkPromotedRecordBacklink struct {
	SchemaVersion int    `json:"schema_version"`
	ManagedID     string `json:"managed_id"`
	RepositoryID  string `json:"repository_id"`
	IssueURL      string `json:"issue_url"`
}

// RenderWorkPromotedRecordBacklink returns a human link with exact machine identity.
func RenderWorkPromotedRecordBacklink(link WorkPromotedRecordBacklink) (string, error) {
	parsed, err := url.Parse(link.IssueURL)
	issueNumber := strings.TrimPrefix(link.ManagedID, "issue:")
	if err != nil || parsed == nil || link.SchemaVersion != 1 || link.ManagedID == "" || link.RepositoryID == "" || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || issueNumber == link.ManagedID || issueNumber == "" || strings.TrimSuffix(parsed.Path, "/") == "" || !strings.HasSuffix(strings.TrimSuffix(parsed.Path, "/"), "/issues/"+issueNumber) || containsSensitiveText(link.ManagedID+"\n"+link.RepositoryID+"\n"+link.IssueURL) {
		return "", errors.New("promoted record backlink is invalid")
	}
	content, err := json.Marshal(link)
	if err != nil {
		return "", errors.New("encode promoted record backlink")
	}
	return "Source issue: [" + link.ManagedID + "](" + link.IssueURL + ") <!-- starter-kit-source-issue:" + base64.RawURLEncoding.EncodeToString(content) + " -->", nil
}

// ParseWorkPromotedRecordBacklink finds one exact backlink inside an authoritative record.
func ParseWorkPromotedRecordBacklink(body string) (WorkPromotedRecordBacklink, error) {
	const prefix = "<!-- starter-kit-source-issue:"
	if strings.Count(body, prefix) != 1 {
		return WorkPromotedRecordBacklink{}, errors.New("promoted record requires exactly one source-issue backlink")
	}
	start := strings.Index(body, prefix) + len(prefix)
	end := strings.Index(body[start:], " -->")
	if end < 0 {
		return WorkPromotedRecordBacklink{}, errors.New("promoted record backlink is malformed")
	}
	content, err := base64.RawURLEncoding.DecodeString(body[start : start+end])
	if err != nil || len(content) > 16<<10 {
		return WorkPromotedRecordBacklink{}, errors.New("promoted record backlink is malformed")
	}
	var link WorkPromotedRecordBacklink
	if json.Unmarshal(content, &link) != nil {
		return WorkPromotedRecordBacklink{}, errors.New("promoted record backlink is malformed")
	}
	canonical, err := RenderWorkPromotedRecordBacklink(link)
	if err != nil || !strings.Contains(body, canonical) {
		return WorkPromotedRecordBacklink{}, errors.New("promoted record backlink is invalid")
	}
	return link, nil
}

// RenderWorkPromotionComment returns a human link plus a canonical machine-verifiable backlink.
func RenderWorkPromotionComment(link WorkPromotionLink) (string, error) {
	if link.SchemaVersion != 1 || link.ManagedID == "" || validateRelativePath(".", link.Path) != nil || containsSensitiveText(link.ManagedID+"\n"+link.Path) {
		return "", errors.New("work promotion link is invalid")
	}
	content, err := json.Marshal(link)
	if err != nil {
		return "", errors.New("encode work promotion link")
	}
	segments := strings.Split(link.Path, "/")
	for index := range segments {
		segments[index] = url.PathEscape(segments[index])
	}
	return "Promoted authoritative record: [" + link.Path + "](../blob/HEAD/" + strings.Join(segments, "/") + ")\n\n<!-- starter-kit-promotion:" + base64.RawURLEncoding.EncodeToString(content) + " -->", nil
}

// ParseWorkPromotionComment verifies one exact DEC-0013 issue-side backlink.
func ParseWorkPromotionComment(body string) (WorkPromotionLink, error) {
	const prefix = "<!-- starter-kit-promotion:"
	if strings.Count(body, prefix) != 1 {
		return WorkPromotionLink{}, errors.New("promotion comment requires exactly one backlink")
	}
	start := strings.Index(body, prefix) + len(prefix)
	end := strings.Index(body[start:], " -->")
	if end < 0 {
		return WorkPromotionLink{}, errors.New("promotion comment backlink is malformed")
	}
	content, err := base64.RawURLEncoding.DecodeString(body[start : start+end])
	if err != nil || len(content) > 16<<10 {
		return WorkPromotionLink{}, errors.New("promotion comment backlink is malformed")
	}
	var link WorkPromotionLink
	if json.Unmarshal(content, &link) != nil {
		return WorkPromotionLink{}, errors.New("promotion comment backlink is malformed")
	}
	canonical, err := RenderWorkPromotionComment(link)
	if err != nil || body != canonical {
		return WorkPromotionLink{}, errors.New("promotion comment backlink is invalid")
	}
	return link, nil
}

// RenderWorkDeliveryClaim returns the stable machine marker #75 places in a pull request body.
func RenderWorkDeliveryClaim(claim WorkDeliveryClaim) (string, error) {
	if claim.SchemaVersion != 1 || claim.ManagedID == "" || claim.SourceRevision == "" || !validSHA256Digest(claim.ContractDigest) || len(claim.ImplementedSources) == 0 || len(claim.ImplementedSources) > 64 || containsSensitiveText(claim.ManagedID+"\n"+claim.SourceRevision) {
		return "", errors.New("work delivery claim is invalid")
	}
	seenIDs := map[string]bool{}
	seenPaths := map[string]bool{}
	for _, source := range claim.ImplementedSources {
		if source.ID == "" || seenIDs[source.ID] || source.Path == "" || seenPaths[source.Path] || !validSHA256Digest(source.Digest) || validateRelativePath(".", source.Path) != nil {
			return "", errors.New("work delivery claim contains an invalid implemented source")
		}
		seenIDs[source.ID] = true
		seenPaths[source.Path] = true
	}
	content, err := json.Marshal(claim)
	if err != nil {
		return "", errors.New("encode work delivery claim")
	}
	return "<!-- starter-kit-delivery:" + base64.RawURLEncoding.EncodeToString(content) + " -->", nil
}

// ParseWorkDeliveryClaim reads one exact claim without treating ordinary PR prose as evidence.
func ParseWorkDeliveryClaim(body string) (WorkDeliveryClaim, error) {
	const prefix = "<!-- starter-kit-delivery:"
	if strings.Count(body, prefix) != 1 {
		return WorkDeliveryClaim{}, errors.New("pull request body requires exactly one delivery claim")
	}
	start := strings.Index(body, prefix) + len(prefix)
	end := strings.Index(body[start:], " -->")
	if end < 0 {
		return WorkDeliveryClaim{}, errors.New("pull request delivery claim is malformed")
	}
	content, err := base64.RawURLEncoding.DecodeString(body[start : start+end])
	if err != nil || len(content) > 16<<10 {
		return WorkDeliveryClaim{}, errors.New("pull request delivery claim is malformed")
	}
	var claim WorkDeliveryClaim
	if err := json.Unmarshal(content, &claim); err != nil {
		return WorkDeliveryClaim{}, errors.New("pull request delivery claim is malformed")
	}
	canonical, err := RenderWorkDeliveryClaim(claim)
	if err != nil || !strings.Contains(body, canonical) {
		return WorkDeliveryClaim{}, errors.New("pull request delivery claim is invalid")
	}
	return claim, nil
}

// WorkFreshnessAssessment explains one source-bound classification without granting authority.
type WorkFreshnessAssessment struct {
	Disposition    WorkFreshnessDisposition `json:"disposition"`
	ContractDigest string                   `json:"contract_digest"`
	SourceDigests  map[string]string        `json:"source_digests"`
	Reasons        []string                 `json:"reasons"`
	Repairs        []string                 `json:"repairs,omitempty"`
}

// ManagedWorkQualification is desired-work provenance, not an external-effect mandate.
type ManagedWorkQualification struct {
	SchemaVersion            int                     `json:"schema_version"`
	ID                       string                  `json:"qualification_id"`
	IssueManagedID           string                  `json:"issue_managed_id"`
	SourceRevision           string                  `json:"source_revision"`
	OperatingProfileRevision string                  `json:"operating_profile_revision"`
	ObservationRevision      string                  `json:"observation_revision"`
	ConfigurationRevision    string                  `json:"configuration_revision"`
	Target                   WorkTarget              `json:"target"`
	Assessment               WorkFreshnessAssessment `json:"assessment"`
}

// ExecutableIssueContractDigest returns the canonical content identity of a parsed contract.
func ExecutableIssueContractDigest(contract ExecutableIssueContract) string {
	contract.ReadinessAssertions = slices.Clone(contract.ReadinessAssertions)
	return digestJSON(contract)
}

// ExecutableIssueContextDigest identifies one exact replaceable Current context fragment.
func ExecutableIssueContextDigest(context string) string {
	return digestBytes([]byte(strings.TrimSpace(strings.ReplaceAll(context, "\r\n", "\n"))))
}

// RenderExecutableIssueContract emits the canonical visible two-layer Ready issue body.
func RenderExecutableIssueContract(contract ExecutableIssueContract) (string, error) {
	if err := validateExecutableIssueContract(contract); err != nil {
		return "", err
	}
	sections := []struct{ heading, body string }{}
	if contract.Parent != "" {
		sections = append(sections, struct{ heading, body string }{"Parent", contract.Parent})
	}
	sections = append(sections,
		struct{ heading, body string }{"Human summary", contract.HumanSummary},
		struct{ heading, body string }{"Current context", contract.CurrentContext},
		struct{ heading, body string }{"Governing decisions, personas, policy, and specifications", contract.GoverningReferences},
		struct{ heading, body string }{"Scope", contract.Scope},
		struct{ heading, body string }{"Out of scope", contract.OutOfScope},
		struct{ heading, body string }{"Acceptance criteria and negative paths", contract.Acceptance},
		struct{ heading, body string }{"Tests, evidence, documentation, and communication required", contract.Verification},
	)
	if contract.Dependencies != "" {
		sections = append(sections, struct{ heading, body string }{"Dependencies, sequencing, and task-specific authority", contract.Dependencies})
	}
	if contract.Subtype != nil && contract.Subtype.Question != nil {
		question := contract.Subtype.Question
		sections = append(sections,
			struct{ heading, body string }{"Question", question.Question},
			struct{ heading, body string }{"Question impact", question.Impact},
			struct{ heading, body string }{"Question relationship", question.Relationship},
			struct{ heading, body string }{"Answer authority", question.AnswerAuthority},
			struct{ heading, body string }{"Evidence needs", question.EvidenceNeeds},
			struct{ heading, body string }{"Resolution criteria", question.ResolutionCriteria},
			struct{ heading, body string }{"Promotion destination", question.PromotionDestination},
		)
		if question.NoPromotionResolution != "" {
			sections = append(sections, struct{ heading, body string }{"No-promotion resolution", question.NoPromotionResolution})
		}
	}
	if contract.Subtype != nil && contract.Subtype.Research != nil {
		research := contract.Subtype.Research
		sections = append(sections,
			struct{ heading, body string }{"Research objective", research.Objective},
			struct{ heading, body string }{"Intended use", research.IntendedUse},
			struct{ heading, body string }{"Research scope", research.Scope},
			struct{ heading, body string }{"Research exclusions", research.Exclusions},
			struct{ heading, body string }{"Source and provenance expectations", research.Provenance},
			struct{ heading, body string }{"Depth or effort budget", research.DepthOrEffort},
			struct{ heading, body string }{"Research authority", research.Authority},
			struct{ heading, body string }{"Stopping conditions", research.StoppingConditions},
			struct{ heading, body string }{"Durable output", research.Output},
			struct{ heading, body string }{"Freshness requirement", research.Freshness},
			struct{ heading, body string }{"Review needs", research.ReviewNeeds},
		)
	}
	assertions := make([]string, 0, len(contract.ReadinessAssertions))
	for _, assertion := range contract.ReadinessAssertions {
		assertions = append(assertions, "- [x] "+assertion)
	}
	sections = append(sections, struct{ heading, body string }{"Readiness assertion", strings.Join(assertions, "\n")})
	var body strings.Builder
	body.WriteString(executableIssueSchemaMarker)
	for _, section := range sections {
		body.WriteString("\n\n## ")
		body.WriteString(section.heading)
		body.WriteString("\n\n")
		body.WriteString(strings.TrimSpace(section.body))
	}
	return body.String(), nil
}

// RenderManagedIssueBody composes the canonical human contract with the machine-owned
// managed-task identity and relationship metadata used by native adapters.
func RenderManagedIssueBody(desired DesiredManagedTask, contract *ExecutableIssueContract) (string, error) {
	if desired.ManagedID == "" {
		return "", errors.New("managed issue body requires a managed ID")
	}
	metadataDesired := desired
	metadataDesired.Title = ""
	metadataDesired.IssueType = ""
	metadataDesired.Readiness = ""
	metadataDesired.Status = ""
	metadataDesired.Horizon = ""
	metadataDesired.ParentHorizon = ""
	metadataDesired.Closed = false
	metadataDesired.ParentContext = nil
	metadataDesired.Dependents = nil
	encoded, err := json.Marshal(metadataDesired)
	if err != nil {
		return "", fmt.Errorf("encode managed issue metadata: %w", err)
	}
	metadata := base64.RawURLEncoding.EncodeToString(encoded)
	machine := "<!-- starter-kit-managed:" + desired.ManagedID + " -->\n<!-- starter-kit-managed-metadata:" + metadata + " -->"
	if contract == nil {
		return machine, nil
	}
	human, err := RenderExecutableIssueContract(*contract)
	if err != nil {
		return "", err
	}
	return human + "\n\n" + machine, nil
}

// ParseExecutableIssueContract parses the canonical task form without inferring missing fields.
func ParseExecutableIssueContract(body string) (ExecutableIssueContract, error) {
	if strings.Count(body, executableIssueSchemaMarker) != 1 {
		return ExecutableIssueContract{}, errors.New("executable issue requires exactly one supported schema marker")
	}
	sections := map[string]string{}
	current := ""
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == executableIssueSchemaMarker || strings.HasPrefix(trimmed, "<!-- starter-kit-managed:") || strings.HasPrefix(trimmed, "<!-- starter-kit-managed-metadata:") {
			continue
		}
		headingPrefix := ""
		if strings.HasPrefix(line, "### ") {
			headingPrefix = "### "
		} else if strings.HasPrefix(line, "## ") {
			headingPrefix = "## "
		}
		if headingPrefix != "" {
			heading := strings.TrimSpace(strings.TrimPrefix(line, headingPrefix))
			canonical, ok := canonicalIssueHeading(heading)
			if !ok {
				return ExecutableIssueContract{}, errors.New("executable issue contains an unknown section")
			}
			if _, duplicate := sections[canonical]; duplicate {
				return ExecutableIssueContract{}, fmt.Errorf("executable issue contains duplicate section: %s", heading)
			}
			sections[canonical] = ""
			current = canonical
			continue
		}
		if current != "" {
			sections[current] += line + "\n"
		} else if trimmed != "" {
			return ExecutableIssueContract{}, errors.New("executable issue contains unexpected content before its first section")
		}
	}
	for key, value := range sections {
		sections[key] = strings.TrimSpace(value)
	}
	if sections["no-promotion-resolution"] == "_No response_" {
		sections["no-promotion-resolution"] = ""
	}
	contract := ExecutableIssueContract{
		SchemaVersion: 1,
		Parent:        sections["parent"], HumanSummary: sections["summary"], CurrentContext: sections["context"],
		GoverningReferences: sections["references"], Scope: sections["scope"], OutOfScope: sections["out-of-scope"],
		Acceptance: sections["acceptance"], Verification: sections["verification"], Dependencies: sections["dependencies"],
	}
	if sections["question"] != "" {
		contract.Subtype = &WorkSubtypeContract{Question: &QuestionWorkContract{
			Question: sections["question"], Impact: sections["question-impact"], Relationship: sections["question-relationship"],
			AnswerAuthority: sections["answer-authority"], EvidenceNeeds: sections["evidence-needs"],
			ResolutionCriteria: sections["resolution-criteria"], PromotionDestination: sections["promotion-destination"],
			NoPromotionResolution: sections["no-promotion-resolution"],
		}}
	}
	if sections["research-objective"] != "" {
		if contract.Subtype == nil {
			contract.Subtype = &WorkSubtypeContract{}
		}
		contract.Subtype.Research = &ResearchWorkContract{
			Objective: sections["research-objective"], IntendedUse: sections["intended-use"], Scope: sections["research-scope"],
			Exclusions: sections["research-exclusions"], Provenance: sections["provenance"], DepthOrEffort: sections["depth-effort"],
			Authority: sections["research-authority"], StoppingConditions: sections["stopping-conditions"], Output: sections["durable-output"],
			Freshness: sections["freshness"], ReviewNeeds: sections["review-needs"],
		}
	}
	for _, line := range strings.Split(sections["readiness"], "\n") {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "- [x] ") {
			contract.ReadinessAssertions = append(contract.ReadinessAssertions, strings.TrimSpace(trimmed[len("- [x] "):]))
		} else if trimmed != "" {
			return ExecutableIssueContract{}, errors.New("executable issue readiness section contains unexpected content")
		}
	}
	if err := validateExecutableIssueContract(contract); err != nil {
		return ExecutableIssueContract{}, err
	}
	return contract, nil
}

// RefreshExecutableIssueContext replaces only the modeled Current context section.
func RefreshExecutableIssueContext(body string, expected ExecutableIssueContract) (string, error) {
	if err := validateExecutableIssueContract(expected); err != nil {
		return "", err
	}
	observed, err := ParseExecutableIssueContract(body)
	if err != nil {
		return "", err
	}
	if !sameExecutableIssueSemantics(observed, expected) {
		return "", errors.New("executable issue semantic sections changed")
	}
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	heading := -1
	headingLine := ""
	end := len(lines)
	for index, line := range lines {
		if line == "## Current context" || line == "### Current context" {
			heading = index
			headingLine = line
			continue
		}
		if heading >= 0 && index > heading && (strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ")) {
			end = index
			break
		}
	}
	if heading < 0 {
		return "", errors.New("executable issue lacks Current context section")
	}
	replacement := []string{headingLine, "", strings.TrimSpace(expected.CurrentContext), ""}
	updated := append(slices.Clone(lines[:heading]), replacement...)
	updated = append(updated, lines[end:]...)
	result := strings.TrimSpace(strings.Join(updated, "\n"))
	parsed, err := ParseExecutableIssueContract(result)
	if err != nil || ExecutableIssueContractDigest(parsed) != ExecutableIssueContractDigest(expected) {
		return "", errors.New("executable issue context refresh did not preserve the governed contract")
	}
	return result, nil
}

func canonicalIssueHeading(heading string) (string, bool) {
	switch heading {
	case "Parent":
		return "parent", true
	case "Human summary":
		return "summary", true
	case "Current context":
		return "context", true
	case "Governing decisions, personas, policy, and specifications":
		return "references", true
	case "Scope":
		return "scope", true
	case "Out of scope":
		return "out-of-scope", true
	case "Acceptance criteria and negative paths":
		return "acceptance", true
	case "Tests, evidence, documentation, and communication required":
		return "verification", true
	case "Dependencies, sequencing, and task-specific authority", "Dependencies, sequencing, and human actions":
		return "dependencies", true
	case "Readiness assertion":
		return "readiness", true
	case "Question":
		return "question", true
	case "Question impact":
		return "question-impact", true
	case "Question relationship":
		return "question-relationship", true
	case "Answer authority":
		return "answer-authority", true
	case "Evidence needs":
		return "evidence-needs", true
	case "Resolution criteria":
		return "resolution-criteria", true
	case "Promotion destination":
		return "promotion-destination", true
	case "No-promotion resolution":
		return "no-promotion-resolution", true
	case "Research objective":
		return "research-objective", true
	case "Intended use":
		return "intended-use", true
	case "Research scope":
		return "research-scope", true
	case "Research exclusions":
		return "research-exclusions", true
	case "Source and provenance expectations":
		return "provenance", true
	case "Depth or effort budget":
		return "depth-effort", true
	case "Research authority":
		return "research-authority", true
	case "Stopping conditions":
		return "stopping-conditions", true
	case "Durable output":
		return "durable-output", true
	case "Freshness requirement":
		return "freshness", true
	case "Review needs":
		return "review-needs", true
	default:
		return "", false
	}
}

func validateExecutableIssueContract(contract ExecutableIssueContract) error {
	if contract.SchemaVersion != 1 {
		return errors.New("executable issue contract uses an unsupported schema")
	}
	for name, value := range map[string]string{
		"human summary": contract.HumanSummary, "current context": contract.CurrentContext,
		"governing references": contract.GoverningReferences, "scope": contract.Scope,
		"out of scope": contract.OutOfScope, "acceptance": contract.Acceptance,
		"verification": contract.Verification,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("executable issue contract lacks %s", name)
		}
		for _, line := range strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n") {
			if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ") {
				return errors.New("executable issue contract content contains an unmodeled section")
			}
		}
	}
	expected := []string{readinessDecisionAssertion, readinessContextAssertion}
	if !slices.Equal(contract.ReadinessAssertions, expected) {
		return errors.New("executable issue contract lacks the exact Ready assertions")
	}
	content := strings.Join([]string{contract.Parent, contract.HumanSummary, contract.CurrentContext, contract.GoverningReferences, contract.Scope, contract.OutOfScope, contract.Acceptance, contract.Verification, contract.Dependencies}, "\n")
	if strings.Contains(content, executableIssueSchemaMarker) || strings.Contains(content, "<!-- starter-kit-managed:") || strings.Contains(content, "<!-- starter-kit-managed-metadata:") {
		return errors.New("executable issue contract content contains reserved machine metadata")
	}
	if containsSensitiveText(content) {
		return errors.New("executable issue contract contains sensitive-looking material")
	}
	if _, err := governedReferenceIDs(contract.GoverningReferences); err != nil {
		return err
	}
	if err := validateIssueSubtypeShape(contract.Subtype); err != nil {
		return err
	}
	return nil
}

func validateIssueSubtypeShape(subtype *WorkSubtypeContract) error {
	if subtype == nil {
		return nil
	}
	if (subtype.Question == nil) == (subtype.Research == nil) {
		return errors.New("executable issue contract requires exactly one subtype")
	}
	if subtype.Question != nil {
		question := subtype.Question
		if strings.TrimSpace(question.Question) == "" || strings.TrimSpace(question.Impact) == "" || !slices.Contains([]string{"blocking", "related"}, question.Relationship) || strings.TrimSpace(question.AnswerAuthority) == "" || strings.TrimSpace(question.EvidenceNeeds) == "" || strings.TrimSpace(question.ResolutionCriteria) == "" || strings.TrimSpace(question.PromotionDestination) == "" {
			return errors.New("question subtype contract is incomplete or invalid")
		}
	}
	if subtype.Research != nil {
		research := subtype.Research
		values := []string{research.Objective, research.IntendedUse, research.Scope, research.Exclusions, research.Provenance, research.DepthOrEffort, research.Authority, research.StoppingConditions, research.Output, research.Freshness, research.ReviewNeeds}
		for _, value := range values {
			if strings.TrimSpace(value) == "" {
				return errors.New("research subtype contract is incomplete")
			}
		}
	}
	return nil
}

func governedReferenceIDs(value string) ([]string, error) {
	ids := []string{}
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		entry := strings.TrimPrefix(line, "- ")
		id, _, ok := strings.Cut(entry, " — ")
		if !strings.HasPrefix(line, "- ") || !ok || id == "" || strings.ContainsAny(id, " \t") || seen[id] {
			return nil, errors.New("governing references require unique '- STABLE-ID — relevance' entries")
		}
		seen[id] = true
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, errors.New("governing references require at least one stable ID")
	}
	slices.Sort(ids)
	return ids, nil
}

func sameExecutableIssueSemantics(left, right ExecutableIssueContract) bool {
	left.CurrentContext = ""
	right.CurrentContext = ""
	return ExecutableIssueContractDigest(left) == ExecutableIssueContractDigest(right)
}

func validateGovernedWorkContract(contract *GovernedWorkContract) error {
	if contract == nil || contract.SchemaVersion != 1 {
		return errors.New("schema-v2 managed work requires a supported governance contract")
	}
	if err := validateExecutableIssueContract(contract.Issue); err != nil {
		return err
	}
	if len(contract.Sources) == 0 {
		return errors.New("governed work contract requires at least one authoritative source binding")
	}
	seenContextDigests := map[string]bool{}
	for _, digest := range contract.RefreshableContextDigests {
		if !validSHA256Digest(digest) || seenContextDigests[digest] {
			return errors.New("governed work contract contains an invalid refreshable context digest")
		}
		seenContextDigests[digest] = true
	}
	seen := map[string]bool{}
	for _, source := range contract.Sources {
		if source.ID == "" || seen[source.ID] || source.Path == "" || !validSHA256Digest(source.Digest) {
			return errors.New("governed work contract contains an invalid source binding")
		}
		if err := validateRelativePath(".", source.Path); err != nil {
			return errors.New("governed work contract contains an unsafe source path")
		}
		seen[source.ID] = true
	}
	referenceIDs, err := governedReferenceIDs(contract.Issue.GoverningReferences)
	if err != nil {
		return err
	}
	sourceIDs := make([]string, 0, len(contract.Sources))
	for _, source := range contract.Sources {
		sourceIDs = append(sourceIDs, source.ID)
	}
	slices.Sort(sourceIDs)
	if !slices.Equal(referenceIDs, sourceIDs) {
		return errors.New("governing reference IDs do not exactly match authoritative source bindings")
	}
	return nil
}

func validateWorkSubtypeContract(task DesiredManagedTask, governance *GovernedWorkContract) error {
	if task.NoPromotionRequired && (task.IssueType != "question" || !task.Closed || task.PromotionRecord != "") {
		return errors.New("no-promotion resolution is valid only for a closed question without a promotion record")
	}
	var subtype *WorkSubtypeContract
	if governance != nil {
		subtype = governance.Issue.Subtype
	}
	if err := validateIssueSubtypeShape(subtype); err != nil {
		return err
	}
	switch task.IssueType {
	case "question":
		if subtype == nil || subtype.Question == nil {
			return errors.New("question work requires exactly one question subtype contract")
		}
		question := subtype.Question
		if strings.TrimSpace(question.Question) == "" || strings.TrimSpace(question.Impact) == "" || !slices.Contains([]string{"blocking", "related"}, question.Relationship) || strings.TrimSpace(question.AnswerAuthority) == "" || strings.TrimSpace(question.EvidenceNeeds) == "" || strings.TrimSpace(question.ResolutionCriteria) == "" || strings.TrimSpace(question.PromotionDestination) == "" {
			return errors.New("question subtype contract is incomplete or invalid")
		}
		if task.Closed && !task.NoPromotionRequired && task.PromotionRecord != question.PromotionDestination {
			return errors.New("closed question promotion does not match its governed destination")
		}
		if task.Closed && !task.NoPromotionRequired && !governanceBindsPath(governance, task.PromotionRecord) {
			return errors.New("closed question promotion lacks an exact governed output binding")
		}
		if task.NoPromotionRequired && strings.TrimSpace(question.NoPromotionResolution) == "" {
			return errors.New("closed question no-promotion resolution must be visible in its executable contract")
		}
		if !task.NoPromotionRequired && question.NoPromotionResolution != "" {
			return errors.New("question promotion conflicts with a visible no-promotion resolution")
		}
	case "research":
		if subtype == nil || subtype.Research == nil {
			return errors.New("research work requires exactly one research subtype contract")
		}
		research := subtype.Research
		if task.Closed && task.PromotionRecord != research.Output {
			return errors.New("closed research promotion does not match its governed output")
		}
		if task.Closed && !governanceBindsPath(governance, task.PromotionRecord) {
			return errors.New("closed research promotion lacks an exact governed output binding")
		}
	default:
		if subtype != nil {
			return errors.New("implementation work cannot carry a question or research subtype contract")
		}
	}
	if subtype != nil && containsSensitiveText(fmt.Sprintf("%v", *subtype)) {
		return errors.New("work subtype contract contains sensitive-looking material")
	}
	return nil
}

func governanceBindsPath(governance *GovernedWorkContract, path string) bool {
	if governance == nil || path == "" {
		return false
	}
	for _, source := range governance.Sources {
		if source.Path == path {
			return true
		}
	}
	return false
}

func qualifyGovernedWork(root string, intent WorkDesiredIntent, observation WorkObservation) (ManagedWorkQualification, error) {
	contract := intent.Governance
	if err := validateGovernedWorkContract(contract); err != nil {
		return ManagedWorkQualification{}, err
	}
	assessment := WorkFreshnessAssessment{
		Disposition: WorkFreshnessFresh, ContractDigest: ExecutableIssueContractDigest(contract.Issue),
		SourceDigests: map[string]string{}, Reasons: []string{},
	}
	for _, source := range contract.Sources {
		assessment.SourceDigests[source.ID] = source.Digest
		if err := validateRelativePath(root, source.Path); err != nil || ensureNoSymlinkComponents(root, source.Path) != nil {
			assessment.Disposition = WorkFreshnessNeedsRefinement
			assessment.Reasons = append(assessment.Reasons, "governed source is unavailable or unsafe: "+source.ID)
			continue
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(source.Path)))
		if err != nil || digestBytes(content) != source.Digest {
			assessment.Disposition = WorkFreshnessNeedsRefinement
			assessment.Reasons = append(assessment.Reasons, "governed source changed or is unavailable: "+source.ID)
		}
		if intent.Task.Closed && intent.Task.PromotionRecord == source.Path && !intent.Task.NoPromotionRequired {
			backlink, backlinkErr := ParseWorkPromotedRecordBacklink(string(content))
			observedIssueURL := ""
			if observation.Task != nil {
				observedIssueURL = observation.Task.IssueURL
			}
			if backlinkErr != nil || backlink.ManagedID != intent.Task.ManagedID || backlink.RepositoryID != intent.Target.RepositoryID || !promotionIssueURLMatchesTarget(backlink.IssueURL, observedIssueURL, intent.Target.Host, intent.Task.ManagedID) {
				assessment.Disposition = WorkFreshnessNeedsRefinement
				assessment.Reasons = append(assessment.Reasons, "promoted output lacks the exact reciprocal managed issue backlink")
			}
			if intent.Task.IssueType == "research" {
				if err := validateResearchRecord(string(content)); err != nil {
					assessment.Disposition = WorkFreshnessNeedsRefinement
					assessment.Reasons = append(assessment.Reasons, err.Error())
				}
			}
		}
	}
	if observation.Task != nil {
		if len(observation.Task.IssueContractProblems) != 0 || observation.Task.IssueContract == nil {
			assessment.Disposition = WorkFreshnessNeedsRefinement
			assessment.Reasons = append(assessment.Reasons, "observed issue lacks a valid executable contract")
		} else if observation.Task.IssueContractDigest != assessment.ContractDigest {
			observedContextDigest := ExecutableIssueContextDigest(observation.Task.IssueContract.CurrentContext)
			if sameExecutableIssueSemantics(*observation.Task.IssueContract, contract.Issue) && slices.Contains(contract.RefreshableContextDigests, observedContextDigest) {
				assessment.Repairs = append(assessment.Repairs, "context")
				assessment.Reasons = append(assessment.Reasons, "non-semantic current context requires a contained refresh")
			} else {
				assessment.Disposition = WorkFreshnessNeedsRefinement
				assessment.Reasons = append(assessment.Reasons, "observed issue outcome, execution brief, or human-owned context changed")
			}
		}
		if observation.Task.Title != intent.Task.Title || observation.Task.IssueType != intent.Task.IssueType {
			assessment.Disposition = WorkFreshnessNeedsRefinement
			assessment.Reasons = append(assessment.Reasons, "human-visible issue identity changed")
		}
		if intent.Task.IssueType == "question" && intent.Task.Closed && intent.Task.PromotionRecord != "" && !intent.Task.NoPromotionRequired && !observation.Task.PromotionBacklink {
			assessment.Repairs = append(assessment.Repairs, "promotion-link")
			assessment.Reasons = append(assessment.Reasons, "the issue-side promotion backlink requires deterministic repair")
		}
	}
	if observation.Delivery != nil {
		switch observation.Delivery.State {
		case "none", "":
		case "complete":
			if slices.Contains([]string{"task", "bug", "feature"}, intent.Task.IssueType) && observation.Delivery.SourceRevision == intent.SourceRevision && observation.Delivery.ContractDigest == assessment.ContractDigest && observation.Delivery.RepositoryRevision != "" && len(observation.Delivery.Evidence) != 0 {
				assessment.Disposition = WorkFreshnessAlreadyDelivered
				assessment.Reasons = append(assessment.Reasons, "the exact outcome is already delivered with retained evidence")
			} else if assessment.Disposition == WorkFreshnessFresh {
				assessment.Disposition = WorkFreshnessNeedsRefinement
				assessment.Reasons = append(assessment.Reasons, "delivery evidence is incomplete or belongs to another governed outcome")
			}
		case "partial":
			if assessment.Disposition == WorkFreshnessFresh {
				assessment.Disposition = WorkFreshnessNeedsRefinement
				assessment.Reasons = append(assessment.Reasons, "partial delivery requires an explicit residual-scope refinement")
			}
		default:
			if assessment.Disposition == WorkFreshnessFresh {
				assessment.Disposition = WorkFreshnessNeedsRefinement
				assessment.Reasons = append(assessment.Reasons, "delivery observation uses an unsupported state")
			}
		}
	}
	if observation.Task != nil && observation.Task.ReadinessOption == intent.Target.OptionIDs["readiness:blocked"] && intent.Task.Readiness == "ready" && assessment.Disposition != WorkFreshnessAlreadyDelivered {
		assessment.Disposition = WorkFreshnessBlocked
		assessment.Reasons = append(assessment.Reasons, "observed Readiness is Blocked without an explicitly selected governed native-blocker transition")
	}
	for _, blocker := range observation.Relationships.Blockers {
		if !blocker.Closed && assessment.Disposition != WorkFreshnessAlreadyDelivered {
			assessment.Disposition = WorkFreshnessBlocked
			assessment.Reasons = append(assessment.Reasons, "a native blocker remains open: "+blocker.ManagedID)
		}
	}
	if assessment.Disposition == WorkFreshnessFresh {
		assessment.Reasons = append(assessment.Reasons, "governed issue, sources, and current native facts match")
	}
	slices.Sort(assessment.Reasons)
	qualification := ManagedWorkQualification{
		SchemaVersion: 1, IssueManagedID: intent.Task.ManagedID, SourceRevision: intent.SourceRevision,
		OperatingProfileRevision: intent.OperatingProfileRevision, ObservationRevision: observation.Revision,
		ConfigurationRevision: observation.ConfigurationRevision, Target: cloneWorkTarget(intent.Target), Assessment: assessment,
	}
	qualification.ID = digestJSON(struct {
		SchemaVersion                                                                                        int
		IssueManagedID, SourceRevision, OperatingProfileRevision, ObservationRevision, ConfigurationRevision string
		Target                                                                                               WorkTarget
		Assessment                                                                                           WorkFreshnessAssessment
	}{qualification.SchemaVersion, qualification.IssueManagedID, qualification.SourceRevision, qualification.OperatingProfileRevision, qualification.ObservationRevision, qualification.ConfigurationRevision, qualification.Target, qualification.Assessment})
	return qualification, nil
}

func promotionIssueURLMatchesTarget(issueURL, observedIssueURL, host, managedID string) bool {
	parsed, err := url.Parse(issueURL)
	issueNumber := strings.TrimPrefix(managedID, "issue:")
	return err == nil && parsed != nil && issueURL == observedIssueURL && parsed.Scheme == "https" && parsed.Host == host && parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == "" && issueNumber != "" && issueNumber != managedID && strings.HasSuffix(strings.TrimSuffix(parsed.Path, "/"), "/issues/"+issueNumber)
}

func validateResearchRecord(body string) error {
	required := []string{"Method", "Sources", "Findings", "Conflicting evidence", "Uncertainty", "Limitations", "Freshness"}
	sections := map[string]string{}
	current := ""
	for _, line := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n") {
		headingPrefix := ""
		if strings.HasPrefix(line, "### ") {
			headingPrefix = "### "
		} else if strings.HasPrefix(line, "## ") {
			headingPrefix = "## "
		}
		if headingPrefix != "" {
			heading := strings.TrimSpace(strings.TrimPrefix(line, headingPrefix))
			if slices.Contains(required, heading) {
				if _, duplicate := sections[heading]; duplicate {
					return errors.New("promoted research record contains a duplicate required section")
				}
				sections[heading] = ""
				current = heading
			} else {
				current = ""
			}
			continue
		}
		if current != "" {
			sections[current] += line + "\n"
		}
	}
	for _, heading := range required {
		if strings.TrimSpace(sections[heading]) == "" {
			return errors.New("promoted research record lacks required method, sources, findings, conflicting evidence, uncertainty, limitations, or freshness")
		}
	}
	return nil
}
