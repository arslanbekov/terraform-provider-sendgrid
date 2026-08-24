package sendgrid

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
		_, _ = w.Write([]byte(`{"username":"jdoe","email":"jdoe@example.com","is_admin":false,"is_sso":true,"user_type":"teammate"}`))
	})
	mux.HandleFunc("/teammates/jdoe/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		// restricted entry echoes an automatic scope (must be stripped); full entry
		// echoes scopes that don't apply (must be dropped) — both would otherwise
		// produce a perpetual diff.
		_, _ = w.Write([]byte(`{
			"has_restricted_subuser_access": true,
			"subuser_access": [
				{"id": 111, "permission_type": "restricted", "scopes": ["mail.send", "2fa_required"]},
				{"id": 222, "permission_type": "admin", "scopes": ["mail.send"]}
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
	if aScopes := byID[222]["scopes"].(*schema.Set); aScopes.Len() != 0 {
		t.Errorf("admin entry scopes = %v, want empty (dropped)", aScopes.List())
	}
}

func TestTeammateRead_SubuserAccessErrorPropagates(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":[{"username":"jdoe","email":"jdoe@example.com"}]}`))
	})
	mux.HandleFunc("/teammates/jdoe", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"username":"jdoe","email":"jdoe@example.com","is_admin":false,"is_sso":true,"user_type":"teammate"}`))
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
		{ID: 2, PermissionType: "admin", Scopes: []string{"mail.send"}},
	}
	out := flattenSubuserAccess(in)
	if len(out) != 2 {
		t.Fatalf("len = %d, want 2", len(out))
	}
	if got := out[0]["scopes"].([]string); len(got) != 1 || got[0] != "mail.send" {
		t.Errorf("restricted scopes = %v, want [mail.send]", got)
	}
	if got := out[1]["scopes"].([]string); len(got) != 0 {
		t.Errorf("admin scopes = %v, want empty", got)
	}
}

// TestTeammateSubuserAccessComputedNoDiff locks in the Optional+Computed behaviour:
// when the prior state carries subuser_access (as Read populates it from the API) but
// the config manages no block, the plan must be empty rather than wanting to remove it.
func TestTeammateSubuserAccessComputedNoDiff(t *testing.T) {
	r := resourceSendgridTeammate()

	// Prior state: API returned one restricted subuser block (built through the
	// resource schema so the TypeSet hashing matches the real Set function).
	prior := r.Data(nil)
	prior.SetId("jdoe@example.com")
	_ = prior.Set("email", "jdoe@example.com")
	_ = prior.Set("is_sso", true)
	_ = prior.Set("is_admin", false)
	_ = prior.Set("scopes", []string{"user.profile.read"})
	_ = prior.Set("subuser_access", []map[string]interface{}{
		{"id": 111, "permission_type": "restricted", "scopes": []string{"mail.send"}},
	})
	state := prior.State()

	// Config manages no subuser_access block.
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"email":    "jdoe@example.com",
		"is_sso":   true,
		"is_admin": false,
	})

	diff, err := r.Diff(context.Background(), state, config, nil)
	if err != nil {
		t.Fatalf("Diff() unexpected error: %v", err)
	}
	if diff != nil && !diff.Empty() {
		t.Errorf("expected empty plan for unmanaged subuser_access (Computed), got diff: %#v", diff.Attributes)
	}
}

// TestTeammateScopesClearedWithSubuserAccess covers the case Computed alone
// doesn't handle: an explicit "scopes = []" against a teammate that already
// has subuser_access managed. Update() can never actually clear a leftover
// scope, so CustomizeDiff must drop it rather than show a diff that never
// converges.
func TestTeammateScopesRejectedWithSubuserAccess(t *testing.T) {
	r := resourceSendgridTeammate()

	prior := r.Data(nil)
	prior.SetId("jdoe@example.com")
	_ = prior.Set("email", "jdoe@example.com")
	_ = prior.Set("is_sso", true)
	_ = prior.Set("is_admin", false)
	_ = prior.Set("scopes", []string{"user.profile.read"})
	_ = prior.Set("subuser_access", []map[string]interface{}{
		{"id": 111, "permission_type": "admin"},
	})
	state := prior.State()

	subuserAccess := []interface{}{
		map[string]interface{}{"id": 111, "permission_type": "admin"},
	}

	tests := []struct {
		name    string
		scopes  interface{}
		wantErr bool
	}{
		// The API cannot revoke root scopes for this teammate, so an explicit
		// empty list must fail the plan instead of applying as a no-op.
		{name: "explicit empty scopes", scopes: []interface{}{}, wantErr: true},
		{name: "different scopes", scopes: []interface{}{"mail.send"}, wantErr: true},
		{name: "unchanged scopes", scopes: []interface{}{"user.profile.read"}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := terraform.NewResourceConfigRaw(map[string]interface{}{
				"email":          "jdoe@example.com",
				"is_sso":         true,
				"is_admin":       false,
				"scopes":         tt.scopes,
				"subuser_access": subuserAccess,
			})

			_, err := r.Diff(context.Background(), state, config, nil)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Diff() succeeded, want an error explaining that scopes cannot change alongside subuser_access")
				}
				if !strings.Contains(err.Error(), "scopes cannot be changed while subuser_access is set") {
					t.Errorf("Diff() error = %v, want the subuser_access scope explanation", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Diff() unexpected error: %v", err)
			}
		})
	}
}

func TestTeammateUnknownSubuserAccessDefersChecks(t *testing.T) {
	r := resourceSendgridTeammate()

	// The SDK's unknown-value sentinel; its own hcl2shim package is internal.
	const unknown = "74D93920-ED26-11E3-AC10-0800200C9A66"

	// subuser_access as a whole is unknown at plan time (for example
	// "subuser_access = var.blocks" where the variable is not resolved yet).
	// Neither check can decide anything on that, so the plan must not fail.
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"email":          "jdoe@example.com",
		"is_sso":         true,
		"is_admin":       false,
		"scopes":         []interface{}{"mail.send"},
		"subuser_access": unknown,
	})

	if _, err := r.Diff(context.Background(), nil, config, nil); err != nil {
		t.Fatalf("Diff() with unknown subuser_access returned %v, want the checks deferred", err)
	}
}

func TestTeammatePermissionTypeValidation(t *testing.T) {
	r := resourceSendgridTeammate()
	elem := r.Schema["subuser_access"].Elem.(*schema.Resource)
	vf := elem.Schema["permission_type"].ValidateFunc
	if vf == nil {
		t.Fatal("permission_type should have a ValidateFunc")
	}
	// Per the SendGrid API, permission_type is either "admin" or "restricted".
	for _, valid := range []string{"admin", "restricted"} {
		if _, errs := vf(valid, "permission_type"); len(errs) != 0 {
			t.Errorf("%q should be valid, got %v", valid, errs)
		}
	}
	// "full" is not a real API value and must be rejected; so must typos.
	for _, invalid := range []string{"full", "typo"} {
		if _, errs := vf(invalid, "permission_type"); len(errs) == 0 {
			t.Errorf("%q should be rejected at plan time", invalid)
		}
	}
}

func TestTeammateResourceSchema(t *testing.T) {
	r := resourceSendgridTeammate()

	tests := []struct {
		field    string
		required bool
		computed bool
	}{
		{"email", true, false},
		{"is_admin", true, false},
		{"is_sso", true, false},
		{"first_name", false, false},
		{"last_name", false, false},
		{"scopes", false, true},
		{"username", false, false},
		{"user_status", false, true},
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

// TestTeammateCreate_SubuserOnly locks in the create-time fix: a placeholder
// scope for the create call, never resent by the subuser_access follow-up.
func TestTeammateCreate_SubuserOnly(t *testing.T) {
	var createBody, patchBody map[string]interface{}

	mux := http.NewServeMux()
	mux.HandleFunc("/sso/teammates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &createBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"username":"jdoe@example.com","email":"jdoe@example.com","first_name":"John","last_name":"Doe","is_sso":true}`))
	})
	mux.HandleFunc("/sso/teammates/jdoe@example.com", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.NotFound(w, r)
			return
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &patchBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"username":"jdoe@example.com","email":"jdoe@example.com"}`))
	})
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":[{"username":"jdoe@example.com","email":"jdoe@example.com"}]}`))
	})
	mux.HandleFunc("/teammates/jdoe@example.com", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"username":"jdoe@example.com","email":"jdoe@example.com","first_name":"John","last_name":"Doe","is_admin":false,"is_sso":true,"user_type":"teammate"}`))
	})
	mux.HandleFunc("/teammates/jdoe@example.com/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"has_restricted_subuser_access":true,"subuser_access":[{"id":111,"permission_type":"admin"}]}`))
	})

	server, config := setupTeammateMockServer(mux)
	defer server.Close()

	r := resourceSendgridTeammate()
	d := r.TestResourceData()
	_ = d.Set("email", "jdoe@example.com")
	_ = d.Set("first_name", "John")
	_ = d.Set("last_name", "Doe")
	_ = d.Set("is_admin", false)
	_ = d.Set("is_sso", true)
	_ = d.Set("subuser_access", []interface{}{
		map[string]interface{}{"id": 111, "permission_type": "admin"},
	})

	if diags := resourceSendgridTeammateCreate(context.Background(), d, config); diags.HasError() {
		t.Fatalf("Create() returned unexpected error: %v", diags)
	}

	createScopes, _ := createBody["scopes"].([]interface{})
	if len(createScopes) != 1 || createScopes[0] != "user.profile.read" {
		t.Errorf("create body scopes = %v, want exactly [user.profile.read] so SendGrid accepts the create", createBody["scopes"])
	}

	if _, ok := patchBody["scopes"]; ok {
		t.Errorf("subuser_access patch body = %v, must not include scopes", patchBody)
	}
	if _, ok := patchBody["subuser_access"]; !ok {
		t.Errorf("subuser_access patch body = %v, want subuser_access", patchBody)
	}
}

