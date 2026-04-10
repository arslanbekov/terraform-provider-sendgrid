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

func TestAccSendgridSSOCertificateBasic(t *testing.T) {
	integrationId := "terraform-integration-" + acctest.RandString(10)
	certificate := generateTestCertificate()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridSSOCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridSSOCertificateConfigBasic(integrationId, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSSOCertificateExists("sendgrid_sso_certificate.test"),
					resource.TestCheckResourceAttr("sendgrid_sso_certificate.test", "integration_id", integrationId),
					resource.TestCheckResourceAttr("sendgrid_sso_certificate.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccSendgridSSOCertificateEnabled(t *testing.T) {
	integrationId := "terraform-integration-enabled-" + acctest.RandString(10)
	certificate := generateTestCertificate()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridSSOCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridSSOCertificateConfigEnabled(integrationId, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSSOCertificateExists("sendgrid_sso_certificate.enabled"),
					resource.TestCheckResourceAttr("sendgrid_sso_certificate.enabled", "integration_id", integrationId),
					resource.TestCheckResourceAttr("sendgrid_sso_certificate.enabled", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccSendgridSSOCertificateUpdate(t *testing.T) {
	integrationId := "terraform-integration-update-" + acctest.RandString(10)
	certificate := generateTestCertificate()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridSSOCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridSSOCertificateConfigBasic(integrationId, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSSOCertificateExists("sendgrid_sso_certificate.test"),
					resource.TestCheckResourceAttr("sendgrid_sso_certificate.test", "enabled", "false"),
				),
			},
			{
				Config: testAccCheckSendgridSSOCertificateConfigEnabled(integrationId, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSSOCertificateExists("sendgrid_sso_certificate.enabled"),
					resource.TestCheckResourceAttr("sendgrid_sso_certificate.enabled", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccSendgridSSOCertificateWithRateLimiting(t *testing.T) {
	integrationId := "terraform-integration-rate-" + acctest.RandString(10)
	certificate := generateTestCertificate()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridSSOCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridSSOCertificateConfigWithTimeouts(integrationId, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSSOCertificateExists("sendgrid_sso_certificate.rate_limit"),
					resource.TestCheckResourceAttr("sendgrid_sso_certificate.rate_limit", "integration_id", integrationId),
				),
			},
		},
	})
}

func testAccCheckSendgridSSOCertificateDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sendgrid_sso_certificate" {
			continue
		}

		certificateID := rs.Primary.ID
		ctx := context.Background()

		_, err := c.ReadSSOCertificate(ctx, certificateID)
		if err.StatusCode != 404 {
			return fmt.Errorf("SSO certificate still exists: %s", certificateID)
		}
	}

	return nil
}

func generateTestCertificate() string {
	return `-----BEGIN CERTIFICATE-----
MIIDBTCCAe2gAwIBAgIJAKZV5i2WjC+zMA0GCSqGSIb3DQEBCwUAMBkxFzAVBgNV
BAMMDnRlc3QuZXhhbXBsZS5jb20wHhcNMjQwMTAxMDAwMDAwWhcNMjUwMTAxMDAw
MDAwWjAZMRcwFQYDVQQDDA50ZXN0LmV4YW1wbGUuY29tMIIBIjANBgkqhkiG9w0B
AQEFAAOCAQ8AMIIBCgKCAQEA1234567890abcdefghijklmnopqrstuvwxyzABCD
EFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyzABCDEF
GHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890ABCDEFGH
IJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJ
KLMNOPQRSTUVWXYZwIDAQABo1MwUTAdBgNVHQ4EFgQU1234567890abcdefghi
jklmnopqrstuvwxyz0wHwYDVR0jBBgwFoAU1234567890abcdefghijklmnopqr
stuvwxyz0wDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEA1234
567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcd
efghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcd
efghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ
-----END CERTIFICATE-----`
}

func testAccCheckSendgridSSOCertificateConfigBasic(integrationId, certificate string) string {
	return fmt.Sprintf(`
resource "sendgrid_sso_certificate" "test" {
	integration_id = "%s"
	public_certificate = "%s"
	enabled = false
}
`, integrationId, certificate)
}

func testAccCheckSendgridSSOCertificateConfigEnabled(integrationId, certificate string) string {
	return fmt.Sprintf(`
resource "sendgrid_sso_certificate" "enabled" {
	integration_id = "%s"
	public_certificate = "%s"
	enabled = true
}
`, integrationId, certificate)
}

func testAccCheckSendgridSSOCertificateConfigWithTimeouts(integrationId, certificate string) string {
	return fmt.Sprintf(`
resource "sendgrid_sso_certificate" "rate_limit" {
	integration_id = "%s"
	public_certificate = "%s"
	enabled = false

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}
`, integrationId, certificate)
}

func testAccCheckSendgridSSOCertificateExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSO certificate ID set")
		}

		config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")
		ctx := context.Background()

		_, err := c.ReadSSOCertificate(ctx, rs.Primary.ID)
		if err.Err != nil {
			return fmt.Errorf("SSO certificate not found: %s", rs.Primary.ID)
		}

		return nil
	}
}
