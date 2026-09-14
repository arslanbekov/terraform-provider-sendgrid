package sendgrid

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupSubuserMockServer(handler http.Handler) (*httptest.Server, *Config) {
	server := httptest.NewServer(handler)
	config := &Config{
		APIKey: "test-api-key",
		Host:   server.URL,
	}
	return server, config
}

func TestSubuserRead_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/subusers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":123,"username":"test","email":"test@example.com","disabled":false}]`))
	})

	server, config := setupSubuserMockServer(mux)
	defer server.Close()

	r := resourceSendgridSubuser()
	d := r.TestResourceData()
	d.SetId("test")

	diags := resourceSendgridSubuserRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("user_id") != 123 {
		t.Errorf("Read() user_id = %v, want 123", d.Get("user_id"))
	}
	if d.Get("email") != "test@example.com" {
		t.Errorf("Read() email = %v, want %q", d.Get("email"), "test@example.com")
	}
	if d.Get("disabled") != false {
		t.Errorf("Read() disabled = %v, want false", d.Get("disabled"))
	}
}

func TestSubuserRead_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/subusers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	})

	server, config := setupSubuserMockServer(mux)
	defer server.Close()

	r := resourceSendgridSubuser()
	d := r.TestResourceData()
	d.SetId("missing-user")

	diags := resourceSendgridSubuserRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() should not return error for 404, got: %v", diags)
	}
	if d.Id() != "" {
		t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
	}
}

func TestSubuserRead_EmptyList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/subusers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})

	server, config := setupSubuserMockServer(mux)
	defer server.Close()

	r := resourceSendgridSubuser()
	d := r.TestResourceData()
	d.SetId("empty-user")

	diags := resourceSendgridSubuserRead(context.Background(), d, config)

	if !diags.HasError() {
		t.Fatal("Read() should return error for empty list, got nil")
	}
}

func TestSubuserRead_Region(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/subusers", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("include_region") != "true" {
			t.Errorf("Read() query = %q, want include_region=true", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":123,"username":"test","email":"test@example.com","disabled":false,"region":"eu"}]`))
	})

	server, config := setupSubuserMockServer(mux)
	defer server.Close()

	r := resourceSendgridSubuser()
	d := r.TestResourceData()
	d.SetId("test")

	diags := resourceSendgridSubuserRead(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Read() returned unexpected error: %v", diags)
	}
	if d.Get("region") != "eu" {
		t.Errorf("Read() region = %v, want %q", d.Get("region"), "eu")
	}
}

func TestSubuserCreate_Region(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/subusers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Create() sent invalid JSON: %v", err)
			}
			if body["region"] != "eu" {
				t.Errorf("Create() body region = %v, want %q", body["region"], "eu")
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":123,"username":"test","email":"test@example.com"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":123,"username":"test","email":"test@example.com","region":"eu"}]`))
	})

	server, config := setupSubuserMockServer(mux)
	defer server.Close()

	r := resourceSendgridSubuser()
	d := r.TestResourceData()
	_ = d.Set("username", "test")
	_ = d.Set("email", "test@example.com")
	_ = d.Set("password", "Passw0rd!")
	_ = d.Set("ips", []interface{}{"127.0.0.1"})
	if err := d.Set("region", "eu"); err != nil {
		t.Fatalf("Set(region) failed: %v", err)
	}

	diags := resourceSendgridSubuserCreate(context.Background(), d, config)

	if diags.HasError() {
		t.Fatalf("Create() returned unexpected error: %v", diags)
	}
	if d.Get("region") != "eu" {
		t.Errorf("Create() region = %v, want %q", d.Get("region"), "eu")
	}
}

func TestSubuserResourceSchema(t *testing.T) {
	r := resourceSendgridSubuser()

	tests := []struct {
		field    string
		required bool
		computed bool
	}{
		{"username", true, false},
		{"password", true, false},
		{"email", true, false},
		{"ips", true, false},
		{"user_id", false, true},
		{"disabled", false, true},
		{"region", false, true},
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
			if s.Computed != tt.computed {
				t.Errorf("%s Computed = %v, want %v", tt.field, s.Computed, tt.computed)
			}
		})
	}

	if r.Importer == nil {
		t.Error("resource should have an Importer configured")
	}
}
