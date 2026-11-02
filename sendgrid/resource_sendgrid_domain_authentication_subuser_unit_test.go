package sendgrid

import (
	"context"
	"io"
	"net/http"
	"testing"
)

const subuserDomainJSON = `{"id":123,"domain":"example.com","subdomain":"em","username":"team-a","valid":true}`

func TestDomainAuthenticationSubuserCreate_Success(t *testing.T) {
	var postBody string
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/123/subuser", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		postBody = string(body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(subuserDomainJSON))
	})
	mux.HandleFunc("/whitelabel/domains/subuser", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(subuserDomainJSON))
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	d := resourceSendgridDomainAuthenticationSubuser().TestResourceData()
	_ = d.Set("domain_id", "123")
	_ = d.Set("username", "team-a")

	diags := resourceSendgridDomainAuthenticationSubuserCreate(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Create() returned unexpected error: %v", diags)
	}
	if postBody != `{"username":"team-a"}` {
		t.Errorf("Create() body = %s", postBody)
	}
	if d.Id() != "123/team-a" {
		t.Errorf("Create() ID = %q, want %q", d.Id(), "123/team-a")
	}
	if d.Get("domain") != "example.com" || d.Get("subdomain") != "em" || d.Get("valid") != true {
		t.Errorf("Create() unexpected computed attrs: %v %v %v", d.Get("domain"), d.Get("subdomain"), d.Get("valid"))
	}
}

func TestDomainAuthenticationSubuserCreate_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/123/subuser", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":[{"message":"domain must be validated"}]}`))
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	d := resourceSendgridDomainAuthenticationSubuser().TestResourceData()
	_ = d.Set("domain_id", "123")
	_ = d.Set("username", "team-a")

	diags := resourceSendgridDomainAuthenticationSubuserCreate(context.Background(), d, config)

	if !diags.HasError() {
		t.Fatal("Create() expected error for 400, got nil")
	}
	if d.Id() != "" {
		t.Errorf("Create() should not set ID on failure, got %q", d.Id())
	}
}

func TestDomainAuthenticationSubuserRead_Success(t *testing.T) {
	var username string
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/subuser", func(w http.ResponseWriter, r *http.Request) {
		username = r.URL.Query().Get("username")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(subuserDomainJSON))
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	d := resourceSendgridDomainAuthenticationSubuser().TestResourceData()
	d.SetId("123/team-a")

	diags := resourceSendgridDomainAuthenticationSubuserRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if username != "team-a" {
		t.Errorf("Read() queried username %q, want %q", username, "team-a")
	}
	if d.Get("domain_id") != "123" || d.Get("username") != "team-a" {
		t.Errorf("Read() domain_id=%v username=%v", d.Get("domain_id"), d.Get("username"))
	}
}

func TestDomainAuthenticationSubuserRead_ImportByUsername(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/subuser", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(subuserDomainJSON))
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	d := resourceSendgridDomainAuthenticationSubuser().TestResourceData()
	d.SetId("team-a")

	diags := resourceSendgridDomainAuthenticationSubuserRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Id() != "123/team-a" {
		t.Errorf("Read() should normalise ID, got %q", d.Id())
	}
	if d.Get("domain_id") != "123" || d.Get("username") != "team-a" {
		t.Errorf("Read() domain_id=%v username=%v", d.Get("domain_id"), d.Get("username"))
	}
}

func TestDomainAuthenticationSubuserRead_Drift(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/subuser", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":999,"domain":"other.example.com","subdomain":"em"}`))
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	d := resourceSendgridDomainAuthenticationSubuser().TestResourceData()
	d.SetId("123/team-a")
	_ = d.Set("domain_id", "123")

	diags := resourceSendgridDomainAuthenticationSubuserRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("domain_id") != "999" {
		t.Errorf("Read() should report the actual domain_id, got %v", d.Get("domain_id"))
	}
	if d.Id() != "123/team-a" {
		t.Errorf("Read() should keep ID, got %q", d.Id())
	}
}

func TestDomainAuthenticationSubuserRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/subuser", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	d := resourceSendgridDomainAuthenticationSubuser().TestResourceData()
	d.SetId("123/team-a")

	diags := resourceSendgridDomainAuthenticationSubuserRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestDomainAuthenticationSubuserRead_NoAssociation(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/subuser", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	d := resourceSendgridDomainAuthenticationSubuser().TestResourceData()
	d.SetId("123/team-a")

	diags := resourceSendgridDomainAuthenticationSubuserRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID when no domain is assigned, got: %s", d.Id())
	}
}

func TestDomainAuthenticationSubuserRead_ServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/subuser", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	d := resourceSendgridDomainAuthenticationSubuser().TestResourceData()
	d.SetId("123/team-a")

	diags := resourceSendgridDomainAuthenticationSubuserRead(context.Background(), d, config)

	if !diags.HasError() {
		t.Fatal("Read() expected error for 500, got nil")
	}
}

func TestDomainAuthenticationSubuserDelete(t *testing.T) {
	var method, username string
	mux := http.NewServeMux()
	// Only the association held in state (domain 999) may be removed.
	mux.HandleFunc("/whitelabel/domains/999/subuser", func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		username = r.URL.Query().Get("username")
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/whitelabel/domains/subuser", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("Delete() must not call the unscoped %s /whitelabel/domains/subuser", r.Method)
		w.WriteHeader(http.StatusNoContent)
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	d := resourceSendgridDomainAuthenticationSubuser().TestResourceData()
	d.SetId("123/team-a")
	_ = d.Set("domain_id", "999")
	_ = d.Set("username", "team-a")

	diags := resourceSendgridDomainAuthenticationSubuserDelete(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Delete() returned unexpected error: %v", diags)
	}
	if method != http.MethodDelete || username != "team-a" {
		t.Errorf("Delete() sent %s username=%q", method, username)
	}
}

func TestDomainAuthenticationSubuserDelete_Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/123/subuser", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	d := resourceSendgridDomainAuthenticationSubuser().TestResourceData()
	d.SetId("123/team-a")
	_ = d.Set("domain_id", "123")
	_ = d.Set("username", "team-a")

	diags := resourceSendgridDomainAuthenticationSubuserDelete(context.Background(), d, config)

	if !diags.HasError() {
		t.Fatal("Delete() expected error for 403, got nil")
	}
}

func TestDomainAuthenticationSubuserResourceSchema(t *testing.T) {
	r := resourceSendgridDomainAuthenticationSubuser()

	for _, field := range []string{"domain_id", "username"} {
		s := r.Schema[field]
		if s == nil || !s.Required || !s.ForceNew {
			t.Errorf("%s should be Required and ForceNew: %+v", field, s)
		}
	}

	for _, field := range []string{"domain", "subdomain", "valid"} {
		s := r.Schema[field]
		if s == nil || !s.Computed {
			t.Errorf("%s should be Computed: %+v", field, s)
		}
	}

	if r.UpdateContext != nil {
		t.Error("resource must not have an update, both arguments force replacement")
	}
	if r.Importer == nil {
		t.Error("resource should support import")
	}
}
