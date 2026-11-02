package sendgrid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	// ErrBodyNotNil low error displayed when the prepared body for a POST call
	// to the API is nil.
	ErrBodyNotNil = errors.New("body must not be nil")

	// ErrNameRequired error displayed when an api key name wasn't specified.
	ErrNameRequired = errors.New("a name is required")

	// ErrAPIKeyIDRequired error displayed when an api key ID wasn't specified.
	ErrAPIKeyIDRequired = errors.New("an API Key ID is required")

	// ErrFailedCreatingAPIKey error displayed when the provider can not create an api key.
	ErrFailedCreatingAPIKey = errors.New("failed creating apiKey")

	// ErrFailedDeletingAPIKey error displayed when the provider can not delete an api key.
	ErrFailedDeletingAPIKey = errors.New("failed deleting apiKey")

	// ErrUsernameRequired error displayed when a subUser username wasn't specified.
	ErrUsernameRequired = errors.New("a username is required")

	// ErrEmailRequired error displayed when a subUser email wasn't specified.
	ErrEmailRequired = errors.New("an email is required")

	// ErrPasswordRequired error displayed when a subUser password wasn't specified.
	ErrPasswordRequired = errors.New("a password is required")

	// ErrIPRequired error displayed when at least one IP per subUser wasn't specified.
	ErrIPRequired = errors.New("at least one ip address is required")

	// ErrFailedCreatingSubUser error displayed when the provider can not create a subuser.
	ErrFailedCreatingSubUser = errors.New("failed creating subUser")

	// ErrFailedDeletingSubUser error displayed when the provider can not delete a subuser.
	ErrFailedDeletingSubUser = errors.New("failed deleting subUser")

	// ErrTemplateIDRequired error displayed when a template ID wasn't specified.
	ErrTemplateIDRequired = errors.New("a template ID is required")

	// ErrTemplateNameRequired error displayed when a template name wasn't specified.
	ErrTemplateNameRequired = errors.New("a template name is required")

	// ErrTemplateVersionIDRequired error displayed when a template version ID wasn't specified.
	ErrTemplateVersionIDRequired = errors.New("a template version ID is required")

	// ErrTemplateVersionNameRequired error displayed when a template version ID wasn't specified.
	ErrTemplateVersionNameRequired = errors.New("a template version name is required")

	// ErrTemplateVersionSubjectRequired error displayed when a template version subject wasn't specified.
	ErrTemplateVersionSubjectRequired = errors.New("a template version subject is required")

	ErrFailedCreatingUnsubscribeGroup = errors.New("failed to create unsubscribe list")

	ErrUnsubscribeGroupIDRequired = errors.New("unsubscribe list id is required")

	ErrFailedDeletingUnsubscribeGroup = errors.New("failed deleting unsubscribe list")

	ErrFailedCreatingParseWebhook = errors.New("failed to create parse webhook")

	ErrFailedDeletingParseWebhook = errors.New("failed deleting parse webhook")

	ErrHostnameRequired = errors.New("a hostname is required")

	ErrURLRequired = errors.New("a url is required")

	ErrFailedPatchingEventWebhook = errors.New("failed to patch event webhook")

	// ErrEventWebhookIDRequired error displayed when an event webhook ID wasn't specified.
	ErrEventWebhookIDRequired = errors.New("an event webhook ID is required")

	// ErrFailedCreatingEventWebhook error displayed when the provider can not create an event webhook.
	ErrFailedCreatingEventWebhook = errors.New("failed to create event webhook")

	// ErrFailedUpdatingEventWebhook error displayed when the provider can not update an event webhook.
	ErrFailedUpdatingEventWebhook = errors.New("failed to update event webhook")

	// ErrFailedDeletingEventWebhook error displayed when the provider can not delete an event webhook.
	ErrFailedDeletingEventWebhook = errors.New("failed to delete event webhook")

	ErrFailedCreatingDomainAuthentication = errors.New("failed to create domain authentication")

	ErrDomainAuthenticationIDRequired = errors.New("id for domain authentication is required")

	ErrDomainAuthenticationValidationFailed = errors.New("domain authentication validation failed")

	ErrFailedDeletingDomainAuthentication = errors.New("failed deleting domain authentication")

	ErrFailedAssociatingDomainAuthenticationSubuser = errors.New("failed associating domain authentication with subuser")

	ErrFailedDisassociatingDomainAuthenticationSubuser = errors.New("failed disassociating domain authentication from subuser")

	ErrLinkBrandingIDRequired = errors.New("link branding id is required")

	ErrFailedDeletingLinkBranding = errors.New("failed to delete link branding")

	ErrFailedCreatingLinkBranding = errors.New("failed to create link branding")

	// ErrSubUserPassword should be empty.
	ErrSubUserPassword = errors.New("new password must be non empty")

	// ErrSSOIntegrationMissingField error displayed when a required SSO integration field is not specified.
	ErrSSOIntegrationMissingField = errors.New("SSO integration field is missing")

	// ErrFailedCreatingSSOIntegration error displayed when an SSO integration creation request fails.
	ErrFailedCreatingSSOIntegration = errors.New("failed to create SSO integration")

	// ErrFailedUpdatingSSOIntegration error displayed when an SSO integration update request fails.
	ErrFailedUpdatingSSOIntegration = errors.New("failed to update SSO integration")

	// ErrSSOCertificateMissingField error displayed when a required SSO certificate field is not specified.
	ErrSSOCertificateMissingField = errors.New("SSO certificate field is missing")

	// ErrFailedCreatingSSOCertificate error displayed when an SSO certificate creation request fails.
	ErrFailedCreatingSSOCertificate = errors.New("failed to create SSO certificate")

	// ErrFailedUpdatingSSOCertificate error displayed when an SSO certificate update request fails.
	ErrFailedUpdatingSSOCertificate = errors.New("failed to update SSO certificate")

	ErrWebhookSecurityPolicyIdRequired = errors.New("a webhook security policy id is required")

	ErrFailedCreatingWebhookSecurityPolicy = errors.New("failed creating webhook security policy")

	ErrFailedDeletingWebhookSecurityPolicy = errors.New("failed deleting webhook security policy")
)

