package sendgrid_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
)

func TestClient_AssociateDomainAuthenticationSubuser(t *testing.T) {
	teardown := setup()
	defer teardown()
	mux.HandleFunc("/whitelabel/domains/123/subuser", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"username":"team-a"}` {
			t.Errorf("unexpected body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":123,"domain":"example.com","subdomain":"em","username":"team-a","valid":true}`))
	})
	client := sendgrid.NewClient("api-key", server.URL, "")

	auth, reqErr := client.AssociateDomainAuthenticationSubuser(context.Background(), "123", "team-a")

	if reqErr.Err != nil {
		t.Fatalf("unexpected error: %v", reqErr.Err)
	}
	if auth.ID != 123 || auth.Domain != "example.com" {
		t.Errorf("unexpected domain: %+v", auth)
	}
}

func TestClient_AssociateDomainAuthenticationSubuser_Error(t *testing.T) {
	teardown := setup()
	defer teardown()
	mux.HandleFunc("/whitelabel/domains/123/subuser", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":[{"message":"domain not validated"}]}`))
	})
	client := sendgrid.NewClient("api-key", server.URL, "")

	_, reqErr := client.AssociateDomainAuthenticationSubuser(context.Background(), "123", "team-a")

	if reqErr.Err == nil {
		t.Fatal("expected error for 400, got nil")
	}
	if reqErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", reqErr.StatusCode, http.StatusBadRequest)
	}
}

func TestClient_AssociateDomainAuthenticationSubuser_MissingArgs(t *testing.T) {
	client := sendgrid.NewClient("api-key", "http://localhost", "")

	if _, reqErr := client.AssociateDomainAuthenticationSubuser(context.Background(), "", "team-a"); reqErr.Err == nil {
		t.Error("expected error for empty domain ID")
	}
	if _, reqErr := client.AssociateDomainAuthenticationSubuser(context.Background(), "123", ""); reqErr.Err == nil {
		t.Error("expected error for empty username")
	}
}

func TestClient_ReadDomainAuthenticationSubuser(t *testing.T) {
	teardown := setup()
	defer teardown()
	mux.HandleFunc("/whitelabel/domains/subuser", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if got := r.URL.Query().Get("username"); got != "team a" {
			t.Errorf("username query = %q, want %q", got, "team a")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":123,"domain":"example.com","subdomain":"em","valid":true}`))
	})
	client := sendgrid.NewClient("api-key", server.URL, "")

	auth, reqErr := client.ReadDomainAuthenticationSubuser(context.Background(), "team a")

	if reqErr.Err != nil {
		t.Fatalf("unexpected error: %v", reqErr.Err)
	}
	if auth.ID != 123 || auth.Subdomain != "em" || !auth.Valid {
		t.Errorf("unexpected domain: %+v", auth)
	}
}

func TestClient_ReadDomainAuthenticationSubuser_NotFound(t *testing.T) {
	teardown := setup()
	defer teardown()
	mux.HandleFunc("/whitelabel/domains/subuser", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})
	client := sendgrid.NewClient("api-key", server.URL, "")

	_, reqErr := client.ReadDomainAuthenticationSubuser(context.Background(), "team-a")

	if reqErr.Err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	if reqErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", reqErr.StatusCode, http.StatusNotFound)
	}
}

func TestClient_DisassociateDomainAuthenticationSubuser(t *testing.T) {
	teardown := setupDisassociate(t, http.StatusNoContent)
	defer teardown()
	client := sendgrid.NewClient("api-key", server.URL, "")

	ok, reqErr := client.DisassociateDomainAuthenticationSubuser(context.Background(), "123", "team-a")

	if reqErr.Err != nil || !ok {
		t.Fatalf("unexpected result: ok=%v err=%v", ok, reqErr.Err)
	}
}

func TestClient_DisassociateDomainAuthenticationSubuser_IgnoresNotFound(t *testing.T) {
	teardown := setupDisassociate(t, http.StatusNotFound)
	defer teardown()
	client := sendgrid.NewClient("api-key", server.URL, "")

	ok, reqErr := client.DisassociateDomainAuthenticationSubuser(context.Background(), "123", "team-a")

	if reqErr.Err != nil || !ok {
		t.Fatalf("404 should be ignored, got ok=%v err=%v", ok, reqErr.Err)
	}
}

func TestClient_DisassociateDomainAuthenticationSubuser_MissingArgs(t *testing.T) {
	client := sendgrid.NewClient("api-key", "http://localhost", "")

	if _, reqErr := client.DisassociateDomainAuthenticationSubuser(context.Background(), "", "team-a"); reqErr.Err == nil {
		t.Error("expected error for empty domain ID")
	}
	if _, reqErr := client.DisassociateDomainAuthenticationSubuser(context.Background(), "123", ""); reqErr.Err == nil {
		t.Error("expected error for empty username")
	}
}

func TestClient_DisassociateDomainAuthenticationSubuser_Error(t *testing.T) {
	teardown := setupDisassociate(t, http.StatusForbidden)
	defer teardown()
	client := sendgrid.NewClient("api-key", server.URL, "")

	_, reqErr := client.DisassociateDomainAuthenticationSubuser(context.Background(), "123", "team-a")

	if reqErr.Err == nil {
		t.Fatal("expected error for 403, got nil")
	}
}

// setupDisassociate registers a DELETE /whitelabel/domains/123/subuser handler answering with status.
func setupDisassociate(t *testing.T, status int) func() {
	t.Helper()
	teardown := setup()
	mux.HandleFunc("/whitelabel/domains/123/subuser", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if got := r.URL.Query().Get("username"); got != "team-a" {
			t.Errorf("username query = %q, want %q", got, "team-a")
		}
		w.WriteHeader(status)
	})

	return teardown
}

func TestClient_ReadDomainAuthentications(t *testing.T) {
	teardown := setup()
	defer teardown()
	mux.HandleFunc("/whitelabel/domains", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("domain") != "example.com" || q.Get("exclude_subusers") != "true" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":1,"domain":"example.com","subdomain":"em"},{"id":2,"domain":"app.example.com","subdomain":"em"}]`))
	})
	client := sendgrid.NewClient("api-key", server.URL, "")

	auths, reqErr := client.ReadDomainAuthentications(context.Background(), "example.com", true)

	if reqErr.Err != nil {
		t.Fatalf("unexpected error: %v", reqErr.Err)
	}
	if len(auths) != 2 || auths[1].Domain != "app.example.com" {
		t.Errorf("unexpected domains: %+v", auths)
	}
}

func TestClient_ReadDomainAuthentications_NoQuery(t *testing.T) {
	teardown := setup()
	defer teardown()
	mux.HandleFunc("/whitelabel/domains", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query, got %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})
	client := sendgrid.NewClient("api-key", server.URL, "")

	auths, reqErr := client.ReadDomainAuthentications(context.Background(), "", false)

	if reqErr.Err != nil || len(auths) != 0 {
		t.Fatalf("unexpected result: %+v %v", auths, reqErr.Err)
	}
}
