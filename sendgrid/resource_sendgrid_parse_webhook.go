/*
Provide a resource to manage parse webhook settings.
Example Usage
```hcl

	resource "sendgrid_parse_webhook" "default" {
		hostname = "parse.foo.bar"
	    url = "https://foo.bar/sendgrid/inbound"
	    spam_check = false
	    send_raw = false
	}

```
Import
An unsubscribe webhook can be imported, e.g.
```hcl
$ terraform import sendgrid_parse_webhook.default hostname
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

func resourceSendgridParseWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSendgridParseWebhookCreate,
		ReadContext:   resourceSendgridParseWebhookRead,
		UpdateContext: resourceSendgridParseWebhookUpdate,
		DeleteContext: resourceSendgridParseWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			if diff.HasChange("webhook_security_policy_id") {
				old, new := diff.GetChange("webhook_security_policy_id")
				if old.(string) != "" && new.(string) == "" {
					return nil
				}
			}
			return nil
		},

		Schema: map[string]*schema.Schema{
			"hostname": {
				Type: schema.TypeString,
				Description: "A specific and unique domain or subdomain " +
					"that you have created to use exclusively to parse your incoming email. " +
					"For example, parse.yourdomain.com.",
				Required: true,
				ForceNew: true,
			},
			"url": {
				Type: schema.TypeString,
				Description: "The public URL where you would like SendGrid to POST the data parsed from your email. " +
					"Any emails sent with the given hostname provided (whose MX records have been updated to point to SendGrid) " +
					"will be parsed and POSTed to this URL.",
				Required: true,
				ForceNew: true,
			},
			"spam_check": {
				Type: schema.TypeBool,
				Description: "Indicates if you would like SendGrid to check the content parsed from your emails for spam " +
					"before POSTing them to your domain.",
				Optional: true,
			},
			"send_raw": {
				Type: schema.TypeBool,
				Description: "Indicates if you would like SendGrid to post the original MIME-type content of your parsed email. " +
					"When this parameter is set to \"true\", SendGrid will send a JSON payload of the content of your email.",
				Optional: true,
			},
			"webhook_security_policy_id": {
				Type: schema.TypeString,
				Description: "The ID of the webhook security policy to apply to this parse webhook. " +
					"See the `sendgrid_webhook_security_policy` resource for more details.",
				Optional: true,
			},
		},
	}
}

func resourceSendgridParseWebhookCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	hostname := d.Get("hostname").(string)
	url := d.Get("url").(string)
	spamCheck := d.Get("spam_check").(bool)
	sendRaw := d.Get("send_raw").(bool)
	securityPolicy := d.Get("webhook_security_policy_id").(string)

	parseWebhookStruct, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.CreateParseWebhook(ctx, hostname, url, spamCheck, sendRaw, securityPolicy)
	})

	webhook := parseWebhookStruct.(*sendgrid.ParseWebhook)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(webhook.Hostname)

	return resourceSendgridParseWebhookRead(ctx, d, m)
}

func resourceSendgridParseWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	webhook, err := c.ReadParseWebhook(ctx, d.Id())
	if err.Err != nil {
		if err.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(err.Err)
	}

	//nolint:errcheck
	d.Set("hostname", webhook.Hostname)
	//nolint:errcheck
	d.Set("url", webhook.URL)
	//nolint:errcheck
	d.Set("spam_check", webhook.SpamCheck)
	//nolint:errcheck
	d.Set("send_raw", webhook.SendRaw)
	//nolint:errcheck
	d.Set("webhook_security_policy_id", webhook.SecurityPolicy)

	return nil
}

func resourceSendgridParseWebhookUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	spamCheck := d.Get("spam_check").(bool)
	sendRaw := d.Get("send_raw").(bool)
	securityPolicy := d.Get("webhook_security_policy_id").(string)

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return nil, c.UpdateParseWebhook(ctx, d.Id(), spamCheck, sendRaw, securityPolicy)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSendgridParseWebhookRead(ctx, d, m)
}

func resourceSendgridParseWebhookDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	if securityPolicyId := d.Get("webhook_security_policy_id").(string); securityPolicyId != "" {
		spamCheck := d.Get("spam_check").(bool)
		sendRaw := d.Get("send_raw").(bool)

		_, updateErr := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
			return nil, c.UpdateParseWebhook(ctx, d.Id(), spamCheck, sendRaw, "")
		})

		if updateErr != nil {
			return diag.FromErr(updateErr)
		}
	}

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.DeleteParseWebhook(ctx, d.Id())
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
