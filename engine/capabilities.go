package engine

import (
	"runtime/debug"

	starterkit "github.com/dragondad22/codex-starter-kit"
)

// ProvenanceStatus is the engine's bounded self-report. External retained evidence must
// establish verification; an executable cannot make itself trusted by assertion.
type ProvenanceStatus string

const ProvenanceUnverified ProvenanceStatus = "unverified"

// EngineIdentity reports observable build facts without claiming external verification.
type EngineIdentity struct {
	Name       string           `json:"name"`
	Version    string           `json:"version"`
	Revision   string           `json:"revision,omitempty"`
	Modified   bool             `json:"modified"`
	Provenance ProvenanceStatus `json:"provenance"`
}

// ProtocolCapability identifies the versioned lifecycle protocol implemented by the
// engine's language-neutral JSON interface.
type ProtocolCapability struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

// CapabilityReport is the read-only compatibility handshake for experience adapters.
type CapabilityReport struct {
	SchemaVersion        int                `json:"schema_version"`
	Engine               EngineIdentity     `json:"engine"`
	Protocol             ProtocolCapability `json:"protocol"`
	Operations           []string           `json:"operations"`
	StatusSchemaVersions []int              `json:"status_schema_versions"`
}

// Capabilities reports static engine and protocol facts without inspecting or mutating a
// repository. Provenance remains unverified until an external trust decision binds this
// identity to retained qualification evidence.
func (e *Engine) Capabilities() CapabilityReport {
	identity := EngineIdentity{
		Name:       "starter-kit",
		Version:    starterkit.Version(),
		Provenance: ProvenanceUnverified,
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				identity.Revision = setting.Value
			case "vcs.modified":
				identity.Modified = setting.Value == "true"
			}
		}
	}
	return CapabilityReport{
		SchemaVersion: 1,
		Engine:        identity,
		Protocol: ProtocolCapability{
			Name:    "starter-kit.lifecycle",
			Version: 1,
		},
		Operations: []string{
			"apply",
			"bootstrap-sandbox",
			"create",
			"inspect",
			"plan",
			"status",
			"verify",
			"verify-plan",
		},
		StatusSchemaVersions: []int{1},
	}
}
