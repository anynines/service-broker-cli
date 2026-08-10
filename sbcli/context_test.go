package sbcli

import "testing"

func TestEnvOrganizationGUIDReturnsFallbackWhenUnset(t *testing.T) {
	t.Setenv(OrganizationGUIDEnvVar, "")
	if got := EnvOrganizationGUID("fallback-org"); got != "fallback-org" {
		t.Fatalf("expected fallback when env is unset, got %q", got)
	}
}

func TestEnvOrganizationGUIDReturnsEnvWhenSet(t *testing.T) {
	t.Setenv(OrganizationGUIDEnvVar, "env-org")
	if got := EnvOrganizationGUID("fallback-org"); got != "env-org" {
		t.Fatalf("expected env value to override fallback, got %q", got)
	}
}

func TestEnvSpaceGUIDReturnsFallbackWhenUnset(t *testing.T) {
	t.Setenv(SpaceGUIDEnvVar, "")
	if got := EnvSpaceGUID("fallback-space"); got != "fallback-space" {
		t.Fatalf("expected fallback when env is unset, got %q", got)
	}
}

func TestEnvSpaceGUIDReturnsEnvWhenSet(t *testing.T) {
	t.Setenv(SpaceGUIDEnvVar, "env-space")
	if got := EnvSpaceGUID("fallback-space"); got != "env-space" {
		t.Fatalf("expected env value to override fallback, got %q", got)
	}
}
