package sendgrid_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
)

// teammateListBody is the GET /teammates?limit=10000 lookup payload used to
// resolve an email to a username before the SSO PATCH / subuser_access GET.
const teammateListBody = `{"result":[{"username":"jdoe","email":"jdoe@example.com"}]}`

func TestClient_ReadSubuserAccess_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(teammateListBody))
	})
	mux.HandleFunc("/teammates/jdoe/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"has_restricted_subuser_access": true,
			"subuser_access": [
				{"id": 111, "username": "sub_a", "permission_type": "restricted", "scopes": ["mail.send"]},
				{"id": 222, "username": "sub_b", "permission_type": "full"}
			]
		}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := sendgrid.NewClient("api-key", server.URL, "")
	resp, reqErr := client.ReadSubuserAccess(context.Background(), "jdoe@example.com")
	if reqErr.Err != nil {
		t.Fatalf("ReadSubuserAccess() unexpected error: %v", reqErr.Err)
	}
	if !resp.HasRestrictedSubuserAccess {
		t.Error("expected HasRestrictedSubuserAccess = true")
	}
	if len(resp.SubuserAccess) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp.SubuserAccess))
	}
	if resp.SubuserAccess[0].ID != 111 || resp.SubuserAccess[0].PermissionType != "restricted" {
		t.Errorf("unexpected first entry: %+v", resp.SubuserAccess[0])
	}
}

func TestClient_ReadSubuserAccess_PropagatesStatusCode(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(teammateListBody))
	})
	mux.HandleFunc("/teammates/jdoe/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"message":"unauthorized"}]}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := sendgrid.NewClient("api-key", server.URL, "")
	_, reqErr := client.ReadSubuserAccess(context.Background(), "jdoe@example.com")
	if reqErr.Err == nil {
		t.Fatal("expected an error for a 401 response")
	}
	if reqErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d (must not be masked)", reqErr.StatusCode, http.StatusUnauthorized)
	}
}

