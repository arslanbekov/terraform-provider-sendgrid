package sendgrid

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func setupMockServer(handler http.Handler) (*httptest.Server, *Config) {
	server := httptest.NewServer(handler)
	config := &Config{
		APIKey: "test-api-key",
		Host:   server.URL + "/",
	}

	return server, config
}

func TestDomainAuthenticationValidationRead_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/123", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":123,"domain":"example.com","valid":true}`))
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	r := resourceSendgridDomainAuthenticationValidation()
	d := r.TestResourceData()
	d.SetId("123")
	_ = d.Set("domain_authentication_id", "123")

	diags := resourceSendgridDomainAuthenticationValidationRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("valid") != true {
		t.Errorf("Read() valid = %v, want true", d.Get("valid"))
	}
	if d.Get("domain_authentication_id") != "123" {
		t.Errorf("Read() domain_authentication_id = %v, want '123'", d.Get("domain_authentication_id"))
	}
}

func TestDomainAuthenticationValidationRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/999", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	r := resourceSendgridDomainAuthenticationValidation()
	d := r.TestResourceData()
	d.SetId("999")
	_ = d.Set("domain_authentication_id", "999")

	diags := resourceSendgridDomainAuthenticationValidationRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestDomainAuthenticationValidationRead_WithOnBehalfOf(t *testing.T) {
	var receivedOnBehalfOf string
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/123", func(w http.ResponseWriter, r *http.Request) {
		receivedOnBehalfOf = r.Header.Get("on-behalf-of")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":123,"domain":"example.com","valid":false}`))
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	r := resourceSendgridDomainAuthenticationValidation()
	d := r.TestResourceData()
	d.SetId("123")
	_ = d.Set("domain_authentication_id", "123")
	_ = d.Set("sub_user_on_behalf_of", "test-subuser")

	diags := resourceSendgridDomainAuthenticationValidationRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if receivedOnBehalfOf != "test-subuser" {
		t.Errorf("Read() on-behalf-of header = %q, want %q", receivedOnBehalfOf, "test-subuser")
	}
	if d.Get("valid") != false {
		t.Errorf("Read() valid = %v, want false", d.Get("valid"))
	}
}

func TestDomainAuthenticationValidationCreate_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/456/validate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("validate expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":456,"valid":true}`))
	})
	mux.HandleFunc("/whitelabel/domains/456", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":456,"domain":"example.com","valid":true}`))
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	r := resourceSendgridDomainAuthenticationValidation()
	d := r.TestResourceData()
	_ = d.Set("domain_authentication_id", "456")

	diags := resourceSendgridDomainAuthenticationValidationCreate(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Create() returned unexpected error: %v", diags)
	}
	if d.Id() != "456" {
		t.Errorf("Create() ID = %q, want %q", d.Id(), "456")
	}
	if d.Get("valid") != true {
		t.Errorf("Create() valid = %v, want true", d.Get("valid"))
	}
}

func TestDomainAuthenticationValidationCreate_ValidationFailed(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains/456/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":456,"valid":false,"validation_results":{"mail_cname":{"valid":false}}}`))
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	r := resourceSendgridDomainAuthenticationValidation()
	d := r.TestResourceData()
	_ = d.Set("domain_authentication_id", "456")

	diags := resourceSendgridDomainAuthenticationValidationCreate(context.Background(), d, config)

	if !diags.HasError() {
		t.Fatal("Create() expected error when validation fails, got nil")
	}
}

func TestDomainAuthenticationValidationDelete(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{}, map[string]interface{}{})

	diags := resourceSendgridDomainAuthenticationValidationDelete(context.Background(), d, nil)

	if diags != nil {
		t.Errorf("Delete() = %v, want nil", diags)
	}
}

func TestDomainAuthenticationValidationResourceSchema(t *testing.T) {
	r := resourceSendgridDomainAuthenticationValidation()

	tests := []struct {
		field    string
		required bool
		forceNew bool
	}{
		{"domain_authentication_id", true, true},
		{"sub_user_on_behalf_of", false, true},
		{"valid", false, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("schema_%s", tt.field), func(t *testing.T) {
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

	if r.Schema["valid"].Computed != true {
		t.Error("valid field should be Computed")
	}
}
