package sendgrid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func setupAPIKeyMockServer(handler http.Handler) (*httptest.Server, *Config) {
	server := httptest.NewServer(handler)
	config := &Config{
		APIKey: "test-api-key",
		Host:   server.URL,
	}
	return server, config
}

func TestAPIKeyRead_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api_keys/123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"test","api_key_id":"123","scopes":["mail.send"]}`))
	})

	server, config := setupAPIKeyMockServer(mux)
	defer server.Close()

	r := resourceSendgridAPIKey()
	d := r.TestResourceData()
	d.SetId("123")

	diags := resourceSendgridAPIKeyRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("name") != "test" {
		t.Errorf("Read() name = %v, want %q", d.Get("name"), "test")
	}
	scopes := d.Get("scopes").(*schema.Set)
	if scopes.Len() != 1 || !scopes.Contains("mail.send") {
		t.Errorf("Read() scopes = %v, want [mail.send]", scopes.List())
	}
}

func TestAPIKeyRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api_keys/missing", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	server, config := setupAPIKeyMockServer(mux)
	defer server.Close()

	r := resourceSendgridAPIKey()
	d := r.TestResourceData()
	d.SetId("missing")

	diags := resourceSendgridAPIKeyRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestAPIKeyRead_ServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api_keys/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"message":"server error"}]}`))
	})

	server, config := setupAPIKeyMockServer(mux)
	defer server.Close()

	r := resourceSendgridAPIKey()
	d := r.TestResourceData()
	d.SetId("error")

	diags := resourceSendgridAPIKeyRead(context.Background(), d, config)

	if !diags.HasError() {
		t.Fatal("Read() expected error for 500, got nil")
	}
	if d.Id() == "" {
		t.Error("Read() should NOT clear ID on non-404 error")
	}
}

func TestAPIKeyResourceSchema(t *testing.T) {
	r := resourceSendgridAPIKey()

	tests := []struct {
		field    string
		required bool
		optional bool
		forceNew bool
	}{
		{"name", true, false, false},
		{"scopes", false, true, false},
		{"sub_user_on_behalf_of", false, true, false},
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
