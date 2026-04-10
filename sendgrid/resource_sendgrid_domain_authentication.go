/*
Provide a resource to manage a domain authentication.
Example Usage
```hcl

	resource "sendgrid_domain_authentication" "default" {
		domain = "example.com"
	    ips = [ "10.10.10.10" ]
	    is_default = true
	    automatic_security = false
	}

```
Import
An unsubscribe group can be imported, e.g.
```hcl
$ terraform import sendgrid_domain_authentication.default domainId
```
*/
package sendgrid

import (
	"context"
	"fmt"
	"net/http"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSendgridDomainAuthentication() *schema.Resource { //nolint:funlen
	return &schema.Resource{
		CreateContext: resourceSendgridDomainAuthenticationCreate,
		ReadContext:   resourceSendgridDomainAuthenticationRead,
		UpdateContext: resourceSendgridDomainAuthenticationUpdate,
		DeleteContext: resourceSendgridDomainAuthenticationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Description: "Domain being authenticated.",
				Required:    true,
				ForceNew:    true,
			},
			"subdomain": {
				Type:        schema.TypeString,
				Description: "The subdomain to use for this authenticated domain.",
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"username": {
				Type:        schema.TypeString,
				Description: "The username associated with this domain.",
				Computed:    true,
			},
			"ips": {
				Type:        schema.TypeSet,
				Description: "The IP addresses that will be included in the custom SPF record for this.",
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"custom_spf": {
				Type: schema.TypeBool,
				Description: "Specify whether to use a custom SPF or allow SendGrid to manage your SPF. " +
					"This option is only available to authenticated domains set up for manual security.",
				Optional: true,
			},
			"is_default": {
				Type: schema.TypeBool,
				Description: "Whether to use this authenticated domain as the fallback " +
					"if no authenticated domains match the sender's domain.",
				Optional: true,
			},
			"automatic_security": {
				Type:        schema.TypeBool,
				Description: "Whether to allow SendGrid to manage your SPF records, DKIM keys, and DKIM key rotation.",
				Optional:    true,
				ForceNew:    true,
			},
			"custom_dkim_selector": {
				Type:        schema.TypeString,
				Description: "Add a custom DKIM selector. Accepts three letters or numbers.",
				Optional:    true,
				ForceNew:    true,
			},
			"valid": {
				Type:        schema.TypeBool,
				Description: "Indicates if this is a valid authenticated domain or not.",
				Optional:    true,
				Computed:    true,
			},
			"dns": {
				Type:        schema.TypeList,
				Description: "The DNS records used to authenticate the sending domain.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"valid": {
							Type:        schema.TypeBool,
							Description: "Indicates if this is a valid CNAME.",
							Computed:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "The type of DNS record.",
							Computed:    true,
						},
						"host": {
							Type:        schema.TypeString,
							Description: "The domain that this CNAME is created for.",
							Computed:    true,
						},
						"data": {
							Type:        schema.TypeString,
							Description: "The actual DNS record.",
							Computed:    true,
						},
					},
				},
			},
			"sub_user_on_behalf_of": {
				Type:        schema.TypeString,
				Description: "The subuser's username. The API call is made on behalf of the subuser account.",
				Optional:    true,
			},
		},
	}
}

func resourceSendgridDomainAuthenticationCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	domain := d.Get("domain").(string)
	subdomain := d.Get("subdomain").(string)
	customSPF := d.Get("custom_spf").(bool)
	isDefault := d.Get("is_default").(bool)
	automaticSecurity := d.Get("automatic_security").(bool)
	customDKIMSelector := d.Get("custom_dkim_selector").(string)
	ipsSet := d.Get("ips").(*schema.Set).List()
	ips := make([]string, 0)

	onBehalfOf := d.Get("sub_user_on_behalf_of").(string)
	if len(onBehalfOf) > 0 {
		c.OnBehalfOf = onBehalfOf
	}

	for _, ip := range ipsSet {
		ips = append(ips, ip.(string))
	}

	apiKeyStruct, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.CreateDomainAuthentication(
			ctx,
			domain,
			subdomain,
			ips,
			customSPF,
			isDefault,
			automaticSecurity,
			customDKIMSelector,
		)
	})

	auth := apiKeyStruct.(*sendgrid.DomainAuthentication)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprint(auth.ID))

	return resourceSendgridDomainAuthenticationRead(ctx, d, m)
}

