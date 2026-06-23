/*
Provide a resource to manage a sendgrid teammate.
Example Usage
```hcl

	resource "sendgrid_teammate" "user" {
		email    = "arslanbekov@gmail.com"
		is_admin = false
		scopes   = [
			"mail.send"
		]
	}

```
*/
package sendgrid

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// validSendgridScopes contains the actual list of valid SendGrid scopes
// Retrieved from https://api.sendgrid.com/v3/scopes as of 2024
var validSendgridScopes = map[string]bool{
	"access_settings.activity.read":             true,
	"access_settings.whitelist.create":          true,
	"access_settings.whitelist.delete":          true,
	"access_settings.whitelist.read":            true,
	"access_settings.whitelist.update":          true,
	"alerts.create":                             true,
	"alerts.delete":                             true,
	"alerts.read":                               true,
	"alerts.update":                             true,
	"api_keys.create":                           true,
	"api_keys.delete":                           true,
	"api_keys.read":                             true,
	"api_keys.update":                           true,
	"asm.groups.create":                         true,
	"asm.groups.delete":                         true,
	"asm.groups.read":                           true,
	"asm.groups.suppressions.create":            true,
	"asm.groups.suppressions.delete":            true,
	"asm.groups.suppressions.read":              true,
	"asm.groups.suppressions.update":            true,
	"asm.groups.update":                         true,
	"asm.suppressions.global.create":            true,
	"asm.suppressions.global.delete":            true,
	"asm.suppressions.global.read":              true,
	"asm.suppressions.global.update":            true,
	"billing.create":                            true,
	"billing.delete":                            true,
	"billing.read":                              true,
	"billing.update":                            true,
	"browsers.stats.read":                       true,
	"categories.create":                         true,
	"categories.delete":                         true,
	"categories.read":                           true,
	"categories.stats.read":                     true,
	"categories.stats.sums.read":                true,
	"categories.update":                         true,
	"clients.desktop.stats.read":                true,
	"clients.phone.stats.read":                  true,
	"clients.stats.read":                        true,
	"clients.tablet.stats.read":                 true,
	"clients.webmail.stats.read":                true,
	"credentials.create":                        true,
	"credentials.delete":                        true,
	"credentials.read":                          true,
	"credentials.update":                        true,
	"design_library.create":                     true,
	"design_library.delete":                     true,
	"design_library.read":                       true,
	"design_library.update":                     true,
	"devices.stats.read":                        true,
	"di.bounce_block_classification.read":       true,
	"email_testing.read":                        true,
	"email_testing.write":                       true,
	"geo.stats.read":                            true,
	"ips.assigned.read":                         true,
	"ips.create":                                true,
	"ips.delete":                                true,
	"ips.pools.create":                          true,
	"ips.pools.delete":                          true,
	"ips.pools.ips.create":                      true,
	"ips.pools.ips.delete":                      true,
	"ips.pools.ips.read":                        true,
	"ips.pools.ips.update":                      true,
	"ips.pools.read":                            true,
	"ips.pools.update":                          true,
	"ips.read":                                  true,
	"ips.update":                                true,
	"ips.warmup.create":                         true,
	"ips.warmup.delete":                         true,
	"ips.warmup.read":                           true,
	"ips.warmup.update":                         true,
	"mail.batch.create":                         true,
	"mail.batch.delete":                         true,
	"mail.batch.read":                           true,
	"mail.batch.update":                         true,
	"mail.send":                                 true,
	"mail_settings.address_whitelist.create":    true,
	"mail_settings.address_whitelist.delete":    true,
	"mail_settings.address_whitelist.read":      true,
	"mail_settings.address_whitelist.update":    true,
	"mail_settings.bcc.create":                  true,
	"mail_settings.bcc.delete":                  true,
	"mail_settings.bcc.read":                    true,
	"mail_settings.bcc.update":                  true,
	"mail_settings.bounce_purge.create":         true,
	"mail_settings.bounce_purge.delete":         true,
	"mail_settings.bounce_purge.read":           true,
	"mail_settings.bounce_purge.update":         true,
	"mail_settings.footer.create":               true,
	"mail_settings.footer.delete":               true,
	"mail_settings.footer.read":                 true,
	"mail_settings.footer.update":               true,
	"mail_settings.forward_bounce.create":       true,
	"mail_settings.forward_bounce.delete":       true,
	"mail_settings.forward_bounce.read":         true,
	"mail_settings.forward_bounce.update":       true,
	"mail_settings.forward_spam.create":         true,
	"mail_settings.forward_spam.delete":         true,
	"mail_settings.forward_spam.read":           true,
	"mail_settings.forward_spam.update":         true,
	"mail_settings.plain_content.create":        true,
	"mail_settings.plain_content.delete":        true,
	"mail_settings.plain_content.read":          true,
	"mail_settings.plain_content.update":        true,
	"mail_settings.read":                        true,
	"mail_settings.spam_check.create":           true,
	"mail_settings.spam_check.delete":           true,
	"mail_settings.spam_check.read":             true,
	"mail_settings.spam_check.update":           true,
	"mail_settings.template.create":             true,
	"mail_settings.template.delete":             true,
	"mail_settings.template.read":               true,
	"mail_settings.template.update":             true,
	"mailbox_providers.stats.read":              true,
	"marketing.automation.read":                 true,
	"marketing.read":                            true,
	"messages.read":                             true,
	"newsletter.create":                         true,
	"newsletter.delete":                         true,
	"newsletter.read":                           true,
	"newsletter.update":                         true,
	"partner_settings.new_relic.create":         true,
	"partner_settings.new_relic.delete":         true,
	"partner_settings.new_relic.read":           true,
	"partner_settings.new_relic.update":         true,
	"partner_settings.read":                     true,
	"partner_settings.sendwithus.create":        true,
	"partner_settings.sendwithus.delete":        true,
	"partner_settings.sendwithus.read":          true,
	"partner_settings.sendwithus.update":        true,
	"recipients.erasejob.create":                true,
	"recipients.erasejob.read":                  true,
	"sender_verification_eligible":              true,
	"signup.trigger_confirmation":               true,
	"sso.settings.create":                       true,
	"sso.settings.delete":                       true,
	"sso.settings.read":                         true,
	"sso.settings.update":                       true,
	"sso.teammates.create":                      true,
	"sso.teammates.update":                      true,
	"stats.global.read":                         true,
	"stats.read":                                true,
	"subusers.create":                           true,
	"subusers.credits.create":                   true,
	"subusers.credits.delete":                   true,
	"subusers.credits.read":                     true,
	"subusers.credits.remaining.create":         true,
	"subusers.credits.remaining.delete":         true,
	"subusers.credits.remaining.read":           true,
	"subusers.credits.remaining.update":         true,
	"subusers.credits.update":                   true,
	"subusers.delete":                           true,
	"subusers.monitor.create":                   true,
	"subusers.monitor.delete":                   true,
	"subusers.monitor.read":                     true,
	"subusers.monitor.update":                   true,
	"subusers.read":                             true,
	"subusers.reputations.read":                 true,
	"subusers.stats.monthly.read":               true,
	"subusers.stats.read":                       true,
	"subusers.stats.sums.read":                  true,
	"subusers.summary.read":                     true,
	"subusers.update":                           true,
	"suppression.blocks.create":                 true,
	"suppression.blocks.delete":                 true,
	"suppression.blocks.read":                   true,
	"suppression.blocks.update":                 true,
	"suppression.bounces.create":                true,
	"suppression.bounces.delete":                true,
	"suppression.bounces.read":                  true,
	"suppression.bounces.update":                true,
	"suppression.create":                        true,
	"suppression.delete":                        true,
	"suppression.invalid_emails.create":         true,
	"suppression.invalid_emails.delete":         true,
	"suppression.invalid_emails.read":           true,
	"suppression.invalid_emails.update":         true,
	"suppression.read":                          true,
	"suppression.spam_reports.create":           true,
	"suppression.spam_reports.delete":           true,
	"suppression.spam_reports.read":             true,
	"suppression.spam_reports.update":           true,
	"suppression.unsubscribes.create":           true,
	"suppression.unsubscribes.delete":           true,
	"suppression.unsubscribes.read":             true,
	"suppression.unsubscribes.update":           true,
	"suppression.update":                        true,
	"teammates.create":                          true,
	"teammates.delete":                          true,
	"teammates.read":                            true,
	"teammates.update":                          true,
	"templates.create":                          true,
	"templates.delete":                          true,
	"templates.read":                            true,
	"templates.update":                          true,
	"templates.versions.activate.create":        true,
	"templates.versions.activate.delete":        true,
	"templates.versions.activate.read":          true,
	"templates.versions.activate.update":        true,
	"templates.versions.create":                 true,
	"templates.versions.delete":                 true,
	"templates.versions.read":                   true,
	"templates.versions.update":                 true,
	"tracking_settings.click.create":            true,
	"tracking_settings.click.delete":            true,
	"tracking_settings.click.read":              true,
	"tracking_settings.click.update":            true,
	"tracking_settings.google_analytics.create": true,
	"tracking_settings.google_analytics.delete": true,
	"tracking_settings.google_analytics.read":   true,
	"tracking_settings.google_analytics.update": true,
	"tracking_settings.open.create":             true,
	"tracking_settings.open.delete":             true,
	"tracking_settings.open.read":               true,
	"tracking_settings.open.update":             true,
	"tracking_settings.read":                    true,
	"tracking_settings.subscription.create":     true,
	"tracking_settings.subscription.delete":     true,
	"tracking_settings.subscription.read":       true,
	"tracking_settings.subscription.update":     true,
	"ui.confirm_email":                          true,
	"ui.provision":                              true,
	"ui.signup_complete":                        true,
	"user.account.read":                         true,
	"user.credits.read":                         true,
	"user.email.read":                           true,
	"user.profile.create":                       true,
	"user.profile.delete":                       true,
	"user.profile.read":                         true,
	"user.profile.update":                       true,
	"user.scheduled_sends.create":               true,
	"user.scheduled_sends.delete":               true,
	"user.scheduled_sends.read":                 true,
	"user.scheduled_sends.update":               true,
	"user.settings.enforced_tls.read":           true,
	"user.settings.enforced_tls.update":         true,
	"user.timezone.create":                      true,
	"user.timezone.delete":                      true,
	"user.timezone.read":                        true,
	"user.timezone.update":                      true,
	"user.username.read":                        true,
	"user.webhooks.event.settings.create":       true,
	"user.webhooks.event.settings.delete":       true,
	"user.webhooks.event.settings.read":         true,
	"user.webhooks.event.settings.update":       true,
	"user.webhooks.event.test.create":           true,
	"user.webhooks.event.test.delete":           true,
	"user.webhooks.event.test.read":             true,
	"user.webhooks.event.test.update":           true,
	"user.webhooks.parse.settings.create":       true,
	"user.webhooks.parse.settings.delete":       true,
	"user.webhooks.parse.settings.read":         true,
	"user.webhooks.parse.settings.update":       true,
	"user.webhooks.parse.stats.read":            true,
	"validations.email.create":                  true,
	"validations.email.read":                    true,
	"whitelabel.create":                         true,
	"whitelabel.delete":                         true,
	"whitelabel.read":                           true,
	"whitelabel.update":                         true,
}