// TestTeammateUpdate_SubuserAccess locks in that scopes is never resent once
// subuser_access is managed, even if a stray value is still configured.
func TestTeammateUpdate_SubuserAccess(t *testing.T) {
	var patchBody map[string]interface{}

	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":[{"username":"jdoe@example.com","email":"jdoe@example.com"}]}`))
	})
	mux.HandleFunc("/sso/teammates/jdoe@example.com", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.NotFound(w, r)
			return
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &patchBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"username":"jdoe@example.com","email":"jdoe@example.com"}`))
	})
	mux.HandleFunc("/teammates/jdoe@example.com", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"username":"jdoe@example.com","email":"jdoe@example.com","first_name":"John","last_name":"Doe","is_admin":false,"is_sso":true,"user_type":"active"}`))
	})
	mux.HandleFunc("/teammates/jdoe@example.com/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"has_restricted_subuser_access":true,"subuser_access":[{"id":111,"permission_type":"admin"}]}`))
	})

	server, config := setupTeammateMockServer(mux)
	defer server.Close()

	r := resourceSendgridTeammate()
	d := r.TestResourceData()
	d.SetId("jdoe@example.com")
	_ = d.Set("email", "jdoe@example.com")
	_ = d.Set("first_name", "John")
	_ = d.Set("last_name", "Doe")
	_ = d.Set("is_admin", false)
	_ = d.Set("is_sso", true)
	_ = d.Set("user_status", "active")
	// Stray leftover scope - must not be resent alongside subuser_access.
	_ = d.Set("scopes", []interface{}{"user.profile.read"})
	_ = d.Set("subuser_access", []interface{}{
		map[string]interface{}{"id": 111, "permission_type": "admin"},
	})

	if diags := resourceSendgridTeammateUpdate(context.Background(), d, config); diags.HasError() {
		t.Fatalf("Update() returned unexpected error: %v", diags)
	}

	if _, ok := patchBody["scopes"]; ok {
		t.Errorf("update patch body = %v, must not include scopes alongside subuser_access", patchBody)
	}
	if _, ok := patchBody["subuser_access"]; !ok {
		t.Errorf("update patch body = %v, want subuser_access", patchBody)
	}
}

