package sendgrid_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSendgridTeammate(t *testing.T) {
	email := "terraform-teammate-data-" + acctest.RandString(10) + "@example.com"
	scopes := []string{"mail.send", "marketing.read"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSendgridTeammateConfig(email, scopes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.sendgrid_teammate.test", "email", email),
					resource.TestCheckResourceAttr("data.sendgrid_teammate.test", "is_admin", "false"),
					resource.TestCheckResourceAttr("data.sendgrid_teammate.test", "scopes.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceSendgridTemplate(t *testing.T) {
	templateName := "terraform-template-data-" + acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSendgridTemplateConfig(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.sendgrid_template.test", "name", templateName),
					resource.TestCheckResourceAttr("data.sendgrid_template.test", "generation", "dynamic"),
					resource.TestCheckResourceAttrSet("data.sendgrid_template.test", "id"),
				),
			},
		},
	})
}

func TestAccDataSourceSendgridUnsubscribeGroup(t *testing.T) {
	name := "terraform-unsubscribe-data-" + acctest.RandString(10)
	description := "Test unsubscribe group for data source"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSendgridUnsubscribeGroupConfig(name, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.sendgrid_unsubscribe_group.test", "name", name),
					resource.TestCheckResourceAttr("data.sendgrid_unsubscribe_group.test", "description", description),
					resource.TestCheckResourceAttrSet("data.sendgrid_unsubscribe_group.test", "id"),
				),
			},
		},
	})
}

func TestAccDataSourceSendgridTemplateVersion(t *testing.T) {
	templateName := "terraform-template-version-data-" + acctest.RandString(10)
	versionName := "terraform-version-data-" + acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSendgridTemplateVersionConfig(templateName, versionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.sendgrid_template_version.test", "name", versionName),
					resource.TestCheckResourceAttr("data.sendgrid_template_version.test", "subject", "Test Subject"),
					resource.TestCheckResourceAttrSet("data.sendgrid_template_version.test", "template_id"),
					resource.TestCheckResourceAttrSet("data.sendgrid_template_version.test", "id"),
				),
			},
		},
	})
}

// Config functions
func testAccDataSourceSendgridTeammateConfig(email string, scopes []string) string {
	return fmt.Sprintf(`
resource "sendgrid_teammate" "test" {
	email    = "%s"
	is_admin = false
	is_sso   = false
	scopes   = %q
}

data "sendgrid_teammate" "test" {
	email = sendgrid_teammate.test.email
}
`, email, scopes)
}

func testAccDataSourceSendgridTemplateConfig(name string) string {
	return fmt.Sprintf(`
resource "sendgrid_template" "test" {
	name       = "%s"
	generation = "dynamic"
}

data "sendgrid_template" "test" {
	depends_on = [sendgrid_template.test]
	name       = "%s"
}
`, name, name)
}

func testAccDataSourceSendgridUnsubscribeGroupConfig(name, description string) string {
	return fmt.Sprintf(`
resource "sendgrid_unsubscribe_group" "test" {
	name        = "%s"
	description = "%s"
	is_default  = false
}

data "sendgrid_unsubscribe_group" "test" {
	depends_on = [sendgrid_unsubscribe_group.test]
	name       = "%s"
}
`, name, description, name)
}

func testAccDataSourceSendgridTemplateVersionConfig(templateName, versionName string) string {
	return fmt.Sprintf(`
resource "sendgrid_template" "test" {
	name       = "%s"
	generation = "dynamic"
}

resource "sendgrid_template_version" "test" {
	template_id            = sendgrid_template.test.id
	name                   = "%s"
	subject                = "Test Subject"
	html_content           = "<html><body>Test</body></html>"
	generate_plain_content = true
	active                 = 1
}

data "sendgrid_template_version" "test" {
	depends_on  = [sendgrid_template_version.test]
	template_id = sendgrid_template.test.id
	name        = "%s"
}
`, templateName, versionName, versionName)
}
