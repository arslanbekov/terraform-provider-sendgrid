/*
Provide a resource to manage a version of template.
Example Usage
```hcl

	resource "sendgrid_template" "template" {
		name       = "my-template"
		generation = "dynamic"
	}

	resource "sendgrid_template_version" "template_version" {
		name                   = "my-template-version"
		template_id            = sendgrid_template.template.id
		active                 = 1
		html_content           = "<%body%>"
		generate_plain_content = true
		subject                = "subject"
	}

```
Import
A template version can be imported, e.g.
```hcl
$ terraform import sendgrid_template_version.template_version templateID/templateVersionId
```
*/
package sendgrid

import (
	"context"
	"net/http"
	"reflect"
	"strings"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ImportSplitParts is the expected length of
// the splitted import string for template versions.
const ImportSplitParts = 2

func resourceSendgridTemplateVersion() *schema.Resource { //nolint:funlen
	return &schema.Resource{
		CreateContext: resourceSendgridTemplateVersionCreate,
		ReadContext:   resourceSendgridTemplateVersionRead,
		UpdateContext: resourceSendgridTemplateVersionUpdate,
		DeleteContext: resourceSendgridTemplateVersionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSendgridTemplateVersionImport,
		},

		Schema: map[string]*schema.Schema{
			"template_id": {
				Type:        schema.TypeString,
				Description: "ID of the transactional template.",
				Required:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "The date and time that this transactional template version was updated.",
				Computed:    true,
			},
			"thumbnail_url": {
				Type:        schema.TypeString,
				Description: "A thumbnail preview of the template's html content.",
				Computed:    true,
			},
			"active": {
				Type: schema.TypeInt,
				Description: "Set the version as the active version associated with the template. " +
					"Only one version of a template can be active. " +
					"The first version created for a template will automatically be set to Active. Allowed values: 0, 1.",
				Optional: true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the transactional template version, max length: 100.",
				Required:    true,
			},
			"html_content": {
				Type:        schema.TypeString,
				Description: "The HTML content of the version, maximum of 1048576 bytes allowed.",
				Optional:    true,
			},
			"plain_content": {
				Type:        schema.TypeString,
				Description: "Text/plain content of the transactional template version, maximum of 1048576 bytes allowed.",
				Computed:    true,
				Optional:    true,
			},
			"generate_plain_content": {
				Type: schema.TypeBool,
				Description: "If true (default), plain_content is always generated from html_content. " +
					"If false, plain_content is not altered.",
				Optional: true,
				Default:  true,
			},
			"subject": {
				Type:        schema.TypeString,
				Description: "Subject of the new transactional template version, max length: 255.",
				Required:    true,
			},
			"editor": {
				Type:         schema.TypeString,
				Description:  "The editor used in the UI, allowed values: code (default), design.",
				Optional:     true,
				Default:      "code",
				ValidateFunc: validation.StringInSlice([]string{"code", "design"}, false),
			},
			"test_data": {
				Type: schema.TypeString,
				Description: "For dynamic templates only, " +
					"the mock json data that will be used for template preview and test sends.",
				Optional: true,
			},
		},
	}
}

func resourceSendgridTemplateVersionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	templateVersionStruct, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.CreateTemplateVersion(ctx, sendgrid.TemplateVersion{
			TemplateID:           d.Get("template_id").(string),
			Active:               d.Get("active").(int),
			Name:                 d.Get("name").(string),
			HTMLContent:          d.Get("html_content").(string),
			PlainContent:         d.Get("plain_content").(string),
			GeneratePlainContent: d.Get("generate_plain_content").(bool),
			Subject:              d.Get("subject").(string),
			Editor:               d.Get("editor").(string),
			TestData:             d.Get("test_data").(string),
		})
	})
	if err != nil {
		return diag.FromErr(err)
	}

	templateVersion := templateVersionStruct.(*sendgrid.TemplateVersion)
	//nolint:errcheck
	d.Set("updated_at", templateVersion.UpdatedAt)
	d.SetId(templateVersion.ID)

	return nil
}

func resourceSendgridTemplateVersionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	templateVersion, readErr := c.ReadTemplateVersion(ctx, d.Get("template_id").(string), d.Id())
	if readErr.Err != nil {
		if readErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(readErr.Err)
	}

	if er := parseTemplateVersion(d, templateVersion); er != nil {
		return diag.FromErr(er)
	}

	return nil
}

func parseTemplateVersion(d *schema.ResourceData, templateVersion *sendgrid.TemplateVersion) error {
	if err := d.Set("updated_at", templateVersion.UpdatedAt); err != nil {
		return ErrSetTemplateVersionUpdatedAt
	}

	if err := d.Set("thumbnail_url", templateVersion.ThumbnailURL); err != nil {
		return ErrSetTemplateVersionThumbnailURL
	}

	if err := d.Set("active", templateVersion.Active); err != nil {
		return ErrSetTemplateVersionActive
	}

	if err := d.Set("name", templateVersion.Name); err != nil {
		return ErrSetTemplateVersionName
	}

	if err := d.Set("html_content", templateVersion.HTMLContent); err != nil {
		return ErrSetTemplateVersionHTMLContent
	}

	if err := d.Set("plain_content", templateVersion.PlainContent); err != nil {
		return ErrSetTemplateVersionPlainContent
	}

	if err := d.Set("generate_plain_content", templateVersion.GeneratePlainContent); err != nil {
		return ErrSetTemplateVersionGenPlainContent
	}

	if err := d.Set("subject", templateVersion.Subject); err != nil {
		return ErrSetTemplateVersionSubject
	}

	if err := d.Set("editor", templateVersion.Editor); err != nil {
		return ErrSetTemplateVersionEditor
	}

	return nil
}

func resourceSendgridTemplateVersionUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	baseTemplateVersion := sendgrid.TemplateVersion{
		ID:         d.Id(),
		TemplateID: d.Get("template_id").(string),
	}
	templateVersion := baseTemplateVersion

	if d.HasChange("active") {
		templateVersion.Active = d.Get("active").(int)
	}

	if templateVersion.Active == 1 {
		_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
			return c.ActivateTemplateVersion(ctx, templateVersion)
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("name") {
		templateVersion.Name = d.Get("name").(string)
	}

	if d.HasChange("html_content") {
		templateVersion.HTMLContent = d.Get("html_content").(string)
	}

	if d.HasChange("plain_content") {
		templateVersion.PlainContent = d.Get("plain_content").(string)
	}

	if d.HasChange("generate_plain_content") {
		templateVersion.GeneratePlainContent = d.Get("generate_plain_content").(bool)
	}

	if d.HasChange("subject") {
		templateVersion.Subject = d.Get("subject").(string)
	}

	if d.HasChange("editor") {
		templateVersion.Editor = d.Get("editor").(string)
	}

	if d.HasChange("test_data") {
		templateVersion.TestData = d.Get("test_data").(string)
	}

	if reflect.DeepEqual(baseTemplateVersion, templateVersion) {
		return nil
	}

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.UpdateTemplateVersion(ctx, templateVersion)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSendgridTemplateVersionRead(ctx, d, m)
}

func resourceSendgridTemplateVersionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	config := m.(*Config)
	c := config.NewClient("")

	_, err := sendgrid.RetryOnRateLimit(ctx, d, func() (interface{}, sendgrid.RequestError) {
		return c.DeleteTemplateVersion(ctx, d.Get("template_id").(string), d.Id())
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSendgridTemplateVersionImport(
	ctx context.Context,
	d *schema.ResourceData,
	_ interface{},
) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != ImportSplitParts {
		return nil, ErrInvalidImportFormat
	}

	//nolint:errcheck
	d.Set("template_id", parts[0])
	d.SetId(parts[1])

	return []*schema.ResourceData{d}, nil
}
