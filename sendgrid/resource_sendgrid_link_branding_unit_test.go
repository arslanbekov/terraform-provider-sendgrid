package sendgrid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupLinkBrandingMockServer(handler http.Handler) (*httptest.Server, *Config) {
	server := httptest.NewServer(handler)
	config := &Config{
		APIKey: "test-api-key",
		Host:   server.URL,
	}
	return server, config
}

func TestLinkBrandingRead_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/links/456", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": 456,
			"domain": "example.com",
			"subdomain": "mail",
			"username": "testuser",
			"default": true,
			"valid": true,
			"dns": {
				"domain_cname": {
					"valid": true,
					"type": "cname",
					"host": "mail.example.com",
					"data": "sendgrid.net"
				},
				"owner_cname": {
					"valid": true,
					"type": "cname",
					"host": "owner.example.com",
					"data": "sendgrid.net"
				}
			}
		}`))
	})

	server, config := setupLinkBrandingMockServer(mux)
	defer server.Close()

	r := resourceSendgridLinkBranding()
	d := r.TestResourceData()
	d.SetId("456")

	diags := resourceSendgridLinkBrandingRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("domain") != "example.com" {
		t.Errorf("Read() domain = %v, want %q", d.Get("domain"), "example.com")
	}
	if d.Get("subdomain") != "mail" {
		t.Errorf("Read() subdomain = %v, want %q", d.Get("subdomain"), "mail")
	}
	if d.Get("username") != "testuser" {
		t.Errorf("Read() username = %v, want %q", d.Get("username"), "testuser")
	}
	if d.Get("is_default") != true {
		t.Errorf("Read() is_default = %v, want true", d.Get("is_default"))
	}
	if d.Get("valid") != true {
		t.Errorf("Read() valid = %v, want true", d.Get("valid"))
	}

	dns := d.Get("dns").([]interface{})
	if len(dns) != 2 {
		t.Fatalf("Read() dns length = %d, want 2", len(dns))
	}

	first := dns[0].(map[string]interface{})
	if first["type"] != "cname" {
		t.Errorf("Read() dns[0].type = %v, want %q", first["type"], "cname")
	}
	if first["host"] != "mail.example.com" {
		t.Errorf("Read() dns[0].host = %v, want %q", first["host"], "mail.example.com")
	}
}

func TestLinkBrandingRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/links/missing", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	server, config := setupLinkBrandingMockServer(mux)
	defer server.Close()

	r := resourceSendgridLinkBranding()
	d := r.TestResourceData()
	d.SetId("missing")

	diags := resourceSendgridLinkBrandingRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestLinkBrandingResourceSchema(t *testing.T) {
	r := resourceSendgridLinkBranding()

	tests := []struct {
		field    string
		required bool
		optional bool
		forceNew bool
	}{
		{"domain", true, false, true},
		{"subdomain", false, true, true},
		{"username", false, false, false},
		{"is_default", false, true, false},
		{"valid", false, true, false},
		{"dns", false, false, false},
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
			if s.Optional != tt.optional {
				t.Errorf("%s Optional = %v, want %v", tt.field, s.Optional, tt.optional)
			}
			if s.ForceNew != tt.forceNew {
				t.Errorf("%s ForceNew = %v, want %v", tt.field, s.ForceNew, tt.forceNew)
			}
		})
	}

	if r.Importer == nil {
		t.Error("resource should have an Importer configured")
	}
}
