package sbcli

import "os"

// Env variables that let the caller supply the organization/space GUIDs
// the CLI sends on provision, update, and bind requests, without having
// to persist them via `sb target`.
const (
	OrganizationGUIDEnvVar = "SB_ORGANIZATION_GUID"
	SpaceGUIDEnvVar        = "SB_SPACE_GUID"
)

// TargetedOrganizationGUID resolves the organization GUID to send on
// outbound OSB requests. Priority: SB_ORGANIZATION_GUID env var >
// value persisted by `sb target -o` in the .sb config > caller
// fallback (a random UUID on create-service, or the instance's stored
// GUID on update/bind).
func TargetedOrganizationGUID(fallback string) string {
	if v := os.Getenv(OrganizationGUIDEnvVar); v != "" {
		return v
	}
	if conf := LoadConfig(); conf.OrganizationGUID != "" {
		return conf.OrganizationGUID
	}
	return fallback
}

// TargetedSpaceGUID mirrors TargetedOrganizationGUID for the space GUID.
func TargetedSpaceGUID(fallback string) string {
	if v := os.Getenv(SpaceGUIDEnvVar); v != "" {
		return v
	}
	if conf := LoadConfig(); conf.SpaceGUID != "" {
		return conf.SpaceGUID
	}
	return fallback
}