// captureSSOBody resolves the email then captures the PATCH body sent to
// /sso/teammates/{username}, returning it as a decoded map.
func captureSSOBody(t *testing.T, call func(c *sendgrid.Client)) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}

	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(teammateListBody))
	})
	mux.HandleFunc("/sso/teammates/jdoe", func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("failed decoding request body %q: %v", string(raw), err)
		}
		_, _ = w.Write([]byte(`{"email":"jdoe@example.com"}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	call(sendgrid.NewClient("api-key", server.URL, ""))
	return body
}

// Blocker-1 regression: with no managed blocks and is_admin=false, neither
// has_restricted_subuser_access nor is_admin may be sent, so existing
// out-of-band admin / restricted-access state is never silently cleared.
func TestClient_UpdateSSOUser_OmitsFlagsWhenUnmanaged(t *testing.T) {
	body := captureSSOBody(t, func(c *sendgrid.Client) {
		_, reqErr := c.UpdateSSOUser(context.Background(), "John", "Doe", "jdoe@example.com", []string{"mail.send"}, false)
		if reqErr.Err != nil {
			t.Fatalf("UpdateSSOUser() unexpected error: %v", reqErr.Err)
		}
	})

	if _, present := body["has_restricted_subuser_access"]; present {
		t.Error("has_restricted_subuser_access must be omitted when no subuser_access is managed")
	}
	if _, present := body["is_admin"]; present {
		t.Error("is_admin must be omitted when false (do not demote out-of-band admins)")
	}
	if _, present := body["subuser_access"]; present {
		t.Error("subuser_access must be omitted when none is managed")
	}
}

func TestClient_UpdateSSOUser_SendsManagedBlocks(t *testing.T) {
	access := []sendgrid.SubuserAccess{{ID: 111, PermissionType: "restricted", Scopes: []string{"mail.send"}}}
	body := captureSSOBody(t, func(c *sendgrid.Client) {
		_, reqErr := c.UpdateSSOUserWithSubuserAccess(context.Background(), "John", "Doe", "jdoe@example.com", nil, false, access)
		if reqErr.Err != nil {
			t.Fatalf("UpdateSSOUserWithSubuserAccess() unexpected error: %v", reqErr.Err)
		}
	})

	if v, _ := body["has_restricted_subuser_access"].(bool); !v {
		t.Error("has_restricted_subuser_access must be true when blocks are managed")
	}
	if _, present := body["subuser_access"]; !present {
		t.Error("subuser_access must be present when blocks are managed")
	}
}

func TestClient_UpdateSSOUser_SendsIsAdminWhenTrue(t *testing.T) {
	body := captureSSOBody(t, func(c *sendgrid.Client) {
		_, _ = c.UpdateSSOUser(context.Background(), "John", "Doe", "jdoe@example.com", nil, true)
	})
	if v, _ := body["is_admin"].(bool); !v {
		t.Error("is_admin must be sent when true")
	}
}

func TestClient_ReadSubuserAccess_WalksPages(t *testing.T) {
	var seen []string

	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(teammateListBody))
	})
	mux.HandleFunc("/teammates/jdoe/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.RawQuery)
		switch r.URL.Query().Get("after_subuser_id") {
		case "":
			_, _ = w.Write([]byte(`{"has_restricted_subuser_access":true,"subuser_access":[
				{"id":111,"permission_type":"admin"},
				{"id":222,"permission_type":"admin"}
			]}`))
		case "222":
			_, _ = w.Write([]byte(`{"has_restricted_subuser_access":true,"subuser_access":[
				{"id":333,"permission_type":"restricted","scopes":["mail.send"]}
			]}`))
		default:
			_, _ = w.Write([]byte(`{"has_restricted_subuser_access":true,"subuser_access":[]}`))
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := sendgrid.NewClient("api-key", server.URL, "")
	resp, reqErr := client.ReadSubuserAccess(context.Background(), "jdoe@example.com")
	if reqErr.Err != nil {
		t.Fatalf("ReadSubuserAccess() unexpected error: %v", reqErr.Err)
	}

	// A truncated read is what the next update would write back, so every page
	// has to be collected.
	if len(resp.SubuserAccess) != 3 {
		t.Fatalf("collected %d entries across pages, want 3: %+v", len(resp.SubuserAccess), resp.SubuserAccess)
	}
	if resp.SubuserAccess[2].ID != 333 {
		t.Errorf("last entry = %+v, want the id 333 from the second page", resp.SubuserAccess[2])
	}
	if !resp.HasRestrictedSubuserAccess {
		t.Error("expected HasRestrictedSubuserAccess = true from the first page")
	}

	// An explicit limit on every request, and the cursor from the second one on.
	if len(seen) != 3 {
		t.Fatalf("made %d requests (%v), want 3: two pages plus the empty one that ends the walk", len(seen), seen)
	}
	if seen[0] != "limit=500" {
		t.Errorf("first request query = %q, want limit=500 rather than the API's default of 100", seen[0])
	}
	if seen[1] != "limit=500&after_subuser_id=222" {
		t.Errorf("second request query = %q, want the cursor at the highest id of the first page", seen[1])
	}
}

func TestClient_ReadSubuserAccess_StopsWhenCursorIsIgnored(t *testing.T) {
	requests := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(teammateListBody))
	})
	// A server that ignores after_subuser_id serves the same page forever.
	mux.HandleFunc("/teammates/jdoe/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"has_restricted_subuser_access":true,"subuser_access":[
			{"id":111,"permission_type":"admin"}
		]}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := sendgrid.NewClient("api-key", server.URL, "")
	resp, reqErr := client.ReadSubuserAccess(context.Background(), "jdoe@example.com")
	if reqErr.Err != nil {
		t.Fatalf("ReadSubuserAccess() unexpected error: %v", reqErr.Err)
	}

	// Stop instead of looping, and do not turn the repeated page into duplicates.
	if requests != 2 {
		t.Errorf("made %d requests, want 2: the walk must stop once a page fails to advance the cursor", requests)
	}
	if len(resp.SubuserAccess) != 1 {
		t.Errorf("collected %d entries, want 1: a repeated page must not duplicate entries", len(resp.SubuserAccess))
	}
}
