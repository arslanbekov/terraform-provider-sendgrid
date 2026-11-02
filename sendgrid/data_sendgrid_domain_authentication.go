package sendgrid

import (
	"context"
	"fmt"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSendgridDomainAuthentication() *schema.Resource {
	s := resourceSendgridDomainAuthentication().Schema

	for _, val := range s {
		val.Computed = true
		val.Optional = false
		val.Required = false
		val.Default = nil
		val.ValidateFunc = nil
		val.ForceNew = true
	}

	s["domain"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The domain of the authenticated domain to retrieve",
	}
	s["subdomain"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		Description: "The subdomain of the authenticated domain to retrieve. " +
			"Required when several authenticated domains share the same domain",
	}

	return &schema.Resource{
		ReadContext: dataSendgridDomainAuthenticationRead,
		Schema:      s,
	}
}

func dataSendgridDomainAuthenticationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	domain := d.Get("domain").(string)
	subdomain := d.Get("subdomain").(string)
	config := m.(*Config)
	c := config.NewClient("")

	// exclude_subusers: only the parent account's own domains can be assigned to subusers.
	authsStruct, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.ReadDomainAuthentications(ctx, domain, true)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// The API's domain filter is a search, so match exactly here.
	matches := make([]sendgrid.DomainAuthentication, 0)

	for _, auth := range authsStruct.([]sendgrid.DomainAuthentication) {
		if auth.Domain == domain && (subdomain == "" || auth.Subdomain == subdomain) {
			matches = append(matches, auth)
		}
	}

	switch len(matches) {
	case 0:
		return diag.Errorf("unable to find an authenticated domain with domain '%s'", domain)
	case 1:
		d.SetId(fmt.Sprint(matches[0].ID))

		// The resource reader leaves automatic_security to the configuration; a data source has none.
		//nolint:errcheck
		d.Set("automatic_security", matches[0].AutomaticSecurity)

		return resourceSendgridDomainAuthenticationRead(ctx, d, m)
	default:
		return diag.Errorf("found %d authenticated domains with domain '%s', set 'subdomain' to select one",
			len(matches), domain)
	}
}
