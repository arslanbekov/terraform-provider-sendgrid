package sendgrid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTemplateMockServer(handler http.Handler) (*httptest.Server, *Config) {
	server := httptest.NewServer(handler)
	config := &Config{
		APIKey: "test-api-key",
		Host:   server.URL,
	}
	return server, config
}

func TestTemplateRead_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/templates/tmpl-123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"tmpl-123","name":"My Template","generation":"dynamic","updated_at":"2024-01-01"}`))
	})

	server, config := setupTemplateMockServer(mux)
	defer server.Close()

	r := resourceSendgridTemplate()
	d := r.TestResourceData()
	d.SetId("tmpl-123")

	diags := resourceSendgridTemplateRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("name") != "My Template" {
		t.Errorf("Read() name = %v, want %q", d.Get("name"), "My Template")
	}
	if d.Get("generation") != "dynamic" {
		t.Errorf("Read() generation = %v, want %q", d.Get("generation"), "dynamic")
	}
	if d.Get("updated_at") != "2024-01-01" {
		t.Errorf("Read() updated_at = %v, want %q", d.Get("updated_at"), "2024-01-01")
	}
}

func TestTemplateRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/templates/tmpl-missing", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	server, config := setupTemplateMockServer(mux)
	defer server.Close()

	r := resourceSendgridTemplate()
	d := r.TestResourceData()
	d.SetId("tmpl-missing")

	diags := resourceSendgridTemplateRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestTemplateResourceSchema(t *testing.T) {
	r := resourceSendgridTemplate()

	tests := []struct {
		field    string
		required bool
		forceNew bool
	}{
		{"name", true, false},
		{"generation", false, true},
		{"updated_at", false, false},
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

	if r.Importer == nil {
		t.Error("resource should have an Importer configured")
	}
}
