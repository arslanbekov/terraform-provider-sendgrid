/*
Provide a resource to manage a subuser.
Example Usage
```hcl

	resource "sendgrid_subuser" "subuser" {
		username = "my-subuser"
		email    = "subuser@example.org"
		password = "Passw0rd!"
		ips      = [
			"127.0.0.1"
		]
	}

```
Import
A subuser can be imported, e.g.
```hcl
$ terraform import sendgrid_subuser.subuser userName
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

func resourceSendgridSubuser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSendgridSubuserCreate,
		ReadContext:   resourceSendgridSubuserRead,
		UpdateContext: resourceSendgridSubuserUpdate,
		DeleteContext: resourceSendgridSubuserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Description: "The name of the subuser.",
				Required:    true,
			},
			"password": {
				Type:        schema.TypeString,
				Description: "The password the subuser will use when logging into SendGrid.",
				Sensitive:   true,
				Required:    true,
			},
			"email": {
				Type:        schema.TypeString,
				Description: "The email of the subuser.",
				Required:    true,
			},
			"ips": {
				Type:        schema.TypeSet,
				Description: "The IP addresses that should be assigned to this subuser.",
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"user_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"signup_session_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorization_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"credit_allocation_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSendgridSubuserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	username := d.Get("username").(string)
	password := d.Get("password").(string)
	email := d.Get("email").(string)

	ipsSet := d.Get("ips").(*schema.Set).List()
	ips := make([]string, 0)

	for _, ip := range ipsSet {
		ips = append(ips, ip.(string))
	}

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.CreateSubuser(ctx, username, email, password, ips)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(username)

	if d.Get("disabled").(bool) {
		return resourceSendgridSubuserUpdate(ctx, d, m)
	}

	return resourceSendgridSubuserRead(ctx, d, m)
}

func resourceSendgridSubuserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	// hack to clear any on behalf of set in create sub user
	// to fix this properly I think we need to pass this down rather than setting global state
	c.OnBehalfOf = ""

	subUser, requestErr := c.ReadSubUser(ctx, d.Id())
	if requestErr.Err != nil {
		if requestErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(requestErr.Err)
	}

	if len(subUser) == 0 {
		return diag.FromErr(subUserNotFound(d.Id()))
	}

	//nolint:errcheck
	d.Set("user_id", subUser[0].ID)
	//nolint:errcheck
	d.Set("disabled", subUser[0].Disabled)
	//nolint:errcheck
	d.Set("email", subUser[0].Email)

	return nil
}

func resourceSendgridSubuserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	if d.HasChange("disabled") {
		if _, requestErr := c.UpdateSubuser(ctx, d.Id(), d.Get("disabled").(bool)); requestErr.Err != nil {
			return diag.FromErr(requestErr.Err)
		}
	}

	if d.HasChange("ips") {
		ipsSet := d.Get("ips").(*schema.Set).List()
		ips := make([]string, 0)

		for _, ip := range ipsSet {
			ips = append(ips, ip.(string))
		}

		if requestErr := c.UpdateSubuserIPs(ctx, d.Id(), ips); requestErr.Err != nil {
			return diag.FromErr(requestErr.Err)
		}
	}

	if d.HasChange("password") {
		oldPassword, newPassword := d.GetChange("password")
		username := d.Get("username").(string)

		if requestErr := c.UpdateSubuserPassword(
			ctx,
			username,
			oldPassword.(string),
			newPassword.(string)); requestErr.Err != nil {
			return diag.FromErr(requestErr.Err)
		}
	}

	return resourceSendgridSubuserRead(ctx, d, m)
}

func resourceSendgridSubuserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.DeleteSubuser(ctx, d.Id())
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
