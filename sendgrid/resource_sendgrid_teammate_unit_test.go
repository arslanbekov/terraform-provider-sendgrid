package sendgrid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func setupTeammateMockServer(handler http.Handler) (*httptest.Server, *Config) {
	server := httptest.NewServer(handler)
	config := &Config{
		APIKey: "test-api-key",
		Host:   server.URL,
	}
	return server, config
}

func TestTeammateRead_Success(t *testing.T) {
	mux := http.NewServeMux()

	// Mock GET /teammates?limit=10000 — returns list with matching email/username
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		query := r.URL.RawQuery

		// GET /teammates?limit=10000 — email lookup
		if path == "/teammates" && strings.Contains(query, "limit=10000") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":[{"username":"jdoe","email":"jdoe@example.com","first_name":"John","last_name":"Doe"}]}`))
			return
		}

		http.NotFound(w, r)
	})

	// Mock GET /teammates/jdoe — returns full user
	mux.HandleFunc("/teammates/jdoe", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"username":"jdoe",
			"email":"jdoe@example.com",
			"first_name":"John",
			"last_name":"Doe",
			"scopes":["mail.send","templates.read"],
			"is_admin":false,
			"user_type":"teammate"
		}`))
	})

	server, config := setupTeammateMockServer(mux)
	defer server.Close()

	r := resourceSendgridTeammate()
	d := r.TestResourceData()
	d.SetId("jdoe@example.com")

	diags := resourceSendgridTeammateRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("email") != "jdoe@example.com" {
		t.Errorf("Read() email = %v, want %q", d.Get("email"), "jdoe@example.com")
	}
	if d.Get("username") != "jdoe" {
		t.Errorf("Read() username = %v, want %q", d.Get("username"), "jdoe")
	}
	if d.Get("first_name") != "John" {
		t.Errorf("Read() first_name = %v, want %q", d.Get("first_name"), "John")
	}
	if d.Get("last_name") != "Doe" {
		t.Errorf("Read() last_name = %v, want %q", d.Get("last_name"), "Doe")
	}
	if d.Get("is_admin") != false {
		t.Errorf("Read() is_admin = %v, want false", d.Get("is_admin"))
	}
	if d.Get("user_status") != "active" {
		t.Errorf("Read() user_status = %v, want %q", d.Get("user_status"), "active")
	}

	scopes := d.Get("scopes").(*schema.Set)
	if scopes.Len() != 2 {
		t.Errorf("Read() scopes length = %d, want 2", scopes.Len())
	}
}

func TestTeammateRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()

	// Mock GET /teammates?limit=10000 — no matching email
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		query := r.URL.RawQuery

		if path == "/teammates" && strings.Contains(query, "limit=10000") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":[]}`))
			return
		}

		http.NotFound(w, r)
	})

	// Mock GET /teammates/pending?limit=10000 — no pending either
	mux.HandleFunc("/teammates/pending", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":[]}`))
	})

	server, config := setupTeammateMockServer(mux)
	defer server.Close()

	r := resourceSendgridTeammate()
	d := r.TestResourceData()
	d.SetId("missing@example.com")

	diags := resourceSendgridTeammateRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestTeammateResourceSchema(t *testing.T) {
	r := resourceSendgridTeammate()

	tests := []struct {
		field    string
		required bool
	}{
		{"email", true},
		{"is_admin", true},
		{"is_sso", true},
		{"first_name", false},
		{"last_name", false},
		{"scopes", false},
		{"username", false},
		{"user_status", false},
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
