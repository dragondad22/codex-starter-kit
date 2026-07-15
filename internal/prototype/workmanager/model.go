// Package workmanager is a throwaway logic prototype for issue #64.
//
// It asks which lifecycle-engine-facing desired-state and adapter-result contract
// makes GitHub reconciliation deterministic. The package is deliberately pure: it
// performs no terminal, filesystem, network, clock, or credential operations.
package workmanager

type CredentialExpectation struct {
	Mode  string `json:"mode"`
	Actor string `json:"actor"`
}

type Target struct {
	Host                  string            `json:"host"`
	RepositoryID          string            `json:"repository_id"`
	ProjectID             string            `json:"project_id"`
	ConfigurationRevision string            `json:"configuration_revision"`
	FieldIDs              map[string]string `json:"field_ids"`
	OptionIDs             map[string]string `json:"option_ids"`
}

type ReviewRequirement struct {
	Role                 string `json:"role"`
	DistinctContext      bool   `json:"distinct_context"`
	QualifiedIndependent bool   `json:"qualified_independent"`
}

type DesiredWorkItem struct {
	ManagedID       string              `json:"managed_id"`
	IssueType       string              `json:"issue_type"`
	Title           string              `json:"title"`
	ParentManagedID string              `json:"parent_managed_id,omitempty"`
	BlockedBy       []string            `json:"blocked_by,omitempty"`
	Readiness       string              `json:"readiness"`
	Status          string              `json:"status"`
	Phase           string              `json:"phase,omitempty"`
	PromotionRecord string              `json:"promotion_record,omitempty"`
	Review          []ReviewRequirement `json:"review,omitempty"`
	Closed          bool                `json:"closed"`
}

type DesiredIntent struct {
	SchemaVersion  int                   `json:"schema_version"`
	SourceRevision string                `json:"source_revision"`
	InputDigests   map[string]string     `json:"input_digests"`
	OperationID    string                `json:"operation_id"`
	Credential     CredentialExpectation `json:"credential"`
	Target         Target                `json:"target"`
	WorkItems      []DesiredWorkItem     `json:"work_items"`
}

type Capability struct {
	Online                bool      `json:"online"`
	Fresh                 bool      `json:"fresh"`
	Mode                  string    `json:"mode"`
	Actor                 string    `json:"actor"`
	Permissions           []string  `json:"permissions"`
	RESTRemaining         int       `json:"rest_remaining"`
	GraphQLRemaining      int       `json:"graphql_remaining"`
	ConfigurationRevision string    `json:"configuration_revision"`
	ObservedAt            string    `json:"observed_at"`
	Rate                  RateState `json:"rate"`
}

type RateState struct {
	Resource          string `json:"resource"`
	Limit             int    `json:"limit"`
	Used              int    `json:"used"`
	Remaining         int    `json:"remaining"`
	ResetAt           string `json:"reset_at"`
	RetryAfterSeconds int    `json:"retry_after_seconds"`
	Attempt           int    `json:"attempt"`
	MaxAttempts       int    `json:"max_attempts"`
	Disposition       string `json:"disposition"`
}

type ObservedWorkItem struct {
	ManagedID       string              `json:"managed_id"`
	IssueNodeID     string              `json:"issue_node_id"`
	ProjectItemID   string              `json:"project_item_id"`
	Title           string              `json:"title"`
	ReadinessOption string              `json:"readiness_option_id"`
	StatusOption    string              `json:"status_option_id"`
	ParentManagedID string              `json:"parent_managed_id,omitempty"`
	BlockedBy       []string            `json:"blocked_by,omitempty"`
	Phase           string              `json:"phase,omitempty"`
	PromotionRecord string              `json:"promotion_record,omitempty"`
	Review          []ReviewRequirement `json:"review,omitempty"`
	Closed          bool                `json:"closed"`
}

type Observation struct {
	Revision              string                      `json:"revision"`
	ConfigurationRevision string                      `json:"configuration_revision"`
	Host                  string                      `json:"host"`
	RepositoryID          string                      `json:"repository_id"`
	ProjectID             string                      `json:"project_id"`
	FieldIDs              map[string]string           `json:"field_ids"`
	OptionIDs             map[string]string           `json:"option_ids"`
	WorkItems             map[string]ObservedWorkItem `json:"work_items"`
}