func resourceSendgridDomainAuthenticationRead( //nolint:funlen,cyclop
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	auth, err := c.ReadDomainAuthentication(ctx, d.Id())
	if err.Err != nil {
		if err.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err.Err)
	}

	//nolint:errcheck
	d.Set("domain", auth.Domain)
	//nolint:errcheck
	d.Set("subdomain", auth.Subdomain)
	//nolint:errcheck
	d.Set("username", auth.Username)
	//nolint:errcheck
	d.Set("is_default", auth.IsDefault)
	//nolint:errcheck
	d.Set("custom_spf", auth.CustomSPF)
	//nolint:errcheck
	d.Set("custom_dkim_selector", auth.CustomDKIMSelector)
	//nolint:errcheck
	d.Set("valid", auth.Valid)

	ips := make([]interface{}, len(auth.IPs))
	for idx, ip := range auth.IPs {
		ips[idx] = ip
	}

	onBehalfOf := d.Get("sub_user_on_behalf_of").(string)
	if len(onBehalfOf) > 0 {
		c.OnBehalfOf = onBehalfOf
	}

	if er := d.Set("ips", schema.NewSet(d.Get("ips").(*schema.Set).F, ips)); er != nil {
		return diag.FromErr(er)
	}

	dns := make([]interface{}, 0)
	if auth.DNS.DKIM1.Type != "" {
		dns = append(dns, map[string]interface{}{
			"type":  auth.DNS.DKIM1.Type,
			"valid": auth.DNS.DKIM1.Valid,
			"host":  auth.DNS.DKIM1.Host,
			"data":  auth.DNS.DKIM1.Data,
		})
	}

	if auth.DNS.DKIM2.Type != "" {
		dns = append(dns, map[string]interface{}{
			"type":  auth.DNS.DKIM2.Type,
			"valid": auth.DNS.DKIM2.Valid,
			"host":  auth.DNS.DKIM2.Host,
			"data":  auth.DNS.DKIM2.Data,
		})
	}

	if auth.DNS.MailCNAME.Type != "" {
		dns = append(dns, map[string]interface{}{
			"type":  auth.DNS.MailCNAME.Type,
			"valid": auth.DNS.MailCNAME.Valid,
			"host":  auth.DNS.MailCNAME.Host,
			"data":  auth.DNS.MailCNAME.Data,
		})
	}

	if auth.DNS.DKIM.Type != "" {
		dns = append(dns, map[string]interface{}{
			"type":  auth.DNS.DKIM.Type,
			"valid": auth.DNS.DKIM.Valid,
			"host":  auth.DNS.DKIM.Host,
			"data":  auth.DNS.DKIM.Data,
		})
	}

	if auth.DNS.SubDomainSPF.Type != "" {
		dns = append(dns, map[string]interface{}{
			"type":  auth.DNS.SubDomainSPF.Type,
			"valid": auth.DNS.SubDomainSPF.Valid,
			"host":  auth.DNS.SubDomainSPF.Host,
			"data":  auth.DNS.SubDomainSPF.Data,
		})
	}

	if auth.DNS.MailServer.Type != "" {
		dns = append(dns, map[string]interface{}{
			"type":  auth.DNS.MailServer.Type,
			"valid": auth.DNS.MailServer.Valid,
			"host":  auth.DNS.MailServer.Host,
			"data":  auth.DNS.MailServer.Data,
		})
	}

	if er := d.Set("dns", dns); er != nil {
		return diag.FromErr(er)
	}

	return nil
}

func resourceSendgridDomainAuthenticationUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	isDefault := d.Get("is_default").(bool)
	customSPF := d.Get("custom_spf").(bool)

	onBehalfOf := d.Get("sub_user_on_behalf_of").(string)
	if len(onBehalfOf) > 0 {
		c.OnBehalfOf = onBehalfOf
	}

	auth, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.UpdateDomainAuthentication(ctx, d.Id(), isDefault, customSPF)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if !auth.(*sendgrid.DomainAuthentication).Valid && d.Get("valid").(bool) {
		if err := c.ValidateDomainAuthentication(ctx, d.Id()); err.Err != nil || err.StatusCode != 200 {
			if err.Err != nil {
				return diag.FromErr(err.Err)
			}

			return diag.Errorf("unable to validate domain DNS configuration")
		}
	}

	return resourceSendgridDomainAuthenticationRead(ctx, d, m)
}

func resourceSendgridDomainAuthenticationDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	onBehalfOf := d.Get("sub_user_on_behalf_of").(string)
	if len(onBehalfOf) > 0 {
		c.OnBehalfOf = onBehalfOf
	}
	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.DeleteDomainAuthentication(ctx, d.Id())
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
