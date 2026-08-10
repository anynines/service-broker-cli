package sbcli

import "os"

// Env variables that let the caller supply the organization/space GUIDs
// the CLI sends on provision, update, and bind requests. When either is
// unset, existing behavior is preserved: a fresh UUID on create, or the
// instance's stored GUID on update/bind.
const (
	OrganizationGUIDEnvVar = "SB_ORGANIZATION_GUID"
	SpaceGUIDEnvVar        = "SB_SPACE_GUID"
)

// EnvOrganizationGUID returns the value of SB_ORGANIZATION_GUID if set
// and non-empty, otherwise fallback.
func EnvOrganizationGUID(fallback string) string {
	if v := os.Getenv(OrganizationGUIDEnvVar); v != "" {
		return v
	}
	return fallback
}

// EnvSpaceGUID returns the value of SB_SPACE_GUID if set and non-empty,
// otherwise fallback.
func EnvSpaceGUID(fallback string) string {
	if v := os.Getenv(SpaceGUIDEnvVar); v != "" {
		return v
	}
	return fallback
}