type APIError struct {
	f interface{} // unknown
}

// RequestError struct permits to embed to return the statucode and the error to the parent function.
type RequestError struct {
	StatusCode int
	Err        error
}

type subUserError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
}

type subUserErrors struct {
	Errors []subUserError `json:"errors,omitempty"`
}

// parseErrorDetails attempts to parse SendGrid API error response for better error messages
func parseErrorDetails(err error) (string, bool) {
	errStr := err.Error()

	// Check for scope-related errors
	if strings.Contains(errStr, "invalid or unassignable scopes") {
		return `invalid or unassignable scopes provided. This can happen when:
1. Using invalid scope names (check SendGrid API documentation)
2. Your SendGrid plan doesn't support certain scopes
3. Including automatically managed scopes like '2fa_exempt' or '2fa_required'

Tip: Run 'terraform plan' first to validate your configuration`, true
	}

	// Check for permission errors
	if strings.Contains(errStr, "permission") || strings.Contains(errStr, "unauthorized") {
		return `permission denied. Check that:
1. Your API key has sufficient permissions
2. You're not trying to access features not available in your SendGrid plan
3. The API key hasn't been revoked or expired`, true
	}

	// Check for resource not found
	if strings.Contains(errStr, "not found") || strings.Contains(errStr, "404") {
		return "resource not found. It may have been deleted outside of Terraform or the ID is incorrect", true
	}

	// Check for validation errors
	if strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid") {
		return "validation error. Please check that all required fields are provided and values are in the correct format", true
	}

	return "", false
}

