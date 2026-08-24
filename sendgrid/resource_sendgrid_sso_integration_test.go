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

func TestAccSendgridSSOIntegrationBasic(t *testing.T) {
	name := "terraform-sso-" + acctest.RandString(10)
	issuer := "https://test-" + acctest.RandString(10) + ".example.com"
	signonURL := "https://sso-" + acctest.RandString(10) + ".example.com/sso"
	signoutURL := "https://sso-" + acctest.RandString(10) + ".example.com/logout"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridSSOIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridSSOIntegrationConfigBasic(name, issuer, signonURL, signoutURL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSSOIntegrationExists("sendgrid_sso_integration.test"),
					resource.TestCheckResourceAttr("sendgrid_sso_integration.test", "name", name),
					resource.TestCheckResourceAttr("sendgrid_sso_integration.test", "issuer", issuer),
					resource.TestCheckResourceAttr("sendgrid_sso_integration.test", "signon_url", signonURL),
					resource.TestCheckResourceAttr("sendgrid_sso_integration.test", "signout_url", signoutURL),
					resource.TestCheckResourceAttr("sendgrid_sso_integration.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccSendgridSSOIntegrationEnabled(t *testing.T) {
	name := "terraform-sso-enabled-" + acctest.RandString(10)
	issuer := "https://test-enabled-" + acctest.RandString(10) + ".example.com"
	signonURL := "https://sso-enabled-" + acctest.RandString(10) + ".example.com/sso"
	signoutURL := "https://sso-enabled-" + acctest.RandString(10) + ".example.com/logout"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridSSOIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridSSOIntegrationConfigEnabled(name, issuer, signonURL, signoutURL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSSOIntegrationExists("sendgrid_sso_integration.enabled"),
					resource.TestCheckResourceAttr("sendgrid_sso_integration.enabled", "name", name),
					resource.TestCheckResourceAttr("sendgrid_sso_integration.enabled", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccSendgridSSOIntegrationUpdate(t *testing.T) {
	name := "terraform-sso-update-" + acctest.RandString(10)
	nameUpdated := "terraform-sso-updated-" + acctest.RandString(10)
	issuer := "https://test-update-" + acctest.RandString(10) + ".example.com"
	signonURL := "https://sso-update-" + acctest.RandString(10) + ".example.com/sso"
	signoutURL := "https://sso-update-" + acctest.RandString(10) + ".example.com/logout"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridSSOIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridSSOIntegrationConfigBasic(name, issuer, signonURL, signoutURL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSSOIntegrationExists("sendgrid_sso_integration.test"),
					resource.TestCheckResourceAttr("sendgrid_sso_integration.test", "name", name),
				),
			},
			{
				Config: testAccCheckSendgridSSOIntegrationConfigBasic(nameUpdated, issuer, signonURL, signoutURL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSSOIntegrationExists("sendgrid_sso_integration.test"),
					resource.TestCheckResourceAttr("sendgrid_sso_integration.test", "name", nameUpdated),
				),
			},
		},
	})
}

func TestAccSendgridSSOIntegrationWithRateLimiting(t *testing.T) {
	name := "terraform-sso-rate-" + acctest.RandString(10)
	issuer := "https://test-rate-" + acctest.RandString(10) + ".example.com"
	signonURL := "https://sso-rate-" + acctest.RandString(10) + ".example.com/sso"
	signoutURL := "https://sso-rate-" + acctest.RandString(10) + ".example.com/logout"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridSSOIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridSSOIntegrationConfigWithTimeouts(name, issuer, signonURL, signoutURL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSSOIntegrationExists("sendgrid_sso_integration.rate_limit"),
					resource.TestCheckResourceAttr("sendgrid_sso_integration.rate_limit", "name", name),
				),
			},
		},
	})
}

func testAccCheckSendgridSSOIntegrationDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sendgrid_sso_integration" {
			continue
		}

		integrationID := rs.Primary.ID
		ctx := context.Background()

		_, err := c.ReadSSOIntegration(ctx, integrationID)
		if err.StatusCode != 404 {
			return fmt.Errorf("SSO integration still exists: %s", integrationID)
		}
	}

	return nil
}

func testAccCheckSendgridSSOIntegrationConfigBasic(name, issuer, signonURL, signoutURL string) string {
	return fmt.Sprintf(`
resource "sendgrid_sso_integration" "test" {
	name        = "%s"
	issuer      = "%s"
	signon_url  = "%s"
	signout_url = "%s"
	enabled     = false
}
`, name, issuer, signonURL, signoutURL)
}

func testAccCheckSendgridSSOIntegrationConfigEnabled(name, issuer, signonURL, signoutURL string) string {
	return fmt.Sprintf(`
resource "sendgrid_sso_integration" "enabled" {
	name        = "%s"
	issuer      = "%s"
	signon_url  = "%s"
	signout_url = "%s"
	enabled     = true
}
`, name, issuer, signonURL, signoutURL)
}

func testAccCheckSendgridSSOIntegrationConfigWithTimeouts(name, issuer, signonURL, signoutURL string) string {
	return fmt.Sprintf(`
resource "sendgrid_sso_integration" "rate_limit" {
	name        = "%s"
	issuer      = "%s"
	signon_url  = "%s"
	signout_url = "%s"
	enabled     = false

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}
`, name, issuer, signonURL, signoutURL)
}

func testAccCheckSendgridSSOIntegrationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSO integration ID set")
		}

		config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")
		ctx := context.Background()

		_, err := c.ReadSSOIntegration(ctx, rs.Primary.ID)
		if err.Err != nil {
			return fmt.Errorf("SSO integration not found: %s", rs.Primary.ID)
		}

		return nil
	}
}
