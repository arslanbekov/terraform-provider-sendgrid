package sendgrid

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name       string
		apiKey     string
		host       string
		onBehalfOf string
		wantNil    bool
		wantHost   string
	}{
		{
			name:       "valid client with default host",
			apiKey:     "test-api-key",
			host:       "",
			onBehalfOf: "",
			wantNil:    false,
			wantHost:   "https://api.sendgrid.com/v3/",
		},
		{
			name:       "valid client with custom host",
			apiKey:     "test-api-key",
			host:       "https://custom.sendgrid.com/v3/",
			onBehalfOf: "",
			wantNil:    false,
			wantHost:   "https://custom.sendgrid.com/v3/",
		},
		{
			name:       "with subuser",
			apiKey:     "test-api-key",
			host:       "",
			onBehalfOf: "subuser@example.com",
			wantNil:    false,
			wantHost:   "https://api.sendgrid.com/v3/",
		},
		{
			name:       "empty api key",
			apiKey:     "",
			host:       "",
			onBehalfOf: "",
			wantNil:    false, // Still creates client, validation happens at API call time
			wantHost:   "https://api.sendgrid.com/v3/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.apiKey, tt.host, tt.onBehalfOf)
			if (client == nil) != tt.wantNil {
				t.Errorf("NewClient() = %v, want nil = %v", client, tt.wantNil)
			}
			if client != nil {
				if client.apiKey != tt.apiKey {
					t.Errorf("NewClient().apiKey = %v, want %v", client.apiKey, tt.apiKey)
				}
				if client.OnBehalfOf != tt.onBehalfOf {
					t.Errorf("NewClient().OnBehalfOf = %v, want %v", client.OnBehalfOf, tt.onBehalfOf)
				}
				if client.host != tt.wantHost {
					t.Errorf("NewClient().host = %v, want %v", client.host, tt.wantHost)
				}
			}
		})
	}
}

func TestBodyToJSON(t *testing.T) {
	tests := []struct {
		name    string
		body    interface{}
		wantErr bool
		errType error
	}{
		{
			name:    "valid struct",
			body:    map[string]string{"key": "value"},
			wantErr: false,
		},
		{
			name:    "valid string",
			body:    "test string",
			wantErr: false,
		},
		{
			name:    "nil body",
			body:    nil,
			wantErr: true,
			errType: ErrBodyNotNil,
		},
		{
			name:    "complex struct",
			body:    struct{ Name string }{"test"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := bodyToJSON(tt.body)
			if tt.wantErr {
				if err == nil {
					t.Errorf("bodyToJSON() error = nil, want error")
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("bodyToJSON() error = %v, want %v", err, tt.errType)
				}
			} else {
				if err != nil {
					t.Errorf("bodyToJSON() error = %v, want nil", err)
				}
				if result == nil {
					t.Error("bodyToJSON() result = nil, want non-nil")
				}
			}
		})
	}
}

// closedServerHost returns the address of a server that is no longer listening,
// so the next request fails in the transport with no response at all.
func closedServerHost(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	host := server.URL
	server.Close()

	return host
}

func TestClientGetTransportFailureReturnsError(t *testing.T) {
	client := NewClient("api-key", closedServerHost(t), "")

	// A transport failure yields no response; reading a status off it panicked
	// the provider, which takes the whole terraform run with it.
	body, status, err := client.Get(context.Background(), "GET", "/teammates")
	if err == nil {
		t.Fatalf("Get() succeeded against a closed server: body=%q status=%d", body, status)
	}
	if status != 0 {
		t.Errorf("status = %d, want 0 - there is no HTTP status to report", status)
	}
}

func TestClientPostTransportFailureReturnsError(t *testing.T) {
	client := NewClient("api-key", closedServerHost(t), "")

	body, status, err := client.Post(context.Background(), "PATCH", "/teammates/jdoe", nil)
	if err == nil {
		t.Fatalf("Post() succeeded against a closed server: body=%q status=%d", body, status)
	}
	if status != 0 {
		t.Errorf("status = %d, want 0 - there is no HTTP status to report", status)
	}
}

func TestClientGetHonorsContextCancellation(t *testing.T) {
	served := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served++
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("api-key", server.URL, "")

	// The happy path still works.
	if _, _, err := client.Get(context.Background(), "GET", "/teammates"); err != nil {
		t.Fatalf("Get() with a live context failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := client.Get(ctx, "GET", "/teammates")
	if err == nil {
		t.Fatal("Get() with a cancelled context succeeded; the context never reached the request")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want it to wrap context.Canceled", err)
	}
	if served != 1 {
		t.Errorf("server served %d requests, want 1 - the cancelled call must not be sent", served)
	}
}

func TestClientPostHonorsContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient("api-key", server.URL, "")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := client.Post(ctx, "PATCH", "/teammates/jdoe", map[string]string{"first_name": "John"})
	if err == nil {
		t.Fatal("Post() with a cancelled context succeeded; the context never reached the request")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want it to wrap context.Canceled", err)
	}
}
