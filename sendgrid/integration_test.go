package sendgrid_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSendgridIntegrationEmailWorkflow(t *testing.T) {
	// Test a complete email workflow: template + teammate + api key + unsubscribe group
	templateName := "terraform-integration-" + acctest.RandString(10)
	versionName := "terraform-version-" + acctest.RandString(10)
	apiKeyName := "terraform-api-key-" + acctest.RandString(10)
	teammateEmail := "terraform-teammate-" + acctest.RandString(10) + "@example.com"
	unsubscribeName := "terraform-unsubscribe-" + acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSendgridIntegrationEmailWorkflowConfig(
					templateName, versionName, apiKeyName, teammateEmail, unsubscribeName,
				),
				Check: resource.ComposeTestCheckFunc(
					// Template checks
					resource.TestCheckResourceAttr("sendgrid_template.integration", "name", templateName),
					resource.TestCheckResourceAttr("sendgrid_template_version.integration", "name", versionName),

					// API Key checks
					resource.TestCheckResourceAttr("sendgrid_api_key.integration", "name", apiKeyName),
					resource.TestCheckResourceAttr("sendgrid_api_key.integration", "scopes.#", "2"),

					// Teammate checks
					resource.TestCheckResourceAttr("sendgrid_teammate.integration", "email", teammateEmail),
					resource.TestCheckResourceAttr("sendgrid_teammate.integration", "is_admin", "false"),

					// Unsubscribe group checks
					resource.TestCheckResourceAttr("sendgrid_unsubscribe_group.integration", "name", unsubscribeName),

					// Cross-references work
					resource.TestCheckResourceAttrSet("sendgrid_template_version.integration", "template_id"),
				),
			},
		},
	})
}

func TestAccSendgridIntegrationRateLimitingStress(t *testing.T) {
	// Test rate limiting with multiple resources created simultaneously
	prefix := "terraform-stress-" + acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSendgridIntegrationRateLimitingStressConfig(prefix),
				Check: resource.ComposeTestCheckFunc(
					// Check all API keys were created
					resource.TestCheckResourceAttr("sendgrid_api_key.stress_0", "name", prefix+"-api-key-0"),
					resource.TestCheckResourceAttr("sendgrid_api_key.stress_1", "name", prefix+"-api-key-1"),
					resource.TestCheckResourceAttr("sendgrid_api_key.stress_2", "name", prefix+"-api-key-2"),

					// Check all templates were created
					resource.TestCheckResourceAttr("sendgrid_template.stress_0", "name", prefix+"-template-0"),
					resource.TestCheckResourceAttr("sendgrid_template.stress_1", "name", prefix+"-template-1"),

					// Check all teammates were created
					resource.TestCheckResourceAttr("sendgrid_teammate.stress_0", "email", prefix+"-teammate-0@example.com"),
					resource.TestCheckResourceAttr("sendgrid_teammate.stress_1", "email", prefix+"-teammate-1@example.com"),
				),
			},
		},
	})
}

func TestAccSendgridIntegrationDomainSetup(t *testing.T) {
	// Test domain authentication + link branding together
	domain := "test-" + acctest.RandString(10) + ".example.com"
	linkDomain := "links-" + acctest.RandString(10) + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSendgridIntegrationDomainSetupConfig(domain, linkDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_domain_authentication.integration", "domain", domain),
					resource.TestCheckResourceAttr("sendgrid_link_branding.integration", "domain", linkDomain),
					resource.TestCheckResourceAttr("sendgrid_domain_authentication.integration", "default", "false"),
					resource.TestCheckResourceAttr("sendgrid_link_branding.integration", "default", "false"),
				),
			},
		},
	})
}

// Config functions
func testAccSendgridIntegrationEmailWorkflowConfig(templateName, versionName, apiKeyName, teammateEmail, unsubscribeName string) string {
	return fmt.Sprintf(`
resource "sendgrid_template" "integration" {
	name       = "%s"
	generation = "dynamic"

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_template_version" "integration" {
	template_id            = sendgrid_template.integration.id
	name                   = "%s"
	subject                = "Integration Test Email"
	html_content           = "<html><body>Hello {{name}}!</body></html>"
	generate_plain_content = true
	active                 = 1

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_api_key" "integration" {
	name   = "%s"
	scopes = ["mail.send", "templates.read"]

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_teammate" "integration" {
	email    = "%s"
	is_admin = false
	is_sso   = false
	scopes   = ["mail.send", "templates.read"]

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_unsubscribe_group" "integration" {
	name        = "%s"
	description = "Integration test unsubscribe group"
	is_default  = false

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}
`, templateName, versionName, apiKeyName, teammateEmail, unsubscribeName)
}

func testAccSendgridIntegrationRateLimitingStressConfig(prefix string) string {
	return fmt.Sprintf(`
resource "sendgrid_api_key" "stress_0" {
	name   = "%s-api-key-0"
	scopes = ["mail.send"]
	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_api_key" "stress_1" {
	name   = "%s-api-key-1"
	scopes = ["mail.send"]
	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_api_key" "stress_2" {
	name   = "%s-api-key-2"
	scopes = ["mail.send"]
	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_template" "stress_0" {
	name       = "%s-template-0"
	generation = "dynamic"
	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_template" "stress_1" {
	name       = "%s-template-1"
	generation = "dynamic"
	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_teammate" "stress_0" {
	email    = "%s-teammate-0@example.com"
	is_admin = false
	is_sso   = false
	scopes   = ["mail.send"]
	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_teammate" "stress_1" {
	email    = "%s-teammate-1@example.com"
	is_admin = false
	is_sso   = false
	scopes   = ["mail.send"]
	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}
`, prefix, prefix, prefix, prefix, prefix, prefix, prefix)
}

func testAccSendgridIntegrationDomainSetupConfig(domain, linkDomain string) string {
	return fmt.Sprintf(`
resource "sendgrid_domain_authentication" "integration" {
	domain             = "%s"
	default            = false
	automatic_security = false

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_link_branding" "integration" {
	domain  = "%s"
	default = false

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}
`, domain, linkDomain)
}
