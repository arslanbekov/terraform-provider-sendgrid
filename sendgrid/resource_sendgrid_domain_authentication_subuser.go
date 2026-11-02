/*
Provide a resource to associate a parent account's authenticated domain with a subuser.
Example Usage
```hcl

	resource "sendgrid_domain_authentication_subuser" "team" {
		domain_id = sendgrid_domain_authentication.main.id
		username  = sendgrid_subuser.team.username
	}

```
Import
A domain authentication subuser association can be imported by `domainId/username` or by `username`, e.g.
```hcl
$ terraform import sendgrid_domain_authentication_subuser.team 12345/team-a
```
*/
package sendgrid

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// https://www.twilio.com/docs/sendgrid/api-reference/domain-authentication/associate-an-authenticated-domain-with-a-subuser
func resourceSendgridDomainAuthenticationSubuser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSendgridDomainAuthenticationSubuserCreate,
		ReadContext:   resourceSendgridDomainAuthenticationSubuserRead,
		DeleteContext: resourceSendgridDomainAuthenticationSubuserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:        schema.TypeString,
				Description: "ID of the parent account's authenticated domain to assign to the subuser.",
				Required:    true,
				ForceNew:    true,
			},
			"username": {
				Type:        schema.TypeString,
				Description: "Username of the subuser the domain is assigned to.",
				Required:    true,
				ForceNew:    true,
			},
			"domain": {
				Type:        schema.TypeString,
				Description: "Domain of the assigned authenticated domain.",
				Computed:    true,
			},
			"subdomain": {
				Type:        schema.TypeString,
				Description: "Subdomain of the assigned authenticated domain.",
				Computed:    true,
			},
			"valid": {
				Type:        schema.TypeBool,
				Description: "Indicates if the assigned authenticated domain is valid.",
				Computed:    true,
			},
		},
	}
}

func resourceSendgridDomainAuthenticationSubuserCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	domainID := d.Get("domain_id").(string)
	username := d.Get("username").(string)

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.AssociateDomainAuthenticationSubuser(ctx, domainID, username)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(domainID + "/" + username)

	return resourceSendgridDomainAuthenticationSubuserRead(ctx, d, m)
}

func resourceSendgridDomainAuthenticationSubuserRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	// The ID is "domainID/username"; an import may pass the bare username.
	username := d.Id()
	if idx := strings.LastIndex(username, "/"); idx >= 0 {
		username = username[idx+1:]
	}

	auth, err := c.ReadDomainAuthenticationSubuser(ctx, username)
	if err.Err != nil {
		if err.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(err.Err)
	}

	if auth.ID == 0 {
		d.SetId("")
		return nil
	}

	domainID := fmt.Sprint(auth.ID)
	if !strings.Contains(d.Id(), "/") {
		d.SetId(domainID + "/" + username)
	}

	// A domain_id that differs from the configuration shows up in the plan as a
	// replacement, which disassociates the current domain and assigns the wanted one.
	//nolint:errcheck
	d.Set("domain_id", domainID)
	//nolint:errcheck
	d.Set("username", username)
	//nolint:errcheck
	d.Set("domain", auth.Domain)
	//nolint:errcheck
	d.Set("subdomain", auth.Subdomain)
	//nolint:errcheck
	d.Set("valid", auth.Valid)

	return nil
}

func resourceSendgridDomainAuthenticationSubuserDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	// domain_id comes from state: after a drifted Read it is the domain actually assigned.
	domainID := d.Get("domain_id").(string)
	username := d.Get("username").(string)

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.DisassociateDomainAuthenticationSubuser(ctx, domainID, username)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
