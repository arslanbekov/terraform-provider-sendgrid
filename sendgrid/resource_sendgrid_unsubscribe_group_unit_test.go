package sendgrid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupUnsubscribeGroupMockServer(handler http.Handler) (*httptest.Server, *Config) {
	server := httptest.NewServer(handler)
	config := &Config{
		APIKey: "test-api-key",
		Host:   server.URL,
	}
	return server, config
}

func TestUnsubscribeGroupRead_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/asm/groups/789", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":789,"name":"test-group","description":"A test group","is_default":true,"unsubscribes":42}`))
	})

	server, config := setupUnsubscribeGroupMockServer(mux)
	defer server.Close()

	r := resourceSendgridUnsubscribeGroup()
	d := r.TestResourceData()
	d.SetId("789")

	diags := resourceSendgridUnsubscribeGroupRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("name") != "test-group" {
		t.Errorf("Read() name = %v, want %q", d.Get("name"), "test-group")
	}
	if d.Get("description") != "A test group" {
		t.Errorf("Read() description = %v, want %q", d.Get("description"), "A test group")
	}
	if d.Get("is_default") != true {
		t.Errorf("Read() is_default = %v, want true", d.Get("is_default"))
	}
	if d.Get("unsubscribes") != 42 {
		t.Errorf("Read() unsubscribes = %v, want 42", d.Get("unsubscribes"))
	}
}

func TestUnsubscribeGroupRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/asm/groups/missing", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	server, config := setupUnsubscribeGroupMockServer(mux)
	defer server.Close()

	r := resourceSendgridUnsubscribeGroup()
	d := r.TestResourceData()
	d.SetId("missing")

	diags := resourceSendgridUnsubscribeGroupRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestUnsubscribeGroupResourceSchema(t *testing.T) {
	r := resourceSendgridUnsubscribeGroup()

	tests := []struct {
		field    string
		required bool
		optional bool
		forceNew bool
	}{
		{"name", true, false, false},
		{"description", false, true, false},
		{"is_default", false, true, false},
		{"unsubscribes", false, false, false},
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
