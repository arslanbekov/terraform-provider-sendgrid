package sendgrid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SSOIntegration is a Sendgrid SSO configuration set.
type SSOIntegration struct {
	CompletedIntegration bool   `json:"completed_integration,omitempty"` //nolint:tagliatelle
	Enabled              bool   `json:"enabled"`
	Name                 string `json:"name"`
	SignInURL            string `json:"signin_url"`  //nolint:tagliatelle
	SignOutURL           string `json:"signout_url"` //nolint:tagliatelle
	EntityID             string `json:"entity_id"`   //nolint:tagliatelle
	ID                   string `json:"id,omitempty"`
	SingleSignOnURL      string `json:"single_signon_url,omitempty"` //nolint:tagliatelle
	AudienceURL          string `json:"audience_url,omitempty"`      //nolint:tagliatelle
}

// CreateSSOIntegration creates an SSO integration and returns it.
func (c Client) CreateSSOIntegration(
	ctx context.Context,
	name string,
	enabled bool,
	signInURL string,
	signOutURL string,
	entityID string,
) (*SSOIntegration, RequestError) {
	if name == "" {
		return nil, RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("%w: name", ErrSSOIntegrationMissingField),
		}
	}

	respBody, statusCode, err := c.Post(ctx, "POST", "/sso/integrations", SSOIntegration{
		Name:       name,
		Enabled:    enabled,
		SignInURL:  signInURL,
		SignOutURL: signOutURL,
		EntityID:   entityID,
	})
	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed to create SSO integration: %w", err),
		}
	}

	if statusCode >= http.StatusMultipleChoices {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        fmt.Errorf("%w, status: %d, response: %s", ErrFailedCreatingSSOIntegration, statusCode, respBody),
		}
	}

	return parseSSOIntegration(respBody)
}

// ReadSSOIntegration retrieves an SSO integration by ID.
func (c Client) ReadSSOIntegration(ctx context.Context, id string) (*SSOIntegration, RequestError) {
	if id == "" {
		return nil, RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("%w: id", ErrSSOIntegrationMissingField),
		}
	}

	respBody, statusCode, err := c.Get(ctx, "GET", fmt.Sprintf("/sso/integrations/%s", id))
	if err != nil {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	return parseSSOIntegration(respBody)
}

// UpdateSSOIntegration updates an existing SSO integration by ID.
func (c Client) UpdateSSOIntegration(
	ctx context.Context,
	id string,
	name string,
	enabled bool,
	signInURL string,
	signOutURL string,
	entityID string,
) (*SSOIntegration, RequestError) {
	if id == "" {
		return nil, RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("%w: id", ErrSSOIntegrationMissingField),
		}
	}

	if name == "" {
		return nil, RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("%w: name", ErrSSOIntegrationMissingField),
		}
	}

	respBody, statusCode, err := c.Post(ctx, "PATCH", fmt.Sprintf("/sso/integrations/%s", id), SSOIntegration{
		Name:       name,
		Enabled:    enabled,
		SignInURL:  signInURL,
		SignOutURL: signOutURL,
		EntityID:   entityID,
	})
	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	if statusCode >= http.StatusMultipleChoices {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        fmt.Errorf("%w, status: %d, response: %s", ErrFailedUpdatingSSOIntegration, statusCode, respBody),
		}
	}

	return parseSSOIntegration(respBody)
}

// DeleteSSOIntegration deletes an SSO integration by ID.
func (c Client) DeleteSSOIntegration(ctx context.Context, id string) (bool, RequestError) {
	if id == "" {
		return false, RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("%w: id", ErrSSOIntegrationMissingField),
		}
	}

	if _, statusCode, err := c.Get(ctx, "DELETE", fmt.Sprintf("/sso/integrations/%s", id)); statusCode > 299 || err != nil {
		return false, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed to delete SSO integration: %w", err),
		}
	}

	return true, RequestError{
		StatusCode: http.StatusOK,
		Err:        nil,
	}
}

// ListSSOIntegrations returns a list of SSO integrations.
func (c Client) ListSSOIntegrations(ctx context.Context) ([]*SSOIntegration, RequestError) {
	respBody, statusCode, err := c.Get(ctx, "GET", "/sso/integrations")
	if err != nil || statusCode >= 300 {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	return parseSSOIntegrations(respBody)
}

func parseSSOIntegration(respBody string) (*SSOIntegration, RequestError) {
	var body SSOIntegration

	if err := json.Unmarshal([]byte(respBody), &body); err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed to parse SSO integration: %w", err),
		}
	}

	return &body, RequestError{StatusCode: http.StatusOK, Err: nil}
}

func parseSSOIntegrations(respBody string) ([]*SSOIntegration, RequestError) {
	var body []*SSOIntegration

	if err := json.Unmarshal([]byte(respBody), &body); err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed to parse SSO integrations: %w", err),
		}
	}

	return body, RequestError{StatusCode: http.StatusOK, Err: nil}
}
