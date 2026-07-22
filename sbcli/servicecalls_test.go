package sbcli

import "testing"

func TestContextPayloadUnmarshalPreservesCFValues(t *testing.T) {
	var ctx ContextPayload
	err := ctx.UnmarshalJSON([]byte(`{
		"platform":"cloudfoundry",
		"organization_guid":"org-from-context",
		"space_guid":"space-from-context",
		"organization_name":"system",
		"space_name":"test",
		"instance_name":"original-name"
	}`))
	if err != nil {
		t.Fatalf("expected unmarshal to succeed, got %v", err)
	}

	if got := ctx.OrganizationGUID; got != "org-from-context" {
		t.Fatalf("expected organization_guid from context, got %v", got)
	}
	if got := ctx.SpaceGUID; got != "space-from-context" {
		t.Fatalf("expected space_guid from context, got %v", got)
	}
	if got := ctx.InstanceName; got != "original-name" {
		t.Fatalf("expected instance_name from context, got %v", got)
	}
	if got := ctx.OrganizationName; got != "system" {
		t.Fatalf("expected organization_name from context, got %v", got)
	}
	if got := ctx.SpaceName; got != "test" {
		t.Fatalf("expected space_name from context, got %v", got)
	}
	if got := ctx.Platform; got != "cloudfoundry" {
		t.Fatalf("expected platform from context, got %v", got)
	}
}

func TestContextPayloadUnmarshalUsesLegacyIDsAsFallback(t *testing.T) {
	var ctx ContextPayload
	err := ctx.UnmarshalJSON([]byte(`{
		"organization_id":"org-from-legacy-id",
		"space_id":"space-from-legacy-id"
	}`))
	if err != nil {
		t.Fatalf("expected unmarshal to succeed, got %v", err)
	}

	if got := ctx.OrganizationGUID; got != "org-from-legacy-id" {
		t.Fatalf("expected organization_guid from legacy id, got %v", got)
	}
	if got := ctx.SpaceGUID; got != "space-from-legacy-id" {
		t.Fatalf("expected space_guid from legacy id, got %v", got)
	}
}
