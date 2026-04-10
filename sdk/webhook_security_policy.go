package sendgrid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Request Model

type OAuthConfigRequest struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	TokenURL     string   `json:"token_url"`
	Scopes       []string `json:"scopes"`
}

type SignatureConfigRequest struct {
	Enabled bool `json:"enabled"`
}

type WebhookSecurityPolicyRequest struct {
	Name      string                  `json:"name"`
	OAuth     *OAuthConfigRequest     `json:"oauth,omitempty"`
	Signature *SignatureConfigRequest `json:"signature,omitempty"`
}

// Response Model

type WebhookSecurityPolicyResponse struct {
	Policy WebhookSecurityPolicyResult `json:"policy"`
}

type WebhookSecurityPolicyResult struct {
	ID        string                        `json:"id,omitempty"`
	Name      string                        `json:"name,omitempty"`
	OAuth     *WebhookOAuthPolicyResponse   `json:"oauth,omitempty"`
	Signature *WebhookSignatureKeysResponse `json:"signature,omitempty"`
}

type WebhookOAuthPolicyResponse struct {
	ClientID string   `json:"client_id,omitempty"`
	TokenURL string   `json:"token_url,omitempty"`
	Scopes   []string `json:"scopes,omitempty"`
}

type WebhookSignatureKeysResponse struct {
	PublicKey string `json:"public_key,omitempty"`
}

func parseSecurityPolicyResponse(respBody string) (*WebhookSecurityPolicyResponse, RequestError) {
	var body WebhookSecurityPolicyResponse
	if err := json.Unmarshal([]byte(respBody), &body); err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed parsing security policy response: %w", err),
		}
	}

	return &body, RequestError{
		StatusCode: http.StatusOK,
		Err:        nil,
	}
}

func convertInterfaceArrayToStringArray(input []interface{}) []string {
	if input == nil {
		return []string{}
	}

	result := make([]string, len(input))
	for i, v := range input {
		if str, ok := v.(string); ok {
			result[i] = str
		} else {
			result[i] = fmt.Sprintf("%v", v)
		}
	}
	return result
}

// https://www.twilio.com/docs/sendgrid/api-reference/settings-inbound-parse/create-a-parse-webhook-security-policy
func (c *Client) CreateWebhookSecurityPolicy(ctx context.Context, name string, oauth map[string]interface{}, signature map[string]interface{}) (*WebhookSecurityPolicyResponse, RequestError) {
	if name == "" {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrNameRequired,
		}
	}

	request := WebhookSecurityPolicyRequest{
		Name: name,
	}

	if len(oauth) > 0 {
		request.OAuth = &OAuthConfigRequest{
			ClientID:     oauth["client_id"].(string),
			ClientSecret: oauth["client_secret"].(string),
			TokenURL:     oauth["token_url"].(string),
			Scopes:       convertInterfaceArrayToStringArray(oauth["scopes"].([]interface{})),
		}
	}

	if len(signature) > 0 {
		request.Signature = &SignatureConfigRequest{
			Enabled: signature["enabled"].(bool),
		}
	}

	respBody, statusCode, err := c.Post(ctx, "POST", "/user/webhooks/security/policies", request)

	if err != nil {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed creating webhook security policy: %w", err),
		}
	}

	if statusCode >= http.StatusMultipleChoices {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        fmt.Errorf("%w, status: %d, response: %s", ErrFailedCreatingWebhookSecurityPolicy, statusCode, respBody),
		}
	}

	return parseSecurityPolicyResponse(respBody)
}

// https://www.twilio.com/docs/sendgrid/api-reference/settings-inbound-parse/retrieve-a-specific-parse-security-policy
func (c *Client) ReadWebhookSecurityPolicy(ctx context.Context, policyId string) (*WebhookSecurityPolicyResponse, RequestError) {
	if policyId == "" {
		return nil, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrWebhookSecurityPolicyIdRequired,
		}
	}

	respBody, statusCode, err := c.Get(ctx, "GET", "/user/webhooks/security/policies/"+policyId)
	if err != nil {
		return nil, RequestError{
			StatusCode: statusCode,
			Err:        err,
		}
	}

	return parseSecurityPolicyResponse(respBody)
}

// https://www.twilio.com/docs/sendgrid/api-reference/settings-inbound-parse/update-a-parse-security-policy
func (c *Client) UpdateWebhookSecurityPolicy(ctx context.Context, policyId string, name string, oauth map[string]interface{}, signature map[string]interface{}) RequestError {
	if policyId == "" {
		return RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrWebhookSecurityPolicyIdRequired,
		}
	}

	if name == "" {
		return RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrNameRequired,
		}
	}

	request := WebhookSecurityPolicyRequest{
		Name: name,
	}

	// Add OAuth configuration if provided
	if len(oauth) > 0 {
		request.OAuth = &OAuthConfigRequest{
			ClientID:     oauth["client_id"].(string),
			ClientSecret: oauth["client_secret"].(string),
			TokenURL:     oauth["token_url"].(string),
			Scopes:       convertInterfaceArrayToStringArray(oauth["scopes"].([]interface{})),
		}
	}

	// Add Signature configuration if provided
	if len(signature) > 0 {
		request.Signature = &SignatureConfigRequest{
			Enabled: signature["enabled"].(bool),
		}
	}

	_, statusCode, err := c.Post(ctx, "PATCH", "/user/webhooks/security/policies/"+policyId, request)
	if err != nil {
		return RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed updating webhook security policy: %w", err),
		}
	}

	if statusCode >= http.StatusMultipleChoices {
		return RequestError{
			StatusCode: statusCode,
			Err:        fmt.Errorf("failed updating webhook security policy, status: %d", statusCode),
		}
	}

	return RequestError{
		StatusCode: http.StatusOK,
		Err:        nil,
	}
}

// https://www.twilio.com/docs/sendgrid/api-reference/settings-inbound-parse/delete-a-parse-security-policy
func (c *Client) DeleteWebhookSecurityPolicy(ctx context.Context, policyId string) (bool, RequestError) {
	if policyId == "" {
		return false, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        ErrWebhookSecurityPolicyIdRequired,
		}
	}

	responseBody, statusCode, err := c.Get(ctx, "DELETE", "/user/webhooks/security/policies/"+policyId)
	if err != nil {
		return false, RequestError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	if statusCode >= http.StatusMultipleChoices && statusCode != http.StatusNotFound { // ignore not found
		return false, RequestError{
			StatusCode: statusCode,
			Err:        fmt.Errorf("%w, status: %d, response: %s", ErrFailedDeletingWebhookSecurityPolicy, statusCode, responseBody),
		}
	}

	return true, RequestError{StatusCode: http.StatusOK, Err: nil}
}
