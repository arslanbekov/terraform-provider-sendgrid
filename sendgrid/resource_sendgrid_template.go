/*
Provide a resource to manage a template of email.
Example Usage
```hcl

	resource "sendgrid_template" "template" {
		name       = "my-template"
		generation = "dynamic"
	}

```
Import
A template can be imported, e.g.
```hcl
$ terraform import sendgrid_template.template templateID
```
*/
package sendgrid

import (
	"context"
	"net/http"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceSendgridTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSendgridTemplateCreate,
		ReadContext:   resourceSendgridTemplateRead,
		UpdateContext: resourceSendgridTemplateUpdate,
		DeleteContext: resourceSendgridTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the template, max length: 100.",
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, maxStringLength),
			},
			"generation": {
				Type:        schema.TypeString,
				Description: "Defines the generation of the template, allowed values: legacy, dynamic (default).",
				Optional:    true,
				Default:     "dynamic",
				ForceNew:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "The date and time of the last update of this template.",
				Computed:    true,
			},
		},
	}
}

func resourceSendgridTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	name := d.Get("name").(string)
	generation := d.Get("generation").(string)

	templateStruct, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.CreateTemplate(ctx, name, generation)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	template := templateStruct.(*sendgrid.Template)
	//nolint:errcheck
	d.Set("updated_at", template.UpdatedAt)
	d.SetId(template.ID)

	return nil
}

func resourceSendgridTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	template, readErr := c.ReadTemplate(ctx, d.Id())
	if readErr.Err != nil {
		if readErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(readErr.Err)
	}

	if err := sendgridTemplateParse(template, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func sendgridTemplateParse(template *sendgrid.Template, d *schema.ResourceData) error {
	if err := d.Set("name", template.Name); err != nil {
		return ErrSetTemplateName
	}

	if err := d.Set("generation", template.Generation); err != nil {
		return ErrSetTemplateGeneration
	}

	if err := d.Set("updated_at", template.UpdatedAt); err != nil {
		return ErrSetTemplateUpdatedAt
	}

	return nil
}

func resourceSendgridTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	if d.HasChange("name") {
		_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
			return c.UpdateTemplate(ctx, d.Id(), d.Get("name").(string))
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceSendgridTemplateRead(ctx, d, m)
}

func resourceSendgridTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.DeleteTemplate(ctx, d.Id())
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