type Effect struct {
	ID              string              `json:"id"`
	Kind            string              `json:"kind"`
	ManagedID       string              `json:"managed_id"`
	Marker          string              `json:"marker,omitempty"`
	Title           string              `json:"title,omitempty"`
	ReadinessOption string              `json:"readiness_option_id,omitempty"`
	StatusOption    string              `json:"status_option_id,omitempty"`
	ParentManagedID string              `json:"parent_managed_id,omitempty"`
	BlockedBy       []string            `json:"blocked_by,omitempty"`
	Phase           string              `json:"phase,omitempty"`
	PromotionRecord string              `json:"promotion_record,omitempty"`
	Review          []ReviewRequirement `json:"review,omitempty"`
	Closed          *bool               `json:"closed,omitempty"`
}

type Plan struct {
	SchemaVersion         int               `json:"schema_version"`
	ID                    string            `json:"id"`
	OperationID           string            `json:"operation_id"`
	SourceRevision        string            `json:"source_revision"`
	InputDigests          map[string]string `json:"input_digests"`
	ObservationRevision   string            `json:"observation_revision"`
	ConfigurationRevision string            `json:"configuration_revision"`
	Host                  string            `json:"host"`
	RepositoryID          string            `json:"repository_id"`
	ProjectID             string            `json:"project_id"`
	FieldIDs              map[string]string `json:"field_ids"`
	OptionIDs             map[string]string `json:"option_ids"`
	Preconditions         []string          `json:"preconditions"`
	RequiredApprovals     []string          `json:"required_approvals"`
	Impact                []string          `json:"impact"`
	Recovery              string            `json:"recovery"`
	ExpiresAt             string            `json:"expires_at"`
	Effects               []Effect          `json:"effects"`
}

type EffectReceipt struct {
	SchemaVersion       int      `json:"schema_version"`
	PlanID              string   `json:"plan_id"`
	OperationID         string   `json:"operation_id"`
	EffectID            string   `json:"effect_id"`
	EffectKind          string   `json:"effect_kind"`
	ManagedID           string   `json:"managed_id"`
	Actor               string   `json:"actor"`
	CredentialMode      string   `json:"credential_mode"`
	Authority           []string `json:"authority"`
	SourceRevision      string   `json:"source_revision"`
	ObservationRevision string   `json:"observation_revision"`
	RepositoryID        string   `json:"repository_id"`
	ProjectID           string   `json:"project_id"`
	Outcome             string   `json:"outcome"`
	Attempt             int      `json:"attempt"`
	Detail              string   `json:"detail"`
}

type State struct {
	SchemaVersion      int             `json:"schema_version"`
	Desired            DesiredIntent   `json:"desired"`
	Capability         Capability      `json:"capability"`
	Observation        Observation     `json:"observation"`
	QueuedIntent       *DesiredIntent  `json:"queued_intent,omitempty"`
	Plan               *Plan           `json:"plan,omitempty"`
	Receipts           []EffectReceipt `json:"receipts"`
	AmbiguousManagedID string          `json:"ambiguous_managed_id,omitempty"`
	Disposition        string          `json:"disposition"`
	Message            string          `json:"message"`
}

type Action string

const (
	PlanReconciliation Action = "plan"
	ApplyNextSuccess   Action = "apply-next-success"
	LoseCreateResponse Action = "lose-create-response"
	ObserveAmbiguous   Action = "observe-ambiguous-create"
	HitRateLimit       Action = "hit-rate-limit"
	GoOffline          Action = "go-offline"
	Reconnect          Action = "reconnect"
	RefreshHandshake   Action = "refresh-handshake"
	MigrateFieldOption Action = "migrate-field-option"
	AcceptMigration    Action = "accept-field-migration"
	CompleteBlocker    Action = "complete-blocker"
	ChangeSource       Action = "change-source"
	ResetRate          Action = "reset-rate"
)
