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

func TestAccSendgridDomainAuthenticationValidationBasic(t *testing.T) {
	domain := "test-" + acctest.RandString(10) + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridDomainAuthenticationValidationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridDomainAuthenticationValidationConfigBasic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridDomainAuthenticationExists("sendgrid_domain_authentication.test"),
					resource.TestCheckResourceAttr("sendgrid_domain_authentication.test", "domain", domain),
					resource.TestCheckResourceAttr("sendgrid_domain_authentication.test", "default", "false"),
					resource.TestCheckResourceAttr("sendgrid_domain_authentication.test", "automatic_security", "false"),
				),
			},
		},
	})
}

func testAccCheckSendgridDomainAuthenticationValidationConfigBasic(domain string) string {
	return fmt.Sprintf(`
resource "sendgrid_domain_authentication" "test" {
	domain             = "%s"
	default            = false
	automatic_security = false
}

resource "sendgrid_domain_authentication_validation" "this" {
  domain_authentication_id = sendgrid_domain_authentication.test.id
}
`, domain)
}

func testAccCheckSendgridDomainAuthenticationValidationDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sendgrid_domain_authentication_validation" {
			continue
		}

		domainValidationID := rs.Primary.ID
		ctx := context.Background()

		_, err := c.ReadDomainAuthentication(ctx, domainValidationID)
		if err.StatusCode != 404 {
			return fmt.Errorf("domain authentication validation still exists: %s", domainValidationID)
		}
	}

	return nil
}
