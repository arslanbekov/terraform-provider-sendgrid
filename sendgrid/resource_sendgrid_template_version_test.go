package sendgrid_test

import (
	"context"
	"fmt"
	"testing"

	provider "github.com/arslanbekov/terraform-provider-sendgrid/sendgrid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccSendgridTemplateVersionBasic(t *testing.T) {
	templateName := "terraform-template-" + acctest.RandString(10)
	templateVersionName := "terraform-template-version-" + acctest.RandString(10)
	subject := "terraform-subject-" + acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridTemplateVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridTemplateVersionConfigBasic(
					templateName,
					templateVersionName,
					subject,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridTemplateVersionExists("sendgrid_template_version.new"),
				),
			},
		},
	})
}

func testAccCheckSendgridTemplateVersionDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sendgrid_template_version" {
			continue
		}

		templateID := rs.Primary.Attributes["template_id"]
		id := rs.Primary.ID

		ctx := context.Background()
		_, requestErr := c.DeleteTemplateVersion(ctx, templateID, id)
		if requestErr.Err != nil {
			return requestErr.Err
		}
	}

	return nil
}

func testAccCheckSendgridTemplateVersionConfigBasic(
	templateName, templateVersionName, subject string,
) string {
	return fmt.Sprintf(`
	resource "sendgrid_template" "template" {
		name = %s
		generation = "dynamic"
	}
	resource "sendgrid_template_version" "template_version" {
		template_id = sendgrid.template.template.id
		name = %s
		subject = %s
	}
	`, templateName, templateVersionName, subject)
}

func testAccCheckSendgridTemplateVersionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No templateVersionID set")
		}

		return nil
	}
}
