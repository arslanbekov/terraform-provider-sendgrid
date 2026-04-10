package sendgrid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupSSOIntegrationMockServer(handler http.Handler) (*httptest.Server, *Config) {
	server := httptest.NewServer(handler)
	config := &Config{
		APIKey: "test-api-key",
		Host:   server.URL,
	}
	return server, config
}

func TestSSOIntegrationRead_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/sso/integrations/integ-123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id":"integ-123",
			"name":"Test IdP",
			"enabled":true,
			"signin_url":"https://idp.example.com/signin",
			"signout_url":"https://idp.example.com/signout",
			"entity_id":"https://idp.example.com/12345",
			"completed_integration":true,
			"single_signon_url":"https://sso.sendgrid.com/saml/integ-123",
			"audience_url":"https://sso.sendgrid.com/saml/integ-123"
		}`))
	})

	server, config := setupSSOIntegrationMockServer(mux)
	defer server.Close()

	r := resourceSendgridSSOIntegration()
	d := r.TestResourceData()
	d.SetId("integ-123")

	diags := resourceSendgridSSOIntegrationRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("name") != "Test IdP" {
		t.Errorf("Read() name = %v, want %q", d.Get("name"), "Test IdP")
	}
	if d.Get("enabled") != true {
		t.Errorf("Read() enabled = %v, want true", d.Get("enabled"))
	}
	if d.Get("signin_url") != "https://idp.example.com/signin" {
		t.Errorf("Read() signin_url = %v, want %q", d.Get("signin_url"), "https://idp.example.com/signin")
	}
	if d.Get("signout_url") != "https://idp.example.com/signout" {
		t.Errorf("Read() signout_url = %v, want %q", d.Get("signout_url"), "https://idp.example.com/signout")
	}
	if d.Get("entity_id") != "https://idp.example.com/12345" {
		t.Errorf("Read() entity_id = %v, want %q", d.Get("entity_id"), "https://idp.example.com/12345")
	}
	if d.Get("completed_integration") != true {
		t.Errorf("Read() completed_integration = %v, want true", d.Get("completed_integration"))
	}
	if d.Get("single_signon_url") != "https://sso.sendgrid.com/saml/integ-123" {
		t.Errorf("Read() single_signon_url = %v, want %q", d.Get("single_signon_url"), "https://sso.sendgrid.com/saml/integ-123")
	}
	if d.Get("audience_url") != "https://sso.sendgrid.com/saml/integ-123" {
		t.Errorf("Read() audience_url = %v, want %q", d.Get("audience_url"), "https://sso.sendgrid.com/saml/integ-123")
	}
}

func TestSSOIntegrationRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/sso/integrations/missing-id", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	server, config := setupSSOIntegrationMockServer(mux)
	defer server.Close()

	r := resourceSendgridSSOIntegration()
	d := r.TestResourceData()
	d.SetId("missing-id")

	diags := resourceSendgridSSOIntegrationRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestSSOIntegrationResourceSchema(t *testing.T) {
	r := resourceSendgridSSOIntegration()

	tests := []struct {
		field    string
		required bool
		computed bool
	}{
		{"name", true, false},
		{"enabled", true, false},
		{"signin_url", false, false},
		{"signout_url", false, false},
		{"entity_id", false, false},
		{"completed_integration", false, true},
		{"single_signon_url", false, true},
		{"audience_url", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			s, ok := r.Schema[tt.field]
			if !ok {
				t.Fatalf("schema missing field %q", tt.field)
			}
			if s.Required != tt.required {
				t.Errorf("%s Required = %v, want %v", tt.field, s.Required, tt.required)
			}
			if s.Computed != tt.computed {
				t.Errorf("%s Computed = %v, want %v", tt.field, s.Computed, tt.computed)
			}
		})
	}

	if r.Importer == nil {
		t.Error("resource should have an Importer configured")
	}
}
