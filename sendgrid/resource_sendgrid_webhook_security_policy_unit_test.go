package sendgrid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupWebhookSecurityPolicyMockServer(handler http.Handler) (*httptest.Server, *Config) {
	server := httptest.NewServer(handler)
	config := &Config{
		APIKey: "test-api-key",
		Host:   server.URL,
	}
	return server, config
}

func TestWebhookSecurityPolicyRead_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/webhooks/security/policies/pol-123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"policy":{
				"id":"pol-123",
				"name":"My Policy",
				"oauth":{
					"client_id":"client-abc",
					"token_url":"https://example.com/token",
					"scopes":["webhook.read"]
				},
				"signature":{
					"public_key":"pk-xyz"
				}
			}
		}`))
	})

	server, config := setupWebhookSecurityPolicyMockServer(mux)
	defer server.Close()

	r := resourceSendgridWebhookSecurityPolicy()
	d := r.TestResourceData()
	d.SetId("pol-123")

	diags := resourceSendgridWebhookSecurityPolicyRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("name") != "My Policy" {
		t.Errorf("Read() name = %v, want %q", d.Get("name"), "My Policy")
	}

	// Check oauth block
	oauthList := d.Get("oauth").([]interface{})
	if len(oauthList) != 1 {
		t.Fatalf("Read() oauth length = %d, want 1", len(oauthList))
	}
	oauth := oauthList[0].(map[string]interface{})
	if oauth["client_id"] != "client-abc" {
		t.Errorf("Read() oauth.client_id = %v, want %q", oauth["client_id"], "client-abc")
	}
	if oauth["token_url"] != "https://example.com/token" {
		t.Errorf("Read() oauth.token_url = %v, want %q", oauth["token_url"], "https://example.com/token")
	}

	// Check signature block
	sigList := d.Get("signature").([]interface{})
	if len(sigList) != 1 {
		t.Fatalf("Read() signature length = %d, want 1", len(sigList))
	}
	sig := sigList[0].(map[string]interface{})
	if sig["public_key"] != "pk-xyz" {
		t.Errorf("Read() signature.public_key = %v, want %q", sig["public_key"], "pk-xyz")
	}
	if sig["enabled"] != true {
		t.Errorf("Read() signature.enabled = %v, want true", sig["enabled"])
	}
}

func TestWebhookSecurityPolicyRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/webhooks/security/policies/pol-missing", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	server, config := setupWebhookSecurityPolicyMockServer(mux)
	defer server.Close()

	r := resourceSendgridWebhookSecurityPolicy()
	d := r.TestResourceData()
	d.SetId("pol-missing")

	diags := resourceSendgridWebhookSecurityPolicyRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestWebhookSecurityPolicyResourceSchema(t *testing.T) {
	r := resourceSendgridWebhookSecurityPolicy()

	// Check name field
	name, ok := r.Schema["name"]
	if !ok {
		t.Fatal("schema missing field \"name\"")
	}
	if !name.Required {
		t.Error("name should be Required")
	}

	// Check oauth block exists
	oauth, ok := r.Schema["oauth"]
	if !ok {
		t.Fatal("schema missing field \"oauth\"")
	}
	if !oauth.Optional {
		t.Error("oauth should be Optional")
	}
	if oauth.MaxItems != 1 {
		t.Errorf("oauth MaxItems = %d, want 1", oauth.MaxItems)
	}
	if !oauth.ForceNew {
		t.Error("oauth should be ForceNew")
	}

	// Check signature block exists
	sig, ok := r.Schema["signature"]
	if !ok {
		t.Fatal("schema missing field \"signature\"")
	}
	if !sig.Optional {
		t.Error("signature should be Optional")
	}
	if sig.MaxItems != 1 {
		t.Errorf("signature MaxItems = %d, want 1", sig.MaxItems)
	}
	if !sig.ForceNew {
		t.Error("signature should be ForceNew")
	}

	if r.Importer == nil {
		t.Error("resource should have an Importer configured")
	}
}
