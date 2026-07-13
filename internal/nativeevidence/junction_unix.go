//go:build !windows

package nativeevidence

func probeJunction(string) Capability {
	return Capability{ID: "directory-junction", State: "not-applicable", Details: "Windows directory junctions are not a native Unix capability"}
}
