package sendgrid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
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

func TestTeammateRead_SubuserAccess(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "limit=10000") {
			_, _ = w.Write([]byte(`{"result":[{"username":"jdoe","email":"jdoe@example.com"}]}`))
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/teammates/jdoe", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"username":"jdoe","email":"jdoe@example.com","is_admin":false,"user_type":"teammate"}`))
	})
	mux.HandleFunc("/teammates/jdoe/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		// restricted entry echoes an automatic scope (must be stripped); full entry
		// echoes scopes that don't apply (must be dropped) — both would otherwise
		// produce a perpetual diff.
		_, _ = w.Write([]byte(`{
			"has_restricted_subuser_access": true,
			"subuser_access": [
				{"id": 111, "permission_type": "restricted", "scopes": ["mail.send", "2fa_required"]},
				{"id": 222, "permission_type": "full", "scopes": ["mail.send"]}
			]
		}`))
	})
	server, config := setupTeammateMockServer(mux)
	defer server.Close()

	r := resourceSendgridTeammate()
	d := r.TestResourceData()
	d.SetId("jdoe@example.com")
	_ = d.Set("is_sso", true)

	if diags := resourceSendgridTeammateRead(context.Background(), d, config); diags.HasError() {
		t.Fatalf("Read() unexpected error: %v", diags)
	}

	access := d.Get("subuser_access").(*schema.Set).List()
	if len(access) != 2 {
		t.Fatalf("subuser_access length = %d, want 2", len(access))
	}
	byID := map[int]map[string]interface{}{}
	for _, a := range access {
		m := a.(map[string]interface{})
		byID[m["id"].(int)] = m
	}

	rScopes := byID[111]["scopes"].(*schema.Set)
	if rScopes.Len() != 1 || !rScopes.Contains("mail.send") {
		t.Errorf("restricted scopes = %v, want [mail.send] (automatic scope stripped)", rScopes.List())
	}
	if fScopes := byID[222]["scopes"].(*schema.Set); fScopes.Len() != 0 {
		t.Errorf("full entry scopes = %v, want empty (dropped)", fScopes.List())
	}
}

func TestTeammateRead_SubuserAccessErrorPropagates(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":[{"username":"jdoe","email":"jdoe@example.com"}]}`))
	})
	mux.HandleFunc("/teammates/jdoe", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"username":"jdoe","email":"jdoe@example.com","is_admin":false,"user_type":"teammate"}`))
	})
	mux.HandleFunc("/teammates/jdoe/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"message":"boom"}]}`))
	})
	server, config := setupTeammateMockServer(mux)
	defer server.Close()

	r := resourceSendgridTeammate()
	d := r.TestResourceData()
	d.SetId("jdoe@example.com")
	_ = d.Set("is_sso", true)

	if diags := resourceSendgridTeammateRead(context.Background(), d, config); !diags.HasError() {
		t.Fatal("expected Read() to surface the 5xx from subuser_access rather than masking it")
	}
}

func TestFlattenSubuserAccess(t *testing.T) {
	in := []sendgrid.SubuserAccessRead{
		{ID: 1, PermissionType: "restricted", Scopes: []string{"mail.send", "2fa_exempt"}},
		{ID: 2, PermissionType: "full", Scopes: []string{"mail.send"}},
	}
	out := flattenSubuserAccess(in)
	if len(out) != 2 {
		t.Fatalf("len = %d, want 2", len(out))
	}
	if got := out[0]["scopes"].([]string); len(got) != 1 || got[0] != "mail.send" {
		t.Errorf("restricted scopes = %v, want [mail.send]", got)
	}
	if got := out[1]["scopes"].([]string); len(got) != 0 {
		t.Errorf("full scopes = %v, want empty", got)
	}
}

func TestTeammatePermissionTypeValidation(t *testing.T) {
	r := resourceSendgridTeammate()
	elem := r.Schema["subuser_access"].Elem.(*schema.Resource)
	vf := elem.Schema["permission_type"].ValidateFunc
	if vf == nil {
		t.Fatal("permission_type should have a ValidateFunc")
	}
	if _, errs := vf("restricted", "permission_type"); len(errs) != 0 {
		t.Errorf("'restricted' should be valid, got %v", errs)
	}
	if _, errs := vf("typo", "permission_type"); len(errs) == 0 {
		t.Error("'typo' should be rejected at plan time")
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
