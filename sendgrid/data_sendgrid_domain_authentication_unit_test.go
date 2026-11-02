package sendgrid

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

const parentDomainsJSON = `[
	{"id":1,"domain":"app.example.com","subdomain":"em1234","valid":true,"automatic_security":true,
	 "dns":{"mail_cname":{"valid":true,"type":"cname","host":"em1234.app.example.com","data":"u7.wl.sendgrid.net"}}},
	{"id":2,"domain":"app.example.com","subdomain":"em5678","valid":false},
	{"id":3,"domain":"other-app.example.com","subdomain":"em","valid":true}
]`

func setupDomainListServer(t *testing.T) (*Config, func()) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("exclude_subusers") != "true" {
			t.Errorf("expected exclude_subusers=true, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(parentDomainsJSON))
	})
	mux.HandleFunc("/whitelabel/domains/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1,"domain":"app.example.com","subdomain":"em1234","valid":true,"automatic_security":true,
		 "dns":{"mail_cname":{"valid":true,"type":"cname","host":"em1234.app.example.com","data":"u7.wl.sendgrid.net"}}}`))
	})
	mux.HandleFunc("/whitelabel/domains/3", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":3,"domain":"other-app.example.com","subdomain":"em","valid":true}`))
	})

	server, config := setupMockServer(mux)

	return config, server.Close
}

func TestDataSourceDomainAuthentication_ByDomainAndSubdomain(t *testing.T) {
	config, teardown := setupDomainListServer(t)
	defer teardown()

	d := dataSendgridDomainAuthentication().TestResourceData()
	_ = d.Set("domain", "app.example.com")
	_ = d.Set("subdomain", "em1234")

	diags := dataSendgridDomainAuthenticationRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Id() != "1" {
		t.Errorf("Read() ID = %q, want %q", d.Id(), "1")
	}
	if d.Get("valid") != true || d.Get("dns.#") != 1 {
		t.Errorf("Read() valid=%v dns.#=%v", d.Get("valid"), d.Get("dns.#"))
	}
}

func TestDataSourceDomainAuthentication_ByDomainOnly(t *testing.T) {
	config, teardown := setupDomainListServer(t)
	defer teardown()

	d := dataSendgridDomainAuthentication().TestResourceData()
	_ = d.Set("domain", "other-app.example.com")

	diags := dataSendgridDomainAuthenticationRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Id() != "3" || d.Get("subdomain") != "em" {
		t.Errorf("Read() ID=%q subdomain=%v", d.Id(), d.Get("subdomain"))
	}
}

func TestDataSourceDomainAuthentication_Ambiguous(t *testing.T) {
	config, teardown := setupDomainListServer(t)
	defer teardown()

	d := dataSendgridDomainAuthentication().TestResourceData()
	_ = d.Set("domain", "app.example.com")

	diags := dataSendgridDomainAuthenticationRead(context.Background(), d, config)

	if !diags.HasError() || !strings.Contains(diags[0].Summary, "set 'subdomain'") {
		t.Fatalf("Read() expected ambiguity error, got: %v", diags)
	}
}

func TestDataSourceDomainAuthentication_NotFound(t *testing.T) {
	config, teardown := setupDomainListServer(t)
	defer teardown()

	d := dataSendgridDomainAuthentication().TestResourceData()
	// substring of a real domain: the API search would return it, exact match must not
	_ = d.Set("domain", "example.com")

	diags := dataSendgridDomainAuthenticationRead(context.Background(), d, config)

	if !diags.HasError() || !strings.Contains(diags[0].Summary, "unable to find an authenticated domain") {
		t.Fatalf("Read() expected not-found error, got: %v", diags)
	}
}

func TestDataSourceDomainAuthentication_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/whitelabel/domains", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	server, config := setupMockServer(mux)
	defer server.Close()

	d := dataSendgridDomainAuthentication().TestResourceData()
	_ = d.Set("domain", "app.example.com")

	diags := dataSendgridDomainAuthenticationRead(context.Background(), d, config)

	if !diags.HasError() {
		t.Fatal("Read() expected error for 500, got nil")
	}
	if d.Id() != "" {
		t.Errorf("Read() should not set ID on API error, got %q", d.Id())
	}
}

func TestDataSourceDomainAuthentication_Schema(t *testing.T) {
	s := dataSendgridDomainAuthentication().Schema

	if !s["domain"].Required {
		t.Error("domain should be Required")
	}
	if !s["subdomain"].Optional || !s["subdomain"].Computed {
		t.Error("subdomain should be Optional and Computed")
	}
	for _, field := range []string{"valid", "dns", "ips", "username", "is_default"} {
		if !s[field].Computed || s[field].Optional || s[field].Required {
			t.Errorf("%s should be Computed only", field)
		}
	}
}
