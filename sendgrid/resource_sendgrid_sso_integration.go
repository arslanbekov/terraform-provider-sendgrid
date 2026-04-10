/*
Provide a resource to manage SSO integrations.

**Note** To finalize the integration, a user must click through the 'enable integration'
workflow once after supplying all required fields including an SSO certificate via `aws_sso_certificate`.
Example Usage
```hcl

	resource "sendgrid_sso_integration" "sso" {
		name    = "IdP"
		enabled = false

		signin_url  = "https://idp.com/signin"
		signout_url = "https://idp.com/signout"
		entity_id   = "https://idp.com/12345"
	}

```
Import
A SSO integration can be imported, e.g.
```sh
$ terraform import sendgrid_sso_integration.sso <integration-id>
```
*/
package sendgrid

import (
	"context"
	"net/http"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSendgridSSOIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSendgridSSOIntegrationCreate,
		ReadContext:   resourceSendgridSSOIntegrationRead,
		UpdateContext: resourceSendgridSSOIntegrationUpdate,
		DeleteContext: resourceSendgridSSOIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the integration.",
				Required:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Indicates if the integration is enabled.",
				Required:    true,
			},
			"signin_url": {
				Type: schema.TypeString,
				Description: `The IdP's SAML POST endpoint. This endpoint should receive requests
					and initiate an SSO login flow. This is called the 'Embed Link' in the Twilio SendGrid UI.`,
				Optional: true,
			},
			"signout_url": {
				Type: schema.TypeString,
				Description: `This URL is relevant only for an IdP-initiated authentication flow.
					If a user authenticates from their IdP, this URL will return them to their IdP when logging out.`,
				Optional: true,
			},
			"entity_id": {
				Type: schema.TypeString,
				Description: `An identifier provided by your IdP to identify Twilio SendGrid in the SAML interaction.
					This is called the 'SAML Issuer ID' in the Twilio SendGrid UI.`,
				Optional: true,
			},
			"completed_integration": {
				Type:        schema.TypeBool,
				Description: "Indicates if the integration is complete.",
				Computed:    true,
			},
			"single_signon_url": {
				Type: schema.TypeString,
				Description: `The URL where your IdP should POST its SAML response.
					This is the Twilio SendGrid URL that is responsible for receiving and parsing a SAML assertion.
					This is the same URL as the Audience URL when using SendGrid.`,
				Computed: true,
			},
			"audience_url": {
				Type: schema.TypeString,
				Description: `The URL where your IdP should POST its SAML response.
					This is the Twilio SendGrid URL that is responsible for receiving and parsing a SAML assertion.
					This is the same URL as the Single Sign-On URL when using SendGrid.`,
				Computed: true,
			},
		},
	}
}

func resourceSendgridSSOIntegrationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	name := d.Get("name").(string)
	enabled := d.Get("enabled").(bool)
	signInURL := d.Get("signin_url").(string)
	signOutURL := d.Get("signout_url").(string)
	entityID := d.Get("entity_id").(string)

	apiKeyStruct, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.CreateSSOIntegration(ctx, name, enabled, signInURL, signOutURL, entityID)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	integration := apiKeyStruct.(*sendgrid.SSOIntegration)

	d.SetId(integration.ID)

	return resourceSendgridSSOIntegrationRead(ctx, d, m)
}

func resourceSendgridSSOIntegrationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	integration, requestErr := c.ReadSSOIntegration(ctx, d.Id())

	if requestErr.Err != nil {
		if requestErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(requestErr.Err)
	}

	//nolint:errcheck
	d.Set("name", integration.Name)
	//nolint:errcheck
	d.Set("enabled", integration.Enabled)
	//nolint:errcheck
	d.Set("signin_url", integration.SignInURL)
	//nolint:errcheck
	d.Set("signout_url", integration.SignOutURL)
	//nolint:errcheck
	d.Set("entity_id", integration.EntityID)
	//nolint:errcheck
	d.Set("completed_integration", integration.CompletedIntegration)
	//nolint:errcheck
	d.Set("single_signon_url", integration.SingleSignOnURL)
	//nolint:errcheck
	d.Set("audience_url", integration.AudienceURL)

	return nil
}

func resourceSendgridSSOIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	id := d.Id()
	name := d.Get("name").(string)
	enabled := d.Get("enabled").(bool)
	signInURL := d.Get("signin_url").(string)
	signOutURL := d.Get("signout_url").(string)
	entityID := d.Get("entity_id").(string)

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.UpdateSSOIntegration(ctx, id, name, enabled, signInURL, signOutURL, entityID)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSendgridSSOIntegrationRead(ctx, d, m)
}

func resourceSendgridSSOIntegrationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.DeleteSSOIntegration(ctx, d.Id())
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
