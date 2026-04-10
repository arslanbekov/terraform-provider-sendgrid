package sendgrid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SSOCertificate maps a public certificate to an SSO integration,
// allowing the SSO integration to verify SAML requests from an IdP.
type SSOCertificate struct {
	PublicCertificate string `json:"public_certificate"` //nolint:tagliatelle
	IntegrationID     string `json:"integration_id"`     //nolint:tagliatelle
	ID                int32  `json:"id,omitempty"`
}

// CreateSSOCertificate creates an SSO certificate and returns it.
func (c Client) CreateSSOCertificate(
	ctx context.Context,
	publicCertificate string,
	integrationID string,
) (*SSOCertificate, RequestError) {
	if integrationID == "" {
		return nil, RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("%w: integration_id", ErrSSOCertificateMissingField),
		}
	}

	if publicCertificate == "" {
		return nil, RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("%w: public_certificate", ErrSSOCertificateMissingField),
		}
	}

	respBody, statusCode, err := c.Post(ctx, "POST", "/sso/certificates", SSOCertificate{
		IntegrationID:     integrationID,
		PublicCertificate: publicCertificate,
	})
	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed to create SSO certificate: %w", err),
		}
	}

	if statusCode >= http.StatusMultipleChoices {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        fmt.Errorf("%w, status: %d, response: %s", ErrFailedCreatingSSOCertificate, statusCode, respBody),
		}
	}

	return parseSSOCertificate(respBody)
}

// ReadSSOCertificate retrieves an SSO certificate by ID.
func (c Client) ReadSSOCertificate(ctx context.Context, id string) (*SSOCertificate, RequestError) {
	if id == "" {
		return nil, RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("%w: id", ErrSSOCertificateMissingField),
		}
	}

	respBody, statusCode, err := c.Get(ctx, "GET", fmt.Sprintf("/sso/certificates/%s", id))
	if err != nil {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	return parseSSOCertificate(respBody)
}

// UpdateSSOCertificate updates an existing SSO certificate by ID.
func (c Client) UpdateSSOCertificate(
	ctx context.Context,
	id string,
	publicCertificate string,
	integrationID string,
) (*SSOCertificate, RequestError) {
	if id == "" {
		return nil, RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("%w: id", ErrSSOCertificateMissingField),
		}
	}

	if integrationID == "" {
		return nil, RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("%w: integration_id", ErrSSOCertificateMissingField),
		}
	}

	if publicCertificate == "" {
		return nil, RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("%w: public_certificate", ErrSSOCertificateMissingField),
		}
	}

	respBody, statusCode, err := c.Post(ctx, "PATCH", fmt.Sprintf("/sso/certificates/%s", id), SSOCertificate{
		IntegrationID:     integrationID,
		PublicCertificate: publicCertificate,
	})
	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed to update SSO certificate: %w", err),
		}
	}

	if statusCode >= http.StatusMultipleChoices {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        fmt.Errorf("%w, status: %d, response: %s", ErrFailedUpdatingSSOCertificate, statusCode, respBody),
		}
	}

	return parseSSOCertificate(respBody)
}

// DeleteSSOCertificate deletes an SSO certificate by ID.
func (c Client) DeleteSSOCertificate(ctx context.Context, id string) (bool, RequestError) {
	if id == "" {
		return false, RequestError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("%w: id", ErrSSOCertificateMissingField),
		}
	}

	if _, statusCode, err := c.Get(ctx, "DELETE", fmt.Sprintf("/sso/certificates/%s", id)); statusCode > 299 || err != nil {
		return false, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed deleting SSO integration: %w", err),
		}
	}

	return true, RequestError{
		StatusCode: http.StatusOK,
		Err:        nil,
	}
}

// ListSSOCertificates retrieves all existing SSO certificates.
func (c Client) ListSSOCertificates(ctx context.Context) ([]*SSOCertificate, RequestError) {
	respBody, statusCode, err := c.Get(ctx, "GET", "/sso/certificates")
	if err != nil || statusCode >= 300 {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	return parseSSOCertificates(respBody)
}

func parseSSOCertificate(respBody string) (*SSOCertificate, RequestError) {
	var body SSOCertificate

	if err := json.Unmarshal([]byte(respBody), &body); err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed to parse SSO certificate: %w", err),
		}
	}

	return &body, RequestError{StatusCode: http.StatusOK, Err: nil}
}

func parseSSOCertificates(respBody string) ([]*SSOCertificate, RequestError) {
	var body []*SSOCertificate

	if err := json.Unmarshal([]byte(respBody), &body); err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed to parse SSO certificates: %w", err),
		}
	}

	return body, RequestError{StatusCode: http.StatusOK, Err: nil}
}
