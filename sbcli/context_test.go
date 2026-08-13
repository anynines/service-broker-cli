package sbcli

import (
	"os"
	"path/filepath"
	"testing"
)

// withIsolatedConfig points $HOME at a temp directory and clears the
// SB_HOST/USERNAME/PASSWORD envs so `Config.load()` reads from a real
// file, not from env credentials. Returns the directory so tests can
// seed a .sb file in it.
func withIsolatedConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("SB_HOST", "")
	t.Setenv("SB_USERNAME", "")
	t.Setenv("SB_PASSWORD", "")
	return dir
}

func writeConfig(t *testing.T, dir string, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ConfigFile), []byte(body), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func TestTargetedOrganizationGUIDFallbackWhenNothingSet(t *testing.T) {
	withIsolatedConfig(t)
	t.Setenv(OrganizationGUIDEnvVar, "")

	if got := TargetedOrganizationGUID("fallback-org"); got != "fallback-org" {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestTargetedOrganizationGUIDPrefersConfigOverFallback(t *testing.T) {
	dir := withIsolatedConfig(t)
	writeConfig(t, dir, `{"host":"http://x","organization_guid":"cfg-org"}`)
	t.Setenv(OrganizationGUIDEnvVar, "")

	if got := TargetedOrganizationGUID("fallback-org"); got != "cfg-org" {
		t.Fatalf("expected config value, got %q", got)
	}
}

func TestTargetedOrganizationGUIDEnvBeatsConfig(t *testing.T) {
	dir := withIsolatedConfig(t)
	writeConfig(t, dir, `{"host":"http://x","organization_guid":"cfg-org"}`)
	t.Setenv(OrganizationGUIDEnvVar, "env-org")

	if got := TargetedOrganizationGUID("fallback-org"); got != "env-org" {
		t.Fatalf("expected env value to win, got %q", got)
	}
}

func TestTargetedSpaceGUIDFallbackWhenNothingSet(t *testing.T) {
	withIsolatedConfig(t)
	t.Setenv(SpaceGUIDEnvVar, "")

	if got := TargetedSpaceGUID("fallback-space"); got != "fallback-space" {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestTargetedSpaceGUIDPrefersConfigOverFallback(t *testing.T) {
	dir := withIsolatedConfig(t)
	writeConfig(t, dir, `{"host":"http://x","space_guid":"cfg-space"}`)
	t.Setenv(SpaceGUIDEnvVar, "")

	if got := TargetedSpaceGUID("fallback-space"); got != "cfg-space" {
		t.Fatalf("expected config value, got %q", got)
	}
}

func TestTargetedSpaceGUIDEnvBeatsConfig(t *testing.T) {
	dir := withIsolatedConfig(t)
	writeConfig(t, dir, `{"host":"http://x","space_guid":"cfg-space"}`)
	t.Setenv(SpaceGUIDEnvVar, "env-space")

	if got := TargetedSpaceGUID("fallback-space"); got != "env-space" {
		t.Fatalf("expected env value to win, got %q", got)
	}
}

func TestTargetPersistsFlagsToConfig(t *testing.T) {
	dir := withIsolatedConfig(t)
	writeConfig(t, dir, `{"host":"http://x","username":"admin"}`)

	cmd := &Commandline{Organization: "new-org", Space: "new-space"}
	Target(cmd)

	// Re-read the config directly to verify persistence.
	conf := LoadConfig()
	if conf.OrganizationGUID != "new-org" {
		t.Fatalf("expected persisted org, got %q", conf.OrganizationGUID)
	}
	if conf.SpaceGUID != "new-space" {
		t.Fatalf("expected persisted space, got %q", conf.SpaceGUID)
	}
}

func TestTargetLeavesUntouchedFieldsAlone(t *testing.T) {
	dir := withIsolatedConfig(t)
	writeConfig(t, dir, `{"host":"http://x","username":"admin","organization_guid":"old-org","space_guid":"old-space"}`)

	// Only update org; space must survive.
	Target(&Commandline{Organization: "new-org"})

	conf := LoadConfig()
	if conf.OrganizationGUID != "new-org" {
		t.Fatalf("expected org updated, got %q", conf.OrganizationGUID)
	}
	if conf.SpaceGUID != "old-space" {
		t.Fatalf("expected space preserved, got %q", conf.SpaceGUID)
	}
}
