/*
Provide a resource to manage a domain authentication validation.
Example Usage
```hcl

	resource "sendgrid_domain_authentication_validation" "foo" {
		domain_authentication_id = sendgrid_domain_authentication.foo.id
	}

```
*/
package sendgrid

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// https://docs.sendgrid.com/api-reference/domain-authentication/validate-a-domain-authentication
func resourceSendgridDomainAuthenticationValidation() *schema.Resource { //nolint:funlen
	return &schema.Resource{
		CreateContext: resourceSendgridDomainAuthenticationValidationCreate,
		ReadContext:   resourceSendgridDomainAuthenticationValidationRead,
		DeleteContext: resourceSendgridDomainAuthenticationValidationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"domain_authentication_id": {
				Type:        schema.TypeString,
				Description: "Id of the domain authentication to validate.",
				Required:    true,
				ForceNew:    true,
			},
			"sub_user_on_behalf_of": {
				Type:        schema.TypeString,
				Description: "The subuser username for on-behalf-of API calls.",
				Optional:    true,
				ForceNew:    true,
			},

			"valid": {
				Type:        schema.TypeBool,
				Description: "Indicates if this is a valid authenticated domain or not.",
				Computed:    true,
			},
		},
	}
}

func resourceSendgridDomainAuthenticationValidationCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	return validateDomain(ctx, d, m)
}

func validateDomain(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	onBehalfOf := d.Get("sub_user_on_behalf_of").(string)
	if onBehalfOf != "" {
		c.OnBehalfOf = onBehalfOf
	}

	domainID := d.Get("domain_authentication_id").(string)

	validateErr := c.ValidateDomainAuthentication(ctx, domainID)
	if validateErr.Err != nil {
		return diag.FromErr(validateErr.Err)
	}

	d.SetId(domainID)

	return resourceSendgridDomainAuthenticationValidationRead(ctx, d, m)
}

func resourceSendgridDomainAuthenticationValidationRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	onBehalfOf := d.Get("sub_user_on_behalf_of").(string)
	if onBehalfOf != "" {
		c.OnBehalfOf = onBehalfOf
	}

	domainID := d.Id()

	auth, err := c.ReadDomainAuthentication(ctx, domainID)
	if err.Err != nil {
		if err.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(err.Err)
	}

	//nolint:errcheck
	d.Set("domain_authentication_id", fmt.Sprint(auth.ID))
	//nolint:errcheck
	d.Set("valid", auth.Valid)

	return nil
}

func resourceSendgridDomainAuthenticationValidationDelete(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics {
	return nil
}
