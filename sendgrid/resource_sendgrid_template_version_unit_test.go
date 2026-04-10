package sendgrid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTemplateVersionMockServer(handler http.Handler) (*httptest.Server, *Config) {
	server := httptest.NewServer(handler)
	config := &Config{
		APIKey: "test-api-key",
		Host:   server.URL,
	}
	return server, config
}

func TestTemplateVersionRead_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/templates/tmpl-123/versions/ver-456", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id":"ver-456",
			"template_id":"tmpl-123",
			"name":"My Version",
			"subject":"Hello",
			"html_content":"<p>Hi</p>",
			"plain_content":"Hi",
			"generate_plain_content":true,
			"active":1,
			"editor":"code",
			"thumbnail_url":"https://example.com/thumb.png",
			"updated_at":"2024-01-01",
			"test_data":"{\"name\":\"test\"}"
		}`))
	})

	server, config := setupTemplateVersionMockServer(mux)
	defer server.Close()

	r := resourceSendgridTemplateVersion()
	d := r.TestResourceData()
	d.SetId("ver-456")
	_ = d.Set("template_id", "tmpl-123")

	diags := resourceSendgridTemplateVersionRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("name") != "My Version" {
		t.Errorf("Read() name = %v, want %q", d.Get("name"), "My Version")
	}
	if d.Get("subject") != "Hello" {
		t.Errorf("Read() subject = %v, want %q", d.Get("subject"), "Hello")
	}
	if d.Get("html_content") != "<p>Hi</p>" {
		t.Errorf("Read() html_content = %v, want %q", d.Get("html_content"), "<p>Hi</p>")
	}
	if d.Get("plain_content") != "Hi" {
		t.Errorf("Read() plain_content = %v, want %q", d.Get("plain_content"), "Hi")
	}
	if d.Get("generate_plain_content") != true {
		t.Errorf("Read() generate_plain_content = %v, want true", d.Get("generate_plain_content"))
	}
	if d.Get("active") != 1 {
		t.Errorf("Read() active = %v, want 1", d.Get("active"))
	}
	if d.Get("editor") != "code" {
		t.Errorf("Read() editor = %v, want %q", d.Get("editor"), "code")
	}
	if d.Get("thumbnail_url") != "https://example.com/thumb.png" {
		t.Errorf("Read() thumbnail_url = %v, want %q", d.Get("thumbnail_url"), "https://example.com/thumb.png")
	}
	if d.Get("updated_at") != "2024-01-01" {
		t.Errorf("Read() updated_at = %v, want %q", d.Get("updated_at"), "2024-01-01")
	}
}

func TestTemplateVersionRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/templates/tmpl-123/versions/ver-missing", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	server, config := setupTemplateVersionMockServer(mux)
	defer server.Close()

	r := resourceSendgridTemplateVersion()
	d := r.TestResourceData()
	d.SetId("ver-missing")
	_ = d.Set("template_id", "tmpl-123")

	diags := resourceSendgridTemplateVersionRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestTemplateVersionResourceSchema(t *testing.T) {
	r := resourceSendgridTemplateVersion()

	tests := []struct {
		field    string
		required bool
	}{
		{"template_id", true},
		{"name", true},
		{"subject", true},
		{"html_content", false},
		{"plain_content", false},
		{"generate_plain_content", false},
		{"active", false},
		{"editor", false},
		{"test_data", false},
		{"updated_at", false},
		{"thumbnail_url", false},
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
		})
	}

	if r.Importer == nil {
		t.Error("resource should have an Importer configured")
	}
}
