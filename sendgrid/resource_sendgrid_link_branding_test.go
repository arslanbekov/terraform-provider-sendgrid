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

func TestAccSendgridLinkBrandingBasic(t *testing.T) {
	domain := "links-" + acctest.RandString(10) + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridLinkBrandingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridLinkBrandingConfigBasic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridLinkBrandingExists("sendgrid_link_branding.test"),
					resource.TestCheckResourceAttr("sendgrid_link_branding.test", "domain", domain),
					resource.TestCheckResourceAttr("sendgrid_link_branding.test", "default", "false"),
				),
			},
		},
	})
}

func TestAccSendgridLinkBrandingWithSubdomain(t *testing.T) {
	domain := "links-" + acctest.RandString(10) + ".example.com"
	subdomain := "mail"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridLinkBrandingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridLinkBrandingConfigWithSubdomain(domain, subdomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridLinkBrandingExists("sendgrid_link_branding.subdomain"),
					resource.TestCheckResourceAttr("sendgrid_link_branding.subdomain", "domain", domain),
					resource.TestCheckResourceAttr("sendgrid_link_branding.subdomain", "subdomain", subdomain),
					resource.TestCheckResourceAttr("sendgrid_link_branding.subdomain", "default", "false"),
				),
			},
		},
	})
}

func TestAccSendgridLinkBrandingUpdate(t *testing.T) {
	domain := "links-update-" + acctest.RandString(10) + ".example.com"
	domainUpdated := "links-updated-" + acctest.RandString(10) + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridLinkBrandingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridLinkBrandingConfigBasic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridLinkBrandingExists("sendgrid_link_branding.test"),
					resource.TestCheckResourceAttr("sendgrid_link_branding.test", "domain", domain),
				),
			},
			{
				Config: testAccCheckSendgridLinkBrandingConfigBasic(domainUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridLinkBrandingExists("sendgrid_link_branding.test"),
					resource.TestCheckResourceAttr("sendgrid_link_branding.test", "domain", domainUpdated),
				),
			},
		},
	})
}

func TestAccSendgridLinkBrandingWithRateLimiting(t *testing.T) {
	domain := "links-rate-limit-" + acctest.RandString(10) + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridLinkBrandingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridLinkBrandingConfigWithTimeouts(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridLinkBrandingExists("sendgrid_link_branding.rate_limit"),
					resource.TestCheckResourceAttr("sendgrid_link_branding.rate_limit", "domain", domain),
				),
			},
		},
	})
}

func testAccCheckSendgridLinkBrandingDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sendgrid_link_branding" {
			continue
		}

		linkBrandingID := rs.Primary.ID
		ctx := context.Background()

		_, err := c.ReadLinkBranding(ctx, linkBrandingID)
		if err.StatusCode != 404 {
			return fmt.Errorf("link branding still exists: %s", linkBrandingID)
		}
	}

	return nil
}

func testAccCheckSendgridLinkBrandingConfigBasic(domain string) string {
	return fmt.Sprintf(`
resource "sendgrid_link_branding" "test" {
	domain  = "%s"
	default = false
}
`, domain)
}

func testAccCheckSendgridLinkBrandingConfigWithSubdomain(domain, subdomain string) string {
	return fmt.Sprintf(`
resource "sendgrid_link_branding" "subdomain" {
	domain    = "%s"
	subdomain = "%s"
	default   = false
}
`, domain, subdomain)
}

func testAccCheckSendgridLinkBrandingConfigWithTimeouts(domain string) string {
	return fmt.Sprintf(`
resource "sendgrid_link_branding" "rate_limit" {
	domain  = "%s"
	default = false

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}
`, domain)
}

func testAccCheckSendgridLinkBrandingExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No link branding ID set")
		}

		config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")
		ctx := context.Background()

		_, err := c.ReadLinkBranding(ctx, rs.Primary.ID)
		if err.Err != nil {
			return fmt.Errorf("link branding not found: %s", rs.Primary.ID)
		}

		return nil
	}
}
