package sendgrid

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type DomainAuthenticationDNS struct {
	MailCNAME    DomainAuthenticationDNSValue `json:"mail_cname,omitempty"` //nolint:tagliatelle
	DKIM1        DomainAuthenticationDNSValue `json:"dkim1,omitempty"`
	DKIM2        DomainAuthenticationDNSValue `json:"dkim2,omitempty"`
	MailServer   DomainAuthenticationDNSValue `json:"mail_server,omitempty"`   //nolint:tagliatelle
	SubDomainSPF DomainAuthenticationDNSValue `json:"subdomain_spf,omitempty"` //nolint:tagliatelle
	DKIM         DomainAuthenticationDNSValue `json:"dkim,omitempty"`
}

type DomainAuthenticationDNSValue struct {
	Valid bool   `json:"valid,omitempty"`
	Type  string `json:"type,omitempty"`
	Host  string `json:"host,omitempty"`
	Data  string `json:"data,omitempty"`
}

// DomainAuthentication is a Sendgrid domain authentication.
type DomainAuthentication struct { //nolint:maligned
	ID                 int32                   `json:"id,omitempty"`
	UserID             int32                   `json:"user_id,omitempty"` //nolint:tagliatelle
	Domain             string                  `json:"domain,omitempty"`
	Subdomain          string                  `json:"subdomain,omitempty"`
	Username           string                  `json:"username,omitempty"`
	IPs                []string                `json:"ips,omitempty"`
	CustomSPF          bool                    `json:"custom_spf"` //nolint:tagliatelle
	IsDefault          bool                    `json:"default"`
	AutomaticSecurity  bool                    `json:"automatic_security"`   //nolint:tagliatelle
	CustomDKIMSelector string                  `json:"custom_dkim_selector"` //nolint:tagliatelle
	Legacy             bool                    `json:"legacy,omitempty"`
	Valid              bool                    `json:"valid,omitempty"`
	DNS                DomainAuthenticationDNS `json:"dns,omitempty"`
}

func ParseDomainAuthentication(respBody string) (*DomainAuthentication, RequestError) {
	var body DomainAuthentication
	if err := json.Unmarshal([]byte(respBody), &body); err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed parsing API key: %w", err),
		}
	}

	return &body, RequestError{StatusCode: http.StatusOK, Err: nil}
}

// CreateDomainAuthentication creates an DomainAuthentication and returns it.
func (c *Client) CreateDomainAuthentication(
	ctx context.Context,
	domain string,
	subdomain string,
	ips []string,
	customSpf bool,
	isDefault bool,
	automaticSecurity bool,
	customDKIMSelector string,
) (*DomainAuthentication, RequestError) {
	if domain == "" {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrNameRequired,
		}
	}

	respBody, statusCode, err := c.Post(ctx, "POST", "/whitelabel/domains", DomainAuthentication{
		Domain:             domain,
		Subdomain:          subdomain,
		IPs:                ips,
		CustomSPF:          customSpf,
		IsDefault:          isDefault,
		AutomaticSecurity:  automaticSecurity,
		CustomDKIMSelector: customDKIMSelector,
	})
	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed creating domain authentication: %w", err),
		}
	}

	if statusCode >= http.StatusMultipleChoices {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        fmt.Errorf("%w, status: %d, response: %s", ErrFailedCreatingDomainAuthentication, statusCode, respBody),
		}
	}

	return ParseDomainAuthentication(respBody)
}

// ReadDomainAuthentication retrieves an DomainAuthentication and returns it.
func (c *Client) ReadDomainAuthentication(ctx context.Context, id string) (*DomainAuthentication, RequestError) {
	if id == "" {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrDomainAuthenticationIDRequired,
		}
	}

	respBody, _, err := c.Get(ctx, "GET", "/whitelabel/domains/"+id)
	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	return ParseDomainAuthentication(respBody)
}

// UpdateDomainAuthentication edits an DomainAuthentication and returns it.
func (c *Client) UpdateDomainAuthentication(
	ctx context.Context,
	id string,
	isDefault bool,
	customSPF bool,
) (*DomainAuthentication, RequestError) {
	if id == "" {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrDomainAuthenticationIDRequired,
		}
	}

	t := DomainAuthentication{}
	t.IsDefault = isDefault
	t.CustomSPF = customSPF

	respBody, _, err := c.Post(ctx, "PATCH", "/whitelabel/domains/"+id, t)
	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	return ParseDomainAuthentication(respBody)
}

func (c *Client) ValidateDomainAuthentication(ctx context.Context, id string) RequestError {
	if id == "" {
		return RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrDomainAuthenticationIDRequired,
		}
	}

	respBody, statusCode, err := c.Post(ctx, "POST", "/whitelabel/domains/"+id+"/validate", map[string]interface{}{})
	if err != nil || statusCode != 200 {
		return RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	res, err := unmarshal(respBody)
	if err != nil {
		return RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	if valid, ok := res["valid"].(bool); ok && !valid {
		return RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("%w: %s", ErrDomainAuthenticationValidationFailed, respBody),
		}
	}

	return RequestError{StatusCode: http.StatusOK, Err: nil}
}

func unmarshal(respBody string) (map[string]interface{}, error) {
	var j map[string]interface{}
	if err := json.Unmarshal([]byte(respBody), &j); err != nil {
		log.Printf("JSON Unmarshalling of the response body: %s, failed with: %s,",
			respBody, err)
		return nil, err
	}
	return j, nil
}

// DeleteDomainAuthentication deletes an DomainAuthentication.
func (c *Client) DeleteDomainAuthentication(ctx context.Context, id string) (bool, RequestError) {
	if id == "" {
		return false, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrDomainAuthenticationIDRequired,
		}
	}

	responseBody, statusCode, err := c.Get(ctx, "DELETE", "/whitelabel/domains/"+id)
	if err != nil {
		return false, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	if statusCode >= http.StatusMultipleChoices && statusCode != http.StatusNotFound { // ignore not found
		return false, RequestError{
			StatusCode: statusCode,
			Err: fmt.Errorf("%w, status: %d, response: %s",
				ErrFailedDeletingDomainAuthentication, statusCode, responseBody),
		}
	}

	return true, RequestError{StatusCode: http.StatusOK, Err: nil}
}
