/*
Provide a resource to manage link branding.
Example Usage
```hcl

	resource "sendgrid_link_branding" "default" {
		domain = "example.com"
	    is_default = true
	}

```
Import
An unsubscribe group can be imported, e.g.
```hcl
$ terraform import sendgrid_link_branding.default linkId
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

func resourceSendgridLinkBranding() *schema.Resource { //nolint:funlen
	return &schema.Resource{
		CreateContext: resourceSendgridLinkBrandingCreate,
		ReadContext:   resourceSendgridLinkBrandingRead,
		UpdateContext: resourceSendgridLinkBrandingUpdate,
		DeleteContext: resourceSendgridLinkBrandingDelete,
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
				Description: "The subdomain to use for this link branding.",
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"username": {
				Type:        schema.TypeString,
				Description: "The username associated with this domain.",
				Computed:    true,
			},
			"is_default": {
				Type:        schema.TypeBool,
				Description: "Indicates if this is the default link branding.",
				Optional:    true,
			},
			"valid": {
				Type: schema.TypeBool,
				Description: "Indicates if this is a valid link branding or not. " +
					"Set to `true` to attempt validation on first update.",
				Optional: true,
				Computed: true,
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
		},
	}
}

func resourceSendgridLinkBrandingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	domain := d.Get("domain").(string)
	subdomain := d.Get("subdomain").(string)
	isDefault := d.Get("is_default").(bool)

	apiKeyStruct, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.CreateLinkBranding(ctx, domain, subdomain, isDefault)
	})

	link := apiKeyStruct.(*sendgrid.LinkBranding)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprint(link.ID))

	return resourceSendgridLinkBrandingRead(ctx, d, m)
}

func resourceSendgridLinkBrandingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	link, err := c.ReadLinkBranding(ctx, d.Id())
	if err.Err != nil {
		if err.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err.Err)
	}

	//nolint:errcheck
	d.Set("domain", link.Domain)
	//nolint:errcheck
	d.Set("subdomain", link.Subdomain)
	//nolint:errcheck
	d.Set("username", link.Username)
	//nolint:errcheck
	d.Set("is_default", link.IsDefault)
	//nolint:errcheck
	d.Set("valid", link.Valid)

	dns := make([]interface{}, 0)
	if link.DNS.DomainCNAME.Type != "" {
		dns = append(dns, map[string]interface{}{
			"type":  link.DNS.DomainCNAME.Type,
			"valid": link.DNS.DomainCNAME.Valid,
			"host":  link.DNS.DomainCNAME.Host,
			"data":  link.DNS.DomainCNAME.Data,
		})
	}

	if link.DNS.OwnerCNAME.Type != "" {
		dns = append(dns, map[string]interface{}{
			"type":  link.DNS.OwnerCNAME.Type,
			"valid": link.DNS.OwnerCNAME.Valid,
			"host":  link.DNS.OwnerCNAME.Host,
			"data":  link.DNS.OwnerCNAME.Data,
		})
	}

	if er := d.Set("dns", dns); er != nil {
		return diag.FromErr(er)
	}

	return nil
}

func resourceSendgridLinkBrandingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	isDefault := d.Get("is_default").(bool)

	link, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.UpdateLinkBranding(ctx, d.Id(), isDefault)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if !link.(*sendgrid.LinkBranding).Valid && d.Get("valid").(bool) {
		if err := c.ValidateLinkBranding(ctx, d.Id()); err.Err != nil || err.StatusCode != 200 {
			if err.Err != nil {
				return diag.FromErr(err.Err)
			}

			return diag.Errorf("unable to validate link branding DNS configuration")
		}
	}

	return resourceSendgridLinkBrandingRead(ctx, d, m)
}

func resourceSendgridLinkBrandingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.DeleteLinkBranding(ctx, d.Id())
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