// sendgridAutomaticScopes are scopes that SendGrid sets automatically and should not be included in user input
var sendgridAutomaticScopes = map[string]bool{
	"2fa_exempt":                 true,
	"2fa_required":               true,
	"sender_verification_legacy": true, // SendGrid manages this scope automatically
}

func resourceSendgridTeammate() *schema.Resource {
	return &schema.Resource{
		Description: `Manages a SendGrid teammate. Teammates are team members who have access to your SendGrid account with specific permissions.

**Important Notes:**
- Admin teammates have full access and don't need scopes
- Scopes '2fa_exempt' and '2fa_required' are set automatically by SendGrid
- Some scopes require specific SendGrid plans (Pro+, Marketing plans, etc.)
- Use timeouts for better reliability with rate limiting`,

		CreateContext: resourceSendgridTeammateCreate,
		ReadContext:   resourceSendgridTeammateRead,
		UpdateContext: resourceSendgridTeammateUpdate,
		DeleteContext: resourceSendgridTeammateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// subuser_access only applies to SSO teammates; the create/update paths
		// silently ignore it for non-SSO users, so reject it at plan time instead.
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			if v, ok := d.GetOk("subuser_access"); ok && v.(*schema.Set).Len() > 0 && !d.Get("is_sso").(bool) {
				return fmt.Errorf("subuser_access is only supported for SSO teammates (is_sso = true)")
			}
			return nil
		},

		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Description: "The email address of the teammate. This will be used as the teammate's login.",
				Required:    true,
			},
			"first_name": {
				Type:             schema.TypeString,
				Description:      "The first name of the teammate. Required for SSO users.",
				Optional:         true,
				DiffSuppressFunc: suppressDiffForPendingUsers,
			},
			"last_name": {
				Type:             schema.TypeString,
				Description:      "The last name of the teammate. Required for SSO users.",
				Optional:         true,
				DiffSuppressFunc: suppressDiffForPendingUsers,
			},
			"is_admin": {
				Type:        schema.TypeBool,
				Description: "Whether the teammate should have admin privileges. Admin teammates have full access to the account and don't need specific scopes.",
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeBool,
				},
			},
			"is_sso": {
				Type:        schema.TypeBool,
				Description: "Whether this is a Single Sign-On (SSO) user. SSO users require first_name and last_name.",
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeBool,
				},
			},
			"scopes": {
				Type:        schema.TypeSet,
				Description: "List of permission scopes for the teammate. Ignored if is_admin is true. Cannot include '2fa_exempt' or '2fa_required' as these are managed automatically by SendGrid. See SendGrid API documentation for available scopes.",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"username": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The username for the teammate. If not provided, the email will be used.",
				DiffSuppressFunc: suppressDiffForPendingUsers,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"user_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the user: 'active' for confirmed users, 'pending' for users who haven't accepted their invitation yet.",
			},
			"subuser_access": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Description: "Subuser permission grants for this SSO teammate. Only applies when `is_sso = true`. " +
					"Setting at least one block sets `has_restricted_subuser_access = true` on the API. " +
					"This attribute is also computed: when no block is managed it reflects the teammate's " +
					"current server-side access without producing a diff, and removing all blocks falls back to " +
					"that server value rather than clearing access (Terraform does not clobber access it does not manage).",
				// Hash on the subuser id so a teammate has one logical block per subuser
				// and reordering or scope echo-back from the API does not produce churn.
				Set: func(v interface{}) int {
					m := v.(map[string]interface{})
					return m["id"].(int)
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Numeric subuser account ID.",
						},
						"permission_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"admin", "restricted"}, false),
							Description:  `Permission level for this subuser: "admin" (full access) or "restricted" (limited to scopes). Per the SendGrid API these are the only two values.`,
						},
						"scopes": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Scopes granted on this subuser. Required when permission_type is \"restricted\"; ignored for \"admin\".",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