// enhanceError provides more helpful error messages based on the error type and context
func enhanceError(originalErr error, statusCode int) error {
	if originalErr == nil {
		return nil
	}

	// Try to parse for specific error details
	if enhancedMsg, enhanced := parseErrorDetails(originalErr); enhanced {
		return fmt.Errorf("%s\n\nOriginal error: %w", enhancedMsg, originalErr)
	}

	// Provide context based on status code
	switch statusCode {
	case http.StatusBadRequest:
		return fmt.Errorf(`bad request (HTTP 400). This usually means:
1. Invalid input data or parameters
2. Malformed request body
3. Business logic validation failure

Original error: %w

Tip: Check the SendGrid API documentation for the correct request format`, originalErr)

	case http.StatusUnauthorized:
		return fmt.Errorf(`unauthorized (HTTP 401). Check that:
1. Your SendGrid API key is correct
2. The API key hasn't been revoked
3. You have the necessary permissions

Original error: %w`, originalErr)

	case http.StatusForbidden:
		return fmt.Errorf(`forbidden (HTTP 403). This means:
1. Your API key lacks the required permissions
2. Your SendGrid plan doesn't support this feature
3. Resource access is restricted

Original error: %w`, originalErr)

	case http.StatusNotFound:
		return fmt.Errorf(`resource not found (HTTP 404). This means:
1. The resource was deleted outside of Terraform
2. The resource ID is incorrect
3. You don't have permission to view the resource

Original error: %w

Tip: Run 'terraform refresh' to sync the state with actual resources`, originalErr)

	case http.StatusTooManyRequests:
		return fmt.Errorf(`rate limit exceeded (HTTP 429). The provider will automatically retry, but you can:
1. Reduce parallelism: terraform apply -parallelism=1
2. Increase timeouts in your resource configuration
3. Check if you're hitting API limits in multiple processes

Original error: %w`, originalErr)

	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf(`SendGrid API server error (HTTP %d). This is usually temporary:
1. Check SendGrid status page: https://status.sendgrid.com/
2. Retry the operation after a few minutes
3. If persists, contact SendGrid support

Original error: %w`, statusCode, originalErr)

	default:
		return fmt.Errorf("request failed with HTTP %d: %w", statusCode, originalErr)
	}
}

// RetryOnRateLimit management of RequestErrors, and launch a retry if needed.
// Enhanced with better error handling and more informative error messages.
func RetryOnRateLimit(
	ctx context.Context, d *schema.ResourceData, f func() (interface{}, RequestError),
) (interface{}, error) {
	var resp interface{}

	err := retry.RetryContext(
		ctx,
		d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
			var requestErr RequestError
			resp, requestErr = f()

			if requestErr.Err != nil {
				// Always retry rate limit errors
				if requestErr.StatusCode == http.StatusTooManyRequests {
					return retry.RetryableError(requestErr.Err)
				}

				// Enhance the error message before returning
				enhancedErr := enhanceError(requestErr.Err, requestErr.StatusCode)
				return retry.NonRetryableError(enhancedErr)
			}

			return nil
		})

	if err != nil {
		// Check for context cancellation
		if strings.Contains(err.Error(), "context canceled") ||
			strings.Contains(err.Error(), "operation was canceled") ||
			strings.Contains(err.Error(), "context deadline exceeded") {
			return resp, fmt.Errorf(`operation was canceled or timed out. This can happen when:
1. You pressed Ctrl+C during execution
2. The operation took longer than the configured timeout
3. Network connectivity issues occurred

What to do next:
1. Check your SendGrid dashboard for partially created resources
2. Run 'terraform refresh' to update the state
3. Re-run the operation with increased timeout if needed

Original error: %w`, err)
		}

		return resp, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (e *APIError) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &e.f); err != nil {
		e.f = string(b)
	}
	return nil
}

func (e *APIError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.f)
}

func (e APIError) Detail() string {
	switch v := e.f.(type) {
	case map[string]interface{}:
		if len(v) == 1 {
			if detail, ok := v["detail"].(string); ok {
				return detail
			}
		}
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (e APIError) Error() string {
	return fmt.Sprintf("sendgrid: %s", e.Detail())
}

// Empty returns true if empty.
func (e APIError) Empty() bool {
	return e.f == nil
}
