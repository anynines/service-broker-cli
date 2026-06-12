package sbcli

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInstanceFallsBackToOfficialEndpointOn404(t *testing.T) {
	legacyCalls := 0
	officialCalls := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/instances/instance-123":
			legacyCalls++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"description":"missing","error":"not_found"}`))
		case "/v2/service_instances/instance-123":
			officialCalls++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":1,"state":"succeeded","service_guid":"svc-1"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := &SBClient{}
	client.SetCredentials(Credentials{Host: ts.URL})

	instance, err := client.Instance("instance-123")
	if err != nil {
		t.Fatalf("expected fallback call to succeed, got error: %v", err)
	}
	if instance == nil {
		t.Fatal("expected non-nil instance")
	}
	if instance.State != "succeeded" {
		t.Fatalf("expected state 'succeeded', got %q", instance.State)
	}
	if legacyCalls != 1 {
		t.Fatalf("expected 1 legacy call, got %d", legacyCalls)
	}
	if officialCalls != 1 {
		t.Fatalf("expected 1 official endpoint call, got %d", officialCalls)
	}
}