// validateTeammateScopes validates the scopes provided for a teammate
func validateTeammateScopes(v interface{}, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	scopes := v.(*schema.Set).List()

	var invalidScopes []string
	var automaticScopes []string

	for _, scope := range scopes {
		scopeStr := scope.(string)

		// Check for automatic scopes that shouldn't be set manually
		if sendgridAutomaticScopes[scopeStr] {
			automaticScopes = append(automaticScopes, scopeStr)
			continue
		}

		// Check for invalid scopes
		if !validSendgridScopes[scopeStr] {
			invalidScopes = append(invalidScopes, scopeStr)
		}
	}

	if len(automaticScopes) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Automatic scopes cannot be manually assigned",
			Detail: fmt.Sprintf(
				"the following scopes are set automatically by SendGrid and cannot be manually assigned: %s",
				strings.Join(automaticScopes, ", "),
			),
			AttributePath: path,
		})
	}

	if len(invalidScopes) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid or unassignable scopes",
			Detail: fmt.Sprintf(
				"the following scopes are not valid or assignable: %s. Please check the SendGrid API documentation for valid scopes",
				strings.Join(invalidScopes, ", "),
			),
			AttributePath: path,
		})
	}

	return diags
}

// sanitizeScopes removes automatic scopes that SendGrid sets automatically
func sanitizeScopes(scopes []string) []string {
	var sanitized []string
	for _, scope := range scopes {
		if !sendgridAutomaticScopes[scope] {
			sanitized = append(sanitized, scope)
		}
	}
	return sanitized
}

