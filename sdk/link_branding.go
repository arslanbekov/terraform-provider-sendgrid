package sendgrid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type LinkBrandingDNS struct {
	DomainCNAME LinkBrandingDNSValue `json:"domain_cname,omitempty"` //nolint:tagliatelle
	OwnerCNAME  LinkBrandingDNSValue `json:"owner_cname,omitempty"`  //nolint:tagliatelle
}

type LinkBrandingDNSValue struct {
	Valid bool   `json:"valid,omitempty"`
	Type  string `json:"type,omitempty"`
	Host  string `json:"host,omitempty"`
	Data  string `json:"data,omitempty"`
}

// LinkBranding is a Sendgrid domain authentication.
type LinkBranding struct {
	ID        int32           `json:"id,omitempty"`
	UserID    int32           `json:"user_id,omitempty"` //nolint:tagliatelle
	Domain    string          `json:"domain,omitempty"`
	Subdomain string          `json:"subdomain,omitempty"`
	Username  string          `json:"username,omitempty"`
	IsDefault bool            `json:"default"`
	Legacy    bool            `json:"legacy,omitempty"`
	Valid     bool            `json:"valid,omitempty"`
	DNS       LinkBrandingDNS `json:"dns,omitempty"`
}

func parseLinkBranding(respBody string) (*LinkBranding, RequestError) {
	var body LinkBranding
	if err := json.Unmarshal([]byte(respBody), &body); err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed parsing API key: %w", err),
		}
	}

	return &body, RequestError{StatusCode: http.StatusOK, Err: nil}
}

// CreateLinkBranding creates an LinkBranding and returns it.
func (c *Client) CreateLinkBranding(ctx context.Context, domain string, subdomain string, isDefault bool) (*LinkBranding, RequestError) {
	if domain == "" {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrNameRequired,
		}
	}

	respBody, statusCode, err := c.Post(ctx, "POST", "/whitelabel/links", LinkBranding{
		Domain:    domain,
		Subdomain: subdomain,
		IsDefault: isDefault,
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
			Err:        fmt.Errorf("%w, status: %d, response: %s", ErrFailedCreatingLinkBranding, statusCode, respBody),
		}
	}

	return parseLinkBranding(respBody)
}

// ReadLinkBranding retrieves an LinkBranding and returns it.
func (c *Client) ReadLinkBranding(ctx context.Context, id string) (*LinkBranding, RequestError) {
	if id == "" {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrLinkBrandingIDRequired,
		}
	}

	respBody, statusCode, err := c.Get(ctx, "GET", "/whitelabel/links/"+id)
	if err != nil {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	return parseLinkBranding(respBody)
}

// UpdateLinkBranding edits an LinkBranding and returns it.
func (c *Client) UpdateLinkBranding(ctx context.Context, id string, isDefault bool) (*LinkBranding, RequestError) {
	if id == "" {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrLinkBrandingIDRequired,
		}
	}

	t := LinkBranding{}
	t.IsDefault = isDefault

	respBody, _, err := c.Post(ctx, "PATCH", "/whitelabel/links/"+id, t)
	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	return parseLinkBranding(respBody)
}

func (c *Client) ValidateLinkBranding(ctx context.Context, id string) RequestError {
	if id == "" {
		return RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrLinkBrandingIDRequired,
		}
	}

	_, statusCode, err := c.Post(ctx, "POST", "/whitelabel/links/"+id+"/validate", nil)
	if err != nil || statusCode != 200 {
		return RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	return RequestError{StatusCode: http.StatusOK, Err: nil}
}

// DeleteLinkBranding deletes an LinkBranding.
func (c *Client) DeleteLinkBranding(ctx context.Context, id string) (bool, RequestError) {
	if id == "" {
		return false, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrLinkBrandingIDRequired,
		}
	}

	responseBody, statusCode, err := c.Get(ctx, "DELETE", "/whitelabel/links/"+id)
	if err != nil {
		return false, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	if statusCode >= http.StatusMultipleChoices && statusCode != http.StatusNotFound { // ignore not found
		return false, RequestError{
			StatusCode: statusCode,
			Err:        fmt.Errorf("%w, status: %d, response: %s", ErrFailedDeletingLinkBranding, statusCode, responseBody),
		}
	}

	return true, RequestError{StatusCode: http.StatusOK, Err: nil}
}