func TestUnpersistedScopes(t *testing.T) {
	tests := []struct {
		name       string
		isAdmin    bool
		userStatus string
		persisted  []string
		requested  []string
		want       []string
	}{
		{
			name:       "scope dropped by sendgrid",
			userStatus: "active",
			persisted:  []string{"mail.send"},
			requested:  []string{"mail.send", "user.profile.update"},
			want:       []string{"user.profile.update"},
		},
		{
			name:       "every scope persisted",
			userStatus: "active",
			persisted:  []string{"mail.send", "templates.read"},
			requested:  []string{"templates.read", "mail.send"},
			want:       nil,
		},
		{
			name:       "pending teammate reports no scopes yet",
			userStatus: "pending",
			requested:  []string{"mail.send"},
			want:       nil,
		},
		{
			name:       "admin teammate tracks no scopes",
			isAdmin:    true,
			userStatus: "active",
			requested:  []string{"mail.send"},
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := resourceSendgridTeammate().TestResourceData()
			d.SetId("jdoe@example.com")
			if err := d.Set("is_admin", tt.isAdmin); err != nil {
				t.Fatalf("set is_admin: %v", err)
			}
			if err := d.Set("user_status", tt.userStatus); err != nil {
				t.Fatalf("set user_status: %v", err)
			}
			if err := d.Set("scopes", tt.persisted); err != nil {
				t.Fatalf("set scopes: %v", err)
			}

			if got := unpersistedScopes(d, tt.requested); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unpersistedScopes() = %v, want %v", got, tt.want)
			}

			diags := warnUnpersistedScopes(d, tt.requested)
			if len(tt.want) == 0 {
				if len(diags) != 0 {
					t.Fatalf("warnUnpersistedScopes() = %v, want no diagnostics", diags)
				}
				return
			}
			if len(diags) != 1 || diags[0].Severity != diag.Warning {
				t.Fatalf("warnUnpersistedScopes() = %v, want one warning", diags)
			}
			if !strings.Contains(diags[0].Detail, tt.want[0]) {
				t.Errorf("warning detail %q does not name %q", diags[0].Detail, tt.want[0])
			}
		})
	}
}

