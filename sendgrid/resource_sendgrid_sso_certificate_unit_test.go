package sendgrid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupSSOCertificateMockServer(handler http.Handler) (*httptest.Server, *Config) {
	server := httptest.NewServer(handler)
	config := &Config{
		APIKey: "test-api-key",
		Host:   server.URL,
	}
	return server, config
}

func TestSSOCertificateRead_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/sso/certificates/123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":123,"public_certificate":"-----BEGIN CERTIFICATE-----\nTEST\n-----END CERTIFICATE-----","integration_id":"integ-456"}`))
	})

	server, config := setupSSOCertificateMockServer(mux)
	defer server.Close()

	r := resourceSendgridSSOCertificate()
	d := r.TestResourceData()
	d.SetId("123")

	diags := resourceSendgridSSOCertificateRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("public_certificate") != "-----BEGIN CERTIFICATE-----\nTEST\n-----END CERTIFICATE-----" {
		t.Errorf("Read() public_certificate = %v, want certificate PEM", d.Get("public_certificate"))
	}
	if d.Get("integration_id") != "integ-456" {
		t.Errorf("Read() integration_id = %v, want %q", d.Get("integration_id"), "integ-456")
	}
}

func TestSSOCertificateRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/sso/certificates/999", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	server, config := setupSSOCertificateMockServer(mux)
	defer server.Close()

	r := resourceSendgridSSOCertificate()
	d := r.TestResourceData()
	d.SetId("999")

	diags := resourceSendgridSSOCertificateRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestSSOCertificateResourceSchema(t *testing.T) {
	r := resourceSendgridSSOCertificate()

	tests := []struct {
		field    string
		required bool
	}{
		{"public_certificate", true},
		{"integration_id", true},
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
		})
	}

	if r.Importer == nil {
		t.Error("resource should have an Importer configured")
	}
}
