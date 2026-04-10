package sendgrid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupParseWebhookMockServer(handler http.Handler) (*httptest.Server, *Config) {
	server := httptest.NewServer(handler)
	config := &Config{
		APIKey: "test-api-key",
		Host:   server.URL,
	}
	return server, config
}

func TestParseWebhookRead_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/webhooks/parse/settings/parse.example.com", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hostname":"parse.example.com","url":"https://example.com/parse","spam_check":true,"send_raw":false,"security_policy":"policy-123"}`))
	})

	server, config := setupParseWebhookMockServer(mux)
	defer server.Close()

	r := resourceSendgridParseWebhook()
	d := r.TestResourceData()
	d.SetId("parse.example.com")

	diags := resourceSendgridParseWebhookRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("hostname") != "parse.example.com" {
		t.Errorf("Read() hostname = %v, want %q", d.Get("hostname"), "parse.example.com")
	}
	if d.Get("url") != "https://example.com/parse" {
		t.Errorf("Read() url = %v, want %q", d.Get("url"), "https://example.com/parse")
	}
	if d.Get("spam_check") != true {
		t.Errorf("Read() spam_check = %v, want true", d.Get("spam_check"))
	}
	if d.Get("send_raw") != false {
		t.Errorf("Read() send_raw = %v, want false", d.Get("send_raw"))
	}
	if d.Get("webhook_security_policy_id") != "policy-123" {
		t.Errorf("Read() webhook_security_policy_id = %v, want %q", d.Get("webhook_security_policy_id"), "policy-123")
	}
}

func TestParseWebhookRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/webhooks/parse/settings/missing.example.com", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	server, config := setupParseWebhookMockServer(mux)
	defer server.Close()

	r := resourceSendgridParseWebhook()
	d := r.TestResourceData()
	d.SetId("missing.example.com")

	diags := resourceSendgridParseWebhookRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestParseWebhookRead_ServerError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/webhooks/parse/settings/error.example.com", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"message":"server error"}]}`))
	})

	server, config := setupParseWebhookMockServer(mux)
	defer server.Close()

	r := resourceSendgridParseWebhook()
	d := r.TestResourceData()
	d.SetId("error.example.com")

	diags := resourceSendgridParseWebhookRead(context.Background(), d, config)

	if !diags.HasError() {
		t.Fatal("Read() expected error for 500, got nil")
	}
	if d.Id() == "" {
		t.Error("Read() should NOT clear ID on non-404 error")
	}
}

func TestParseWebhookDelete(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/webhooks/parse/settings/parse.example.com", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	})

	server, config := setupParseWebhookMockServer(mux)
	defer server.Close()

	r := resourceSendgridParseWebhook()
	d := r.TestResourceData()
	d.SetId("parse.example.com")

	diags := resourceSendgridParseWebhookDelete(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Delete() returned unexpected error: %v", diags)
	}
}

func TestParseWebhookResourceSchema(t *testing.T) {
	r := resourceSendgridParseWebhook()

	tests := []struct {
		field    string
		required bool
		forceNew bool
	}{
		{"hostname", true, true},
		{"url", true, true},
		{"spam_check", false, false},
		{"send_raw", false, false},
		{"webhook_security_policy_id", false, false},
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
			if s.ForceNew != tt.forceNew {
				t.Errorf("%s ForceNew = %v, want %v", tt.field, s.ForceNew, tt.forceNew)
			}
		})
	}

	if r.Importer == nil {
		t.Error("resource should have an Importer configured")
	}
}

func TestParseWebhookUpdate_Success(t *testing.T) {
	mux := http.NewServeMux()
	var patchReceived bool
	mux.HandleFunc("/user/webhooks/parse/settings/parse.example.com", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PATCH" {
			patchReceived = true
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hostname":"parse.example.com","url":"https://example.com/parse","spam_check":false,"send_raw":true}`))
	})

	server, config := setupParseWebhookMockServer(mux)
	defer server.Close()

	r := resourceSendgridParseWebhook()
	d := r.TestResourceData()
	d.SetId("parse.example.com")
	_ = d.Set("spam_check", false)
	_ = d.Set("send_raw", true)
	_ = d.Set("webhook_security_policy_id", "")

	// Need a mock ResourceData with proper schema for RetryOnRateLimit
	// Use a raw schema.TestResourceDataRaw for the timeout
	diags := resourceSendgridParseWebhookUpdate(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Update() returned unexpected error: %v", diags)
	}
	if !patchReceived {
		t.Error("Update() should send PATCH request, but PATCH was not received")
	}
}

func TestParseWebhookCustomizeDiff(t *testing.T) {
	r := resourceSendgridParseWebhook()

	if r.CustomizeDiff == nil {
		t.Error("resource should have a CustomizeDiff configured")
	}
}

func TestParseWebhookDelete_WithSecurityPolicy(t *testing.T) {
	mux := http.NewServeMux()
	var patchReceived bool
	var deleteReceived bool
	mux.HandleFunc("/user/webhooks/parse/settings/parse.example.com", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PATCH":
			patchReceived = true
		case "DELETE":
			deleteReceived = true
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	})

	server, config := setupParseWebhookMockServer(mux)
	defer server.Close()

	r := resourceSendgridParseWebhook()
	d := r.TestResourceData()
	d.SetId("parse.example.com")
	_ = d.Set("webhook_security_policy_id", "policy-123")
	_ = d.Set("spam_check", true)
	_ = d.Set("send_raw", false)

	diags := resourceSendgridParseWebhookDelete(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Delete() returned unexpected error: %v", diags)
	}
	if !patchReceived {
		t.Error("Delete() with security policy should PATCH to clear policy first")
	}
	if !deleteReceived {
		t.Error("Delete() should send DELETE request")
	}
}
