package sendgrid_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

// fullSubuserAccessPage renders a page at the walk's own limit, so the walk sees
// a page that could have more behind it.
func fullSubuserAccessPage(firstID int, cursor string) string {
	entries := make([]string, 0, 500)
	for i := 0; i < 500; i++ {
		entries = append(entries, fmt.Sprintf(`{"id":%d,"permission_type":"admin"}`, firstID+i))
	}

	return fmt.Sprintf(`{"has_restricted_subuser_access":true,"subuser_access":[%s]%s}`,
		strings.Join(entries, ","), cursor)
}

func TestClient_ReadSubuserAccess_WalksPagesByPublishedCursor(t *testing.T) {
	var seen []string

	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(teammateListBody))
	})
	mux.HandleFunc("/teammates/jdoe/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.RawQuery)
		switch r.URL.Query().Get("after_subuser_id") {
		case "":
			// The server says there is more, so the walk must continue even though
			// this page is far shorter than the requested limit.
			_, _ = w.Write([]byte(`{"has_restricted_subuser_access":true,"subuser_access":[
				{"id":111,"permission_type":"admin"},
				{"id":222,"permission_type":"admin"}
			],"_metadata":{"next_params":{"after_subuser_id":"222"}}}`))
		case "222":
			// No cursor: nothing is outstanding.
			_, _ = w.Write([]byte(`{"has_restricted_subuser_access":true,"subuser_access":[
				{"id":333,"permission_type":"restricted","scopes":["mail.send"]}
			],"_metadata":{"next_params":{}}}`))
		default:
			t.Errorf("unexpected cursor %q", r.URL.RawQuery)
			_, _ = w.Write([]byte(`{"subuser_access":[]}`))
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := sendgrid.NewClient("api-key", server.URL, "")
	resp, reqErr := client.ReadSubuserAccess(context.Background(), "jdoe@example.com")
	if reqErr.Err != nil {
		t.Fatalf("ReadSubuserAccess() unexpected error: %v", reqErr.Err)
	}

	if len(resp.SubuserAccess) != 3 {
		t.Fatalf("collected %d entries, want 3: %+v", len(resp.SubuserAccess), resp.SubuserAccess)
	}
	// The published cursor ends the walk, so the last page costs no extra request.
	if len(seen) != 2 {
		t.Fatalf("made %d requests (%v), want 2", len(seen), seen)
	}
	if seen[0] != "limit=500" {
		t.Errorf("first request query = %q, want limit=500 rather than the API's default of 100", seen[0])
	}
	if seen[1] != "limit=500&after_subuser_id=222" {
		t.Errorf("second request query = %q, want the cursor the server published", seen[1])
	}
}

func TestClient_ReadSubuserAccess_KeepsFirstPageRestrictedFlag(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(teammateListBody))
	})
	mux.HandleFunc("/teammates/jdoe/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("after_subuser_id") == "" {
			_, _ = w.Write([]byte(`{"has_restricted_subuser_access":true,"subuser_access":[
				{"id":111,"permission_type":"admin"}
			],"_metadata":{"next_params":{"after_subuser_id":111}}}`))
			return
		}
		// A later page disagreeing about the flag must not win.
		_, _ = w.Write([]byte(`{"has_restricted_subuser_access":false,"subuser_access":[
			{"id":222,"permission_type":"admin"}
		],"_metadata":{}}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := sendgrid.NewClient("api-key", server.URL, "")
	resp, reqErr := client.ReadSubuserAccess(context.Background(), "jdoe@example.com")
	if reqErr.Err != nil {
		t.Fatalf("ReadSubuserAccess() unexpected error: %v", reqErr.Err)
	}
	if !resp.HasRestrictedSubuserAccess {
		t.Error("HasRestrictedSubuserAccess = false, want the first page's true")
	}
	if len(resp.SubuserAccess) != 2 {
		t.Errorf("collected %d entries, want 2", len(resp.SubuserAccess))
	}
}

func TestClient_ReadSubuserAccess_SkipsPagingForUnrestrictedTeammate(t *testing.T) {
	requests := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(teammateListBody))
	})
	// For an administrator the endpoint returns every subuser on the account; the
	// list says nothing about this teammate, so there is nothing to page through.
	mux.HandleFunc("/teammates/jdoe/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"has_restricted_subuser_access":false,"subuser_access":[
			{"id":111,"permission_type":"admin"},
			{"id":222,"permission_type":"admin"}
		],"_metadata":{"next_params":{"after_subuser_id":222}}}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := sendgrid.NewClient("api-key", server.URL, "")
	resp, reqErr := client.ReadSubuserAccess(context.Background(), "jdoe@example.com")
	if reqErr.Err != nil {
		t.Fatalf("ReadSubuserAccess() unexpected error: %v", reqErr.Err)
	}
	if requests != 1 {
		t.Errorf("made %d requests, want 1 - an unrestricted teammate must not be paged", requests)
	}
	if resp.HasRestrictedSubuserAccess {
		t.Error("HasRestrictedSubuserAccess = true, want false")
	}
}

func TestClient_ReadSubuserAccess_ErrorsWhenCursorDoesNotAdvance(t *testing.T) {
	requests := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(teammateListBody))
	})
	// A server that ignores after_subuser_id serves the same full page forever.
	mux.HandleFunc("/teammates/jdoe/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		requests++
		_, _ = w.Write([]byte(fullSubuserAccessPage(1, "")))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := sendgrid.NewClient("api-key", server.URL, "")
	resp, reqErr := client.ReadSubuserAccess(context.Background(), "jdoe@example.com")

	// A partial list must never be handed back as success: it becomes state, and
	// the next update writes state over the real access.
	if reqErr.Err == nil {
		t.Fatalf("ReadSubuserAccess() succeeded with %d entries, want an error", len(resp.SubuserAccess))
	}
	if resp != nil {
		t.Errorf("ReadSubuserAccess() returned a list alongside the error: %d entries", len(resp.SubuserAccess))
	}
	if !strings.Contains(reqErr.Err.Error(), "did not advance") {
		t.Errorf("error = %v, want it to name the stalled cursor", reqErr.Err)
	}
	if requests != 2 {
		t.Errorf("made %d requests, want 2 - the walk must stop as soon as it cannot progress", requests)
	}
}

func TestClient_ReadSubuserAccess_ErrorsWhenPageBudgetExhausted(t *testing.T) {
	requests := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(teammateListBody))
	})
	// Always a full page, always a fresh cursor: progress that never ends.
	mux.HandleFunc("/teammates/jdoe/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		requests++
		first := requests * 1000
		cursor := fmt.Sprintf(`,"_metadata":{"next_params":{"after_subuser_id":%d}}`, first+499)
		_, _ = w.Write([]byte(fullSubuserAccessPage(first, cursor)))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := sendgrid.NewClient("api-key", server.URL, "")
	resp, reqErr := client.ReadSubuserAccess(context.Background(), "jdoe@example.com")

	if reqErr.Err == nil {
		t.Fatalf("ReadSubuserAccess() succeeded with %d entries, want an error at the page budget", len(resp.SubuserAccess))
	}
	if !strings.Contains(reqErr.Err.Error(), "did not finish within") {
		t.Errorf("error = %v, want it to name the page budget", reqErr.Err)
	}
	if requests != 40 {
		t.Errorf("made %d requests, want the 40-page budget", requests)
	}
}

func TestClient_ReadSubuserAccess_MidWalkFailureIsNotNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/teammates", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(teammateListBody))
	})
	mux.HandleFunc("/teammates/jdoe/subuser_access", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("after_subuser_id") == "" {
			_, _ = w.Write([]byte(`{"has_restricted_subuser_access":true,"subuser_access":[
				{"id":111,"permission_type":"admin"}
			],"_metadata":{"next_params":{"after_subuser_id":111}}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := sendgrid.NewClient("api-key", server.URL, "")
	_, reqErr := client.ReadSubuserAccess(context.Background(), "jdoe@example.com")

	if reqErr.Err == nil {
		t.Fatal("ReadSubuserAccess() succeeded, want the mid-walk failure surfaced")
	}
	// The resource clears itself from state on a 404, so only the first request
	// may report one - a missing cursor page is not a missing teammate.
	if reqErr.StatusCode == http.StatusNotFound {
		t.Errorf("StatusCode = 404 for a cursor page; that tells the caller the teammate is gone")
	}
}
