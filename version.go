// Package starterkit exposes release identity shared by the product's distribution
// surfaces. Protocol, schema, policy, and managed-repository versions remain separate
// compatibility contracts.
package starterkit

import (
	_ "embed"
	"encoding/json"
)

//go:embed product-version.json
var productVersionDocument []byte

var productVersion = loadProductVersion()

// Version returns the canonical Codex Starter Kit product release version.
func Version() string {
	return productVersion
}

func loadProductVersion() string {
	var document struct {
		SchemaVersion int    `json:"schema_version"`
		Product       string `json:"product"`
		Version       string `json:"version"`
	}
	if err := json.Unmarshal(productVersionDocument, &document); err != nil || document.SchemaVersion != 1 || document.Product != "codex-starter-kit" || document.Version == "" {
		panic("invalid embedded product-version.json")
	}
	return document.Version
}
