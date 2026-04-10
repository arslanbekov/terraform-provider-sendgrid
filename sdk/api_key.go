package sendgrid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// APIKey is a Sendgrid API key.
type APIKey struct {
	ID     string   `json:"api_key_id,omitempty"` //nolint:tagliatelle
	APIKey string   `json:"api_key,omitempty"`    //nolint:tagliatelle
	Name   string   `json:"name,omitempty"`
	Scopes []string `json:"scopes,omitempty"`
}

func parseAPIKey(respBody string) (*APIKey, RequestError) {
	var body APIKey
	if err := json.Unmarshal([]byte(respBody), &body); err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed parsing API key: %w", err),
		}
	}

	return &body, RequestError{StatusCode: http.StatusOK, Err: nil}
}

func parseAPIKeys(respBody string) ([]APIKey, RequestError) {
	var body []APIKey
	if err := json.Unmarshal([]byte(respBody), &body); err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed parsing API key: %w", err),
		}
	}

	return body, RequestError{StatusCode: http.StatusOK, Err: nil}
}

// CreateAPIKey creates an APIKey and returns it.
func (c *Client) CreateAPIKey(ctx context.Context, name string, scopes []string) (*APIKey, RequestError) {
	if name == "" {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrNameRequired,
		}
	}

	respBody, statusCode, err := c.Post(ctx, "POST", "/api_keys", APIKey{
		Name:   name,
		Scopes: scopes,
	})
	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed creating API key: %w", err),
		}
	}

	if statusCode >= http.StatusMultipleChoices {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        fmt.Errorf("%w, status: %d, response: %s", ErrFailedCreatingAPIKey, statusCode, respBody),
		}
	}

	return parseAPIKey(respBody)
}

// ReadAPIKey retreives an APIKey and returns it.
func (c *Client) ReadAPIKey(ctx context.Context, id string) (*APIKey, RequestError) {
	if id == "" {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrAPIKeyIDRequired,
		}
	}

	respBody, statusCode, err := c.Get(ctx, "GET", "/api_keys/"+id)
	if err != nil {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	return parseAPIKey(respBody)
}

func (c *Client) ReadAPIKeys(ctx context.Context) ([]APIKey, RequestError) {
	respBody, statusCode, err := c.Get(ctx, "GET", "/api_keys")
	if err != nil {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	return parseAPIKeys(respBody)
}

// UpdateAPIKey edits an APIKey and returns it.
func (c *Client) UpdateAPIKey(ctx context.Context, id, name string, scopes []string) (*APIKey, RequestError) {
	if id == "" {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrAPIKeyIDRequired,
		}
	}

	t := APIKey{}
	if name != "" {
		t.Name = name
	}

	if len(scopes) > 0 {
		t.Scopes = scopes
	}

	respBody, _, err := c.Post(ctx, "PUT", "/api_keys/"+id, t)
	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	return parseAPIKey(respBody)
}

// DeleteAPIKey deletes an APIKey.
func (c *Client) DeleteAPIKey(ctx context.Context, id string) (bool, RequestError) {
	if id == "" {
		return false, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrAPIKeyIDRequired,
		}
	}

	responseBody, statusCode, err := c.Get(ctx, "DELETE", "/api_keys/"+id)
	if err != nil {
		return false, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	if statusCode >= http.StatusMultipleChoices && statusCode != http.StatusNotFound { // ignore not found
		return false, RequestError{
			StatusCode: statusCode,
			Err:        fmt.Errorf("%w, status: %d, response: %s", ErrFailedDeletingAPIKey, statusCode, responseBody),
		}
	}

	return true, RequestError{StatusCode: http.StatusOK, Err: nil}
}
