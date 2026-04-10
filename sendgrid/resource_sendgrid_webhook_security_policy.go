package sendgrid

import (
	"context"
	"net/http"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSendgridWebhookSecurityPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a SendGrid webhook security policy. Security policies provide authentication mechanisms for webhooks, including OAuth and signature verification.",
		CreateContext: resourceSendgridWebhookSecurityPolicyCreate,
		ReadContext:   resourceSendgridWebhookSecurityPolicyRead,
		UpdateContext: resourceSendgridWebhookSecurityPolicyUpdate,
		DeleteContext: resourceSendgridWebhookSecurityPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the webhook security policy.",
			},
			"oauth": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				ForceNew:    true,
				Description: "OAuth configuration for webhook authentication. Can be used together with signature.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The OAuth client ID.",
						},
						"client_secret": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "The OAuth client secret. This value is sensitive and will not be displayed in logs.",
						},
						"token_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The OAuth token URL.",
						},
						"scopes": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Required:    true,
							Description: "List of OAuth scopes required for webhook access.",
						},
					},
				},
				AtLeastOneOf: []string{"oauth", "signature"},
			},
			"signature": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				ForceNew:    true,
				Description: "Signature configuration for webhook authentication. Can be used together with oauth.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Whether signature verification is enabled for this webhook security policy.",
						},
						"public_key": {
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
							Description: "The public key used for signature verification. This is computed by SendGrid when signature is enabled.",
						},
					},
				},
				AtLeastOneOf: []string{"oauth", "signature"},
			},
		},
	}
}

func resourceSendgridWebhookSecurityPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	name := d.Get("name").(string)

	var oauth map[string]interface{}
	if v, ok := d.GetOk("oauth"); ok {
		oauthList := v.([]interface{})
		if len(oauthList) > 0 && oauthList[0] != nil {
			oauth = oauthList[0].(map[string]interface{})
		}
	}

	var signature map[string]interface{}
	if v, ok := d.GetOk("signature"); ok {
		sigList := v.([]interface{})
		if len(sigList) > 0 && sigList[0] != nil {
			signature = sigList[0].(map[string]interface{})
		}
	}

	parseWebhookStruct, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.CreateWebhookSecurityPolicy(ctx, name, oauth, signature)
	})

	securityPolicy := parseWebhookStruct.(*sendgrid.WebhookSecurityPolicyResponse)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(securityPolicy.Policy.ID)

	return resourceSendgridWebhookSecurityPolicyRead(ctx, d, m)
}

func resourceSendgridWebhookSecurityPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	webhook, err := c.ReadWebhookSecurityPolicy(ctx, d.Id())
	if err.Err != nil {
		if err.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err.Err)
	}

	//nolint:errcheck
	d.Set("name", webhook.Policy.Name)

	if webhook.Policy.OAuth != nil {
		var existingOAuth map[string]interface{}
		if v, ok := d.GetOk("oauth"); ok {
			oauthList := v.([]interface{})
			if len(oauthList) > 0 && oauthList[0] != nil {
				existingOAuth = oauthList[0].(map[string]interface{})
			}
		}

		oauthMap := map[string]interface{}{
			"client_id": webhook.Policy.OAuth.ClientID,
			"token_url": webhook.Policy.OAuth.TokenURL,
			"scopes":    webhook.Policy.OAuth.Scopes,
		}

		if existingOAuth != nil {
			if clientSecret, exists := existingOAuth["client_secret"]; exists {
				oauthMap["client_secret"] = clientSecret
			}
		}

		//nolint:errcheck
		d.Set("oauth", []interface{}{oauthMap})
	} else {
		//nolint:errcheck
		d.Set("oauth", []interface{}{})
	}

	if webhook.Policy.Signature != nil {
		signatureMap := map[string]interface{}{
			"enabled": webhook.Policy.Signature.PublicKey != "",
		}

		if webhook.Policy.Signature.PublicKey != "" {
			signatureMap["public_key"] = webhook.Policy.Signature.PublicKey
		}

		//nolint:errcheck
		d.Set("signature", []interface{}{signatureMap})
	} else {
		if v, ok := d.GetOk("signature"); ok {
			//nolint:errcheck
			d.Set("signature", v)
		} else {
			//nolint:errcheck
			d.Set("signature", []interface{}{})
		}
	}

	return nil
}

func resourceSendgridWebhookSecurityPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	name := d.Get("name").(string)

	var oauth map[string]interface{}
	if v, ok := d.GetOk("oauth"); ok {
		oauthList := v.([]interface{})
		if len(oauthList) > 0 && oauthList[0] != nil {
			oauth = oauthList[0].(map[string]interface{})
		}
	}

	var signature map[string]interface{}
	if v, ok := d.GetOk("signature"); ok {
		sigList := v.([]interface{})
		if len(sigList) > 0 && sigList[0] != nil {
			signature = sigList[0].(map[string]interface{})
		}
	}

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return nil, c.UpdateWebhookSecurityPolicy(ctx, d.Id(), name, oauth, signature)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSendgridWebhookSecurityPolicyRead(ctx, d, m)
}

func resourceSendgridWebhookSecurityPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.DeleteWebhookSecurityPolicy(ctx, d.Id())
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