// suppressDiffForPendingUsers suppresses diff for fields that are not available for pending users
func suppressDiffForPendingUsers(k, old, new string, d *schema.ResourceData) bool {
	userStatus := d.Get("user_status").(string)
	isSSO := d.Get("is_sso").(bool)

	// For pending users, suppress diff if old value is empty and new value is set
	// This prevents Terraform from showing changes for fields that can't be set until user accepts invitation
	if userStatus == "pending" {
		return old == "" && new != ""
	}

	// For non-SSO users, first_name and last_name cannot be updated via API
	// Suppress diff when trying to change these fields from what API returned
	if !isSSO && (k == "first_name" || k == "last_name") {
		// Allow setting to what API returned (old value), but prevent changes
		return old != new
	}

	return false
}

// extractSubuserAccess converts the TypeSet value from the schema into the SDK slice.
func extractSubuserAccess(d *schema.ResourceData) []sendgrid.SubuserAccess {
	raw := d.Get("subuser_access").(*schema.Set).List()
	if len(raw) == 0 {
		return nil
	}

	out := make([]sendgrid.SubuserAccess, 0, len(raw))
	for _, item := range raw {
		m := item.(map[string]interface{})
		entry := sendgrid.SubuserAccess{
			ID:             m["id"].(int),
			PermissionType: m["permission_type"].(string),
		}
		if scopesRaw, ok := m["scopes"]; ok {
			for _, s := range scopesRaw.(*schema.Set).List() {
				entry.Scopes = append(entry.Scopes, s.(string))
			}
		}
		out = append(out, entry)
	}
	return out
}