func TestSubuserAccessScopePolicy(t *testing.T) {
	block := func(scopes []string) []interface{} {
		return []interface{}{map[string]interface{}{
			"id":              1,
			"permission_type": "restricted",
			"scopes":          scopes,
		}}
	}

	tests := []struct {
		name        string
		scopes      []string
		wantErr     bool
		wantWritten []string
	}{
		{
			name:        "automatic scope in a block is rejected and never written",
			scopes:      []string{"mail.send", "user.profile.update"},
			wantErr:     true,
			wantWritten: []string{"mail.send"},
		},
		{
			name:        "ordinary scopes pass and are written unchanged",
			scopes:      []string{"mail.send", "templates.read"},
			wantErr:     false,
			wantWritten: []string{"mail.send", "templates.read"},
		},
		{
			name:        "block without scopes is not validated",
			scopes:      nil,
			wantErr:     false,
			wantWritten: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := resourceSendgridTeammate().TestResourceData()
			if err := d.Set("subuser_access", block(tt.scopes)); err != nil {
				t.Fatalf("set subuser_access: %v", err)
			}

			diags := validateSubuserAccessScopes(d)
			if got := diags.HasError(); got != tt.wantErr {
				t.Fatalf("validateSubuserAccessScopes() error = %v (%v), want %v", got, diags, tt.wantErr)
			}

			// The write payload must drop exactly what flattenSubuserAccess drops on
			// read, or the block re-plans forever.
			entries := extractSubuserAccess(d)
			if len(entries) != 1 {
				t.Fatalf("extractSubuserAccess() returned %d entries, want 1", len(entries))
			}
			written := append([]string(nil), entries[0].Scopes...)
			sort.Strings(written)
			want := append([]string(nil), tt.wantWritten...)
			sort.Strings(want)
			if !reflect.DeepEqual(written, want) {
				t.Errorf("written scopes = %v, want %v", written, want)
			}
		})
	}
}

// scopes became Optional+Computed for every teammate, not only those with
// subuser_access. This pins what that changed for everyone else: omitting scopes
// now keeps the teammate's current scopes instead of planning them away.
func TestTeammateOmittedScopesKeepServerValueWithoutSubuserAccess(t *testing.T) {
	r := resourceSendgridTeammate()

	prior := r.Data(nil)
	prior.SetId("jdoe@example.com")
	_ = prior.Set("email", "jdoe@example.com")
	_ = prior.Set("is_sso", true)
	_ = prior.Set("is_admin", false)
	_ = prior.Set("scopes", []string{"mail.send", "templates.read"})
	state := prior.State()

	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"email":    "jdoe@example.com",
		"is_sso":   true,
		"is_admin": false,
	})

	diff, err := r.Diff(context.Background(), state, config, nil)
	if err != nil {
		t.Fatalf("Diff() unexpected error: %v", err)
	}
	if diff != nil {
		if attr, ok := diff.Attributes["scopes.#"]; ok {
			t.Errorf("omitting scopes planned a change to them (%#v); Computed must keep the server value", attr)
		}
	}
}
