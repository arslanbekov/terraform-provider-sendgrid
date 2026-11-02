/*
Resources List

API key Resource

	sendgrid_api_key

Domain authentication Resources

	sendgrid_domain_authentication
	sendgrid_domain_authentication_validation
	sendgrid_domain_authentication_subuser

Link branding Resource

	sendgrid_link_branding

SSO Resources

	sendgrid_sso_certificate
	sendgrid sso_integration

Subuser Resource

	sendgrid_subuser

Template Resources

	sendgrid_template
	sendgrid_template_version

Unsubscribe Group Resource

	sendgrid_unsubscribe_group

WebHook Resources

	sendgrid_parse_webhook
	sendgrid_event_webhook
*/
package sendgrid

import (
	"context"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	maxStringLength        = 100
	unsubscribeGroupLength = 30
)

// Config holds the provider configuration.
type Config struct {
	APIKey  string
	Host    string
	Subuser string
}

// NewClient creates a new SendGrid client from the config.
// This should be called for each operation to avoid shared state issues.
func (c *Config) NewClient(onBehalfOf string) *sendgrid.Client {
	subuser := c.Subuser
	if onBehalfOf != "" {
		subuser = onBehalfOf
	}
	return sendgrid.NewClient(c.APIKey, c.Host, subuser)
}

// Provider terraform.ResourceProvider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("SENDGRID_API_KEY", nil),
			},
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SENDGRID_HOST", nil),
			},
			"subuser": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SENDGRID_SUBUSER", nil),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"sendgrid_template":              dataSendgridTemplate(),
			"sendgrid_template_version":      dataSendgridTemplateVersion(),
			"sendgrid_unsubscribe_group":     dataSendgridUnsubscribeGroup(),
			"sendgrid_teammate":              dataSendgridTeammate(),
			"sendgrid_domain_authentication": dataSendgridDomainAuthentication(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"sendgrid_api_key":                          resourceSendgridAPIKey(),
			"sendgrid_subuser":                          resourceSendgridSubuser(),
			"sendgrid_template":                         resourceSendgridTemplate(),
			"sendgrid_template_version":                 resourceSendgridTemplateVersion(),
			"sendgrid_unsubscribe_group":                resourceSendgridUnsubscribeGroup(),
			"sendgrid_parse_webhook":                    resourceSendgridParseWebhook(),
			"sendgrid_event_webhook":                    resourceSendgridEventWebhook(),
			"sendgrid_domain_authentication":            resourceSendgridDomainAuthentication(),
			"sendgrid_domain_authentication_validation": resourceSendgridDomainAuthenticationValidation(),
			"sendgrid_domain_authentication_subuser":    resourceSendgridDomainAuthenticationSubuser(),
			"sendgrid_link_branding":                    resourceSendgridLinkBranding(),
			"sendgrid_sso_integration":                  resourceSendgridSSOIntegration(),
			"sendgrid_sso_certificate":                  resourceSendgridSSOCertificate(),
			"sendgrid_teammate":                         resourceSendgridTeammate(),
			"sendgrid_webhook_security_policy":          resourceSendgridWebhookSecurityPolicy(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiKey, ok := d.Get("api_key").(string)
	if !ok || apiKey == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Sendgrid API key wasn't provided",
			Detail: "Unable to retrieve the API key, " +
				"either from the configuration of the provider, " +
				"nor the env variable SENDGRID_API_KEY",
		})

		return nil, diags
	}

	host := d.Get("host").(string)
	subuser := d.Get("subuser").(string)

	config := &Config{
		APIKey:  apiKey,
		Host:    host,
		Subuser: subuser,
	}

	return config, diags
}