// enhancedRetryOnScopeErrors wraps the standard retry function with enhanced error handling for scope-related errors
func enhancedRetryOnScopeErrors(ctx context.Context, d *schema.ResourceData, f func() (interface{}, sendgrid.RequestError)) (interface{}, error) {
	resp, err := sendgrid.RetryOnRateLimit(ctx, d, f)
	if err != nil {
		// Check if this is a scope-related error
		if strings.Contains(err.Error(), "invalid or unassignable scopes") {
			return nil, fmt.Errorf(`request failed due to invalid scopes. This can happen when:
1. You provided an invalid scope name (check SendGrid API documentation for valid scopes)
2. Your SendGrid plan doesn't support certain scopes
3. You included automatically managed scopes like '2fa_exempt' or '2fa_required'

Original error: %w

Tip: Use 'terraform plan' to validate your configuration before applying`, err)
		}

		// Check for user cancellation scenarios
		if strings.Contains(err.Error(), "context canceled") || strings.Contains(err.Error(), "operation was canceled") {
			return nil, fmt.Errorf(`operation was canceled. If you canceled the operation during execution, some resources may be in an intermediate state. Please check your SendGrid dashboard and run 'terraform refresh' to update the state.

Original error: %w`, err)
		}
	}
	return resp, err
}

func resourceSendgridTeammateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.NewClient("")
	email := d.Get("email").(string)
	isAdmin := d.Get("is_admin").(bool)
	isSSO := d.Get("is_sso").(bool)
	firstName := d.Get("first_name").(string)
	lastName := d.Get("last_name").(string)

	scopesSet := d.Get("scopes").(*schema.Set)

	// Validate scopes if not admin
	if !isAdmin && scopesSet.Len() > 0 {
		path := cty.GetAttrPath("scopes")
		if diags := validateTeammateScopes(scopesSet, path); diags.HasError() {
			return diags
		}
	}

	var scopes []string
	if !isAdmin {
		scopesList := scopesSet.List()
		for _, scope := range scopesList {
			scopes = append(scopes, scope.(string))
		}
		// Sanitize scopes to remove any automatic ones
		scopes = sanitizeScopes(scopes)
	}

	tflog.Debug(ctx, "Creating teammate", map[string]interface{}{
		"first_name": firstName, "last_name": lastName,
		"email": email, "is_admin": isAdmin, "scopes": scopes,
	})

	userStruct, err := enhancedRetryOnScopeErrors(ctx, d, func() (interface{}, sendgrid.RequestError) {
		if isSSO {
			return client.CreateSSOUser(ctx, firstName, lastName, email, scopes, isAdmin)
		} else {
			return client.CreateUser(ctx, email, scopes, isAdmin)
		}
	})
	if err != nil {
		return diag.FromErr(err)
	}

	user := userStruct.(*sendgrid.User)
	d.SetId(user.Email)
	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(err)
	}

	// Apply subuser_access immediately after creation (SSO only).
	// The create endpoint does not accept subuser_access, so we do a follow-up PATCH.
	// The teammate ID is already set above, so if this PATCH fails the teammate
	// exists in state without subuser access; the error makes that explicit and the
	// next apply reconciles it.
	if isSSO {
		subuserAccess := extractSubuserAccess(d)
		if len(subuserAccess) > 0 {
			_, saErr := enhancedRetryOnScopeErrors(ctx, d, func() (interface{}, sendgrid.RequestError) {
				return client.UpdateSSOUserWithSubuserAccess(ctx, firstName, lastName, email, scopes, isAdmin, subuserAccess)
			})
			if saErr != nil {
				return diag.FromErr(fmt.Errorf(
					"teammate %s was created but applying subuser_access failed; "+
						"the teammate exists without subuser access and the next apply will reconcile it: %w",
					email, saErr))
			}
		}
	}

	return resourceSendgridTeammateRead(ctx, d, meta)
}

func resourceSendgridTeammateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.NewClient("")

	var diags diag.Diagnostics
	email := d.Id()

	teammate, readErr := client.ReadUser(ctx, email)
	if readErr.Err != nil {
		if readErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return append(diags, diag.FromErr(readErr.Err)...)
	}

	// There is no need to track admin scopes since they have full access.
	if teammate.IsAdmin {
		teammate.Scopes = nil
	}

	var filteredScopes []string
	for _, s := range teammate.Scopes {
		// Sendgrid sets these scopes automatically. If you try to set them, you will get a 400 error.
		if !sendgridAutomaticScopes[s] {
			filteredScopes = append(filteredScopes, s)
		}
	}

	// Sort scopes to ensure consistent ordering and prevent drift
	sort.Strings(filteredScopes)

	// Determine user status based on UserType
	userStatus := "active"
	if teammate.UserType == "pending" {
		userStatus = "pending"
	}

	d.SetId(teammate.Email)

	// Always set all fields that API returns, regardless of user type
	// DiffSuppressFunc will handle preventing unwanted changes
	retErr := multierror.Append(
		d.Set("email", teammate.Email),
		d.Set("username", teammate.Username),
		d.Set("first_name", teammate.FirstName),
		d.Set("last_name", teammate.LastName),
		d.Set("scopes", filteredScopes),
		d.Set("is_admin", teammate.IsAdmin),
		d.Set("user_status", userStatus),
	)
	if retErr.ErrorOrNil() != nil {
		return diag.FromErr(retErr.ErrorOrNil())
	}

	// Read subuser_access for active SSO teammates.
	isSSO := d.Get("is_sso").(bool)
	if isSSO && userStatus == "active" {
		saResp, saErr := client.ReadSubuserAccess(ctx, email)
		if saErr.Err != nil {
			// Propagate the real status code rather than masking it (see #95):
			// a vanished teammate (404) clears state; auth/server errors (401/5xx)
			// surface as errors instead of silently leaving stale state behind.
			if saErr.StatusCode == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return append(diags, diag.FromErr(saErr.Err)...)
		}
		if err := d.Set("subuser_access", flattenSubuserAccess(saResp.SubuserAccess)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

// flattenSubuserAccess converts the API read response into the schema shape.
// Scopes are only meaningful for "restricted" entries; for "admin" (full access)
// the API may echo scopes that cannot be managed, so they are dropped to avoid a
// perpetual diff. Automatic scopes are stripped as for top-level scopes.
func flattenSubuserAccess(entries []sendgrid.SubuserAccessRead) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		m := map[string]interface{}{
			"id":              entry.ID,
			"permission_type": entry.PermissionType,
			"scopes":          []string{},
		}
		if entry.PermissionType == "restricted" {
			m["scopes"] = sanitizeScopes(entry.Scopes)
		}
		out = append(out, m)
	}
	return out
}

func resourceSendgridTeammateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.NewClient("")
	email := d.Get("email").(string)
	isAdmin := d.Get("is_admin").(bool)
	isSSO := d.Get("is_sso").(bool)
	firstName := d.Get("first_name").(string)
	lastName := d.Get("last_name").(string)

	// Check if user is pending - pending users are read-only after invitation is sent
	userStatus := d.Get("user_status").(string)
	if userStatus == "pending" {
		tflog.Info(ctx, "Pending user detected - skipping update. Pending users are read-only until they accept their invitation", map[string]interface{}{
			"email": email,
		})
		// For pending users, we only refresh their current state
		return resourceSendgridTeammateRead(ctx, d, meta)
	}

	scopesSet := d.Get("scopes").(*schema.Set)

	// Validate scopes if not admin
	if !isAdmin && scopesSet.Len() > 0 {
		path := cty.GetAttrPath("scopes")
		if diags := validateTeammateScopes(scopesSet, path); diags.HasError() {
			return diags
		}
	}

	var scopes []string
	if !isAdmin {
		scopesList := scopesSet.List()
		for _, scope := range scopesList {
			scopes = append(scopes, scope.(string))
		}
		// Sanitize scopes to remove any automatic ones
		scopes = sanitizeScopes(scopes)
	}

	_, err := enhancedRetryOnScopeErrors(ctx, d, func() (interface{}, sendgrid.RequestError) {
		if isSSO {
			subuserAccess := extractSubuserAccess(d)
			return client.UpdateSSOUserWithSubuserAccess(ctx, firstName, lastName, email, scopes, isAdmin, subuserAccess)
		}
		return client.UpdateUser(ctx, email, scopes, isAdmin)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSendgridTeammateRead(ctx, d, meta)
}

func resourceSendgridTeammateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.NewClient("")

	var diags diag.Diagnostics
	userEmail := d.Id()

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return client.DeleteUser(ctx, userEmail)
	})
	if err != nil {
		// Enhanced error handling for delete operations
		if strings.Contains(err.Error(), "context canceled") || strings.Contains(err.Error(), "operation was canceled") {
			return append(diags, diag.Errorf("Delete operation was canceled. The teammate may still exist in SendGrid. Please check your SendGrid dashboard and re-run the delete operation if needed.")...)
		}
		return append(diags, diag.FromErr(err)...)
	}
	return diags
}
