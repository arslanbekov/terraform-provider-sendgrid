package sendgrid_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
)

func setupParseWebhookServer() (*http.ServeMux, *httptest.Server, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	return mux, server, func() { server.Close() }
}

func TestClient_ReadParseWebhook_Success(t *testing.T) {
	mux, server, teardown := setupParseWebhookServer()
	defer teardown()

	mux.HandleFunc("/user/webhooks/parse/settings/parse.example.com", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hostname":"parse.example.com","url":"https://example.com/parse","spam_check":true,"send_raw":false}`))
	})
	client := sendgrid.NewClient("api-key", server.URL, "")

	webhook, reqErr := client.ReadParseWebhook(context.Background(), "parse.example.com")

	if reqErr.Err != nil {
		t.Fatalf("ReadParseWebhook() unexpected error: %v", reqErr.Err)
	}
	if webhook.Hostname != "parse.example.com" {
		t.Errorf("ReadParseWebhook() Hostname = %q, want %q", webhook.Hostname, "parse.example.com")
	}
	if webhook.URL != "https://example.com/parse" {
		t.Errorf("ReadParseWebhook() URL = %q, want %q", webhook.URL, "https://example.com/parse")
	}
	if !webhook.SpamCheck {
		t.Error("ReadParseWebhook() SpamCheck = false, want true")
	}
}

func TestClient_ReadParseWebhook_NotFound(t *testing.T) {
	mux, server, teardown := setupParseWebhookServer()
	defer teardown()

	mux.HandleFunc("/user/webhooks/parse/settings/missing.example.com", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})
	client := sendgrid.NewClient("api-key", server.URL, "")

	_, reqErr := client.ReadParseWebhook(context.Background(), "missing.example.com")

	if reqErr.Err == nil {
		t.Fatal("ReadParseWebhook() expected error for 404, got nil")
	}
	if reqErr.StatusCode != http.StatusNotFound {
		t.Errorf("ReadParseWebhook() StatusCode = %d, want %d", reqErr.StatusCode, http.StatusNotFound)
	}
}

func TestClient_ReadParseWebhook_EmptyHostname(t *testing.T) {
	client := sendgrid.NewClient("api-key", "http://localhost", "")

	_, reqErr := client.ReadParseWebhook(context.Background(), "")

	if reqErr.Err == nil {
		t.Fatal("ReadParseWebhook() expected error for empty hostname, got nil")
	}
}

func TestClient_UpdateParseWebhook_Success(t *testing.T) {
	mux, server, teardown := setupParseWebhookServer()
	defer teardown()

	var receivedMethod string
	mux.HandleFunc("/user/webhooks/parse/settings/parse.example.com", func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	client := sendgrid.NewClient("api-key", server.URL, "")

	reqErr := client.UpdateParseWebhook(context.Background(), "parse.example.com", true, false, "")

	if reqErr.Err != nil {
		t.Fatalf("UpdateParseWebhook() unexpected error: %v", reqErr.Err)
	}
	if receivedMethod != "PATCH" {
		t.Errorf("UpdateParseWebhook() HTTP method = %q, want PATCH", receivedMethod)
	}
}

func TestClient_UpdateParseWebhook_EmptyHostname(t *testing.T) {
	client := sendgrid.NewClient("api-key", "http://localhost", "")

	reqErr := client.UpdateParseWebhook(context.Background(), "", true, false, "")

	if reqErr.Err == nil {
		t.Fatal("UpdateParseWebhook() expected error for empty hostname, got nil")
	}
}

func TestClient_CreateParseWebhook_Success(t *testing.T) {
	mux, server, teardown := setupParseWebhookServer()
	defer teardown()

	mux.HandleFunc("/user/webhooks/parse/settings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"hostname":"parse.example.com","url":"https://example.com/parse","spam_check":true,"send_raw":false}`))
	})
	client := sendgrid.NewClient("api-key", server.URL, "")

	webhook, reqErr := client.CreateParseWebhook(context.Background(), "parse.example.com", "https://example.com/parse", true, false, "")

	if reqErr.Err != nil {
		t.Fatalf("CreateParseWebhook() unexpected error: %v", reqErr.Err)
	}
	if webhook.Hostname != "parse.example.com" {
		t.Errorf("CreateParseWebhook() Hostname = %q, want %q", webhook.Hostname, "parse.example.com")
	}
}

func TestClient_CreateParseWebhook_EmptyHostname(t *testing.T) {
	client := sendgrid.NewClient("api-key", "http://localhost", "")

	_, reqErr := client.CreateParseWebhook(context.Background(), "", "https://example.com/parse", true, false, "")

	if reqErr.Err == nil {
		t.Fatal("CreateParseWebhook() expected error for empty hostname, got nil")
	}
}

func TestClient_CreateParseWebhook_EmptyURL(t *testing.T) {
	client := sendgrid.NewClient("api-key", "http://localhost", "")

	_, reqErr := client.CreateParseWebhook(context.Background(), "parse.example.com", "", true, false, "")

	if reqErr.Err == nil {
		t.Fatal("CreateParseWebhook() expected error for empty URL, got nil")
	}
}

func TestClient_DeleteParseWebhook_Success(t *testing.T) {
	mux, server, teardown := setupParseWebhookServer()
	defer teardown()

	mux.HandleFunc("/user/webhooks/parse/settings/parse.example.com", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	})
	client := sendgrid.NewClient("api-key", server.URL, "")

	ok, reqErr := client.DeleteParseWebhook(context.Background(), "parse.example.com")

	if reqErr.Err != nil {
		t.Fatalf("DeleteParseWebhook() unexpected error: %v", reqErr.Err)
	}
	if !ok {
		t.Error("DeleteParseWebhook() = false, want true")
	}
}

func TestClient_DeleteParseWebhook_EmptyHostname(t *testing.T) {
	client := sendgrid.NewClient("api-key", "http://localhost", "")

	_, reqErr := client.DeleteParseWebhook(context.Background(), "")

	if reqErr.Err == nil {
		t.Fatal("DeleteParseWebhook() expected error for empty hostname, got nil")
	}
}
