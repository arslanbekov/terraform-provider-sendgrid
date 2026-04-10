package sendgrid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupDomainAuthenticationMockServer(handler http.Handler) (*httptest.Server, *Config) {
	server := httptest.NewServer(handler)
	config := &Config{
		APIKey: "test-api-key",
		Host:   server.URL,
	}
	return server, config
}

func TestDomainAuthenticationRead_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id":123,
			"domain":"example.com",
			"subdomain":"mail",
			"username":"user",
			"default":false,
			"custom_spf":false,
			"automatic_security":true,
			"valid":true,
			"dns":{}
		}`))
	})

	server, config := setupDomainAuthenticationMockServer(mux)
	defer server.Close()

	r := resourceSendgridDomainAuthentication()
	d := r.TestResourceData()
	d.SetId("123")

	diags := resourceSendgridDomainAuthenticationRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("domain") != "example.com" {
		t.Errorf("Read() domain = %v, want %q", d.Get("domain"), "example.com")
	}
	if d.Get("subdomain") != "mail" {
		t.Errorf("Read() subdomain = %v, want %q", d.Get("subdomain"), "mail")
	}
	if d.Get("username") != "user" {
		t.Errorf("Read() username = %v, want %q", d.Get("username"), "user")
	}
	if d.Get("is_default") != false {
		t.Errorf("Read() is_default = %v, want false", d.Get("is_default"))
	}
	if d.Get("custom_spf") != false {
		t.Errorf("Read() custom_spf = %v, want false", d.Get("custom_spf"))
	}
	if d.Get("valid") != true {
		t.Errorf("Read() valid = %v, want true", d.Get("valid"))
	}
}

func TestDomainAuthenticationRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/999", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	server, config := setupDomainAuthenticationMockServer(mux)
	defer server.Close()

	r := resourceSendgridDomainAuthentication()
	d := r.TestResourceData()
	d.SetId("999")

	diags := resourceSendgridDomainAuthenticationRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestDomainAuthenticationResourceSchema(t *testing.T) {
	r := resourceSendgridDomainAuthentication()

	tests := []struct {
		field    string
		required bool
		forceNew bool
	}{
		{"domain", true, true},
		{"subdomain", false, true},
		{"username", false, false},
		{"is_default", false, false},
		{"custom_spf", false, false},
		{"valid", false, false},
		{"ips", false, true},
		{"automatic_security", false, true},
		{"custom_dkim_selector", false, true},
		{"sub_user_on_behalf_of", false, false},
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
			if s.ForceNew != tt.forceNew {
				t.Errorf("%s ForceNew = %v, want %v", tt.field, s.ForceNew, tt.forceNew)
			}
		})
	}

	// Verify dns is a computed list
	dns, ok := r.Schema["dns"]
	if !ok {
		t.Fatal("schema missing field \"dns\"")
	}
	if !dns.Computed {
		t.Error("dns should be Computed")
	}

	if r.Importer == nil {
		t.Error("resource should have an Importer configured")
	}
}
