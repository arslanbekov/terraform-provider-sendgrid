package sendgrid_test

import (
	"context"
	"fmt"
	"testing"

	sendgrid "github.com/arslanbekov/terraform-provider-sendgrid/sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccSendgridParseWebhookBasic(t *testing.T) {
	hostname := "parse-" + acctest.RandString(10) + ".example.com"
	url := "https://example-" + acctest.RandString(10) + ".com/parse"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridParseWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridParseWebhookConfigBasic(hostname, url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridParseWebhookExists("sendgrid_parse_webhook.test"),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.test", "hostname", hostname),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.test", "url", url),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.test", "spam_check", "true"),
				),
			},
			// Import test
			{
				ResourceName:      "sendgrid_parse_webhook.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSendgridParseWebhookWithSendRaw(t *testing.T) {
	hostname := "parse-raw-" + acctest.RandString(10) + ".example.com"
	url := "https://raw-" + acctest.RandString(10) + ".com/parse"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridParseWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridParseWebhookConfigWithSendRaw(hostname, url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridParseWebhookExists("sendgrid_parse_webhook.raw"),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.raw", "hostname", hostname),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.raw", "url", url),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.raw", "send_raw", "true"),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.raw", "spam_check", "false"),
				),
			},
		},
	})
}

func TestAccSendgridParseWebhookUpdate(t *testing.T) {
	hostname := "parse-update-" + acctest.RandString(10) + ".example.com"
	url := "https://update-" + acctest.RandString(10) + ".com/parse"
	urlUpdated := "https://updated-" + acctest.RandString(10) + ".com/parse"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridParseWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridParseWebhookConfigBasic(hostname, url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridParseWebhookExists("sendgrid_parse_webhook.test"),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.test", "url", url),
				),
			},
			{
				Config: testAccCheckSendgridParseWebhookConfigBasic(hostname, urlUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridParseWebhookExists("sendgrid_parse_webhook.test"),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.test", "url", urlUpdated),
				),
			},
		},
	})
}

func TestAccSendgridParseWebhookUpdateInPlace(t *testing.T) {
	hostname := "parse-inplace-" + acctest.RandString(10) + ".example.com"
	url := "https://inplace-" + acctest.RandString(10) + ".com/parse"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridParseWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridParseWebhookConfigBasic(hostname, url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridParseWebhookExists("sendgrid_parse_webhook.test"),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.test", "spam_check", "true"),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.test", "send_raw", "false"),
				),
			},
			{
				Config: testAccCheckSendgridParseWebhookConfigUpdatedFields(hostname, url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridParseWebhookExists("sendgrid_parse_webhook.test"),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.test", "hostname", hostname),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.test", "url", url),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.test", "spam_check", "false"),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.test", "send_raw", "true"),
				),
			},
		},
	})
}

func TestAccSendgridParseWebhookWithRateLimiting(t *testing.T) {
	hostname := "parse-rate-" + acctest.RandString(10) + ".example.com"
	url := "https://rate-limit-" + acctest.RandString(10) + ".com/parse"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridParseWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridParseWebhookConfigWithTimeouts(hostname, url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridParseWebhookExists("sendgrid_parse_webhook.rate_limit"),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.rate_limit", "hostname", hostname),
				),
			},
		},
	})
}

func TestAccSendgridParseWebhookWithSecurityPolicy(t *testing.T) {
	hostname := "parse-security-policy-" + acctest.RandString(10) + ".example.com"
	url := "https://security-policy-" + acctest.RandString(10) + ".com/parse"
	policyId := acctest.RandString(32)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridParseWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridParseWebhookConfigWithWebhookSecurityPolicy(hostname, url, policyId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridParseWebhookExists("sendgrid_parse_webhook.security_policy"),
					resource.TestCheckResourceAttr("sendgrid_parse_webhook.security_policy", "webhook_security_policy_id", policyId),
				),
			},
		},
	})
}

func testAccCheckSendgridParseWebhookDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*sendgrid.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sendgrid_parse_webhook" {
			continue
		}

		hostname := rs.Primary.ID
		ctx := context.Background()

		_, err := c.ReadParseWebhook(ctx, hostname)
		if err.StatusCode != 404 {
			return fmt.Errorf("parse webhook still exists: %s", hostname)
		}
	}

	return nil
}

func testAccCheckSendgridParseWebhookConfigBasic(hostname, url string) string {
	return fmt.Sprintf(`
resource "sendgrid_parse_webhook" "test" {
	hostname   = "%s"
	url        = "%s"
	spam_check = true
	send_raw   = false
}
`, hostname, url)
}

func testAccCheckSendgridParseWebhookConfigUpdatedFields(hostname, url string) string {
	return fmt.Sprintf(`
resource "sendgrid_parse_webhook" "test" {
	hostname   = "%s"
	url        = "%s"
	spam_check = false
	send_raw   = true
}
`, hostname, url)
}

func testAccCheckSendgridParseWebhookConfigWithSendRaw(hostname, url string) string {
	return fmt.Sprintf(`
resource "sendgrid_parse_webhook" "raw" {
	hostname   = "%s"
	url        = "%s"
	spam_check = false
	send_raw   = true
}
`, hostname, url)
}

func testAccCheckSendgridParseWebhookConfigWithTimeouts(hostname, url string) string {
	return fmt.Sprintf(`
resource "sendgrid_parse_webhook" "rate_limit" {
	hostname   = "%s"
	url        = "%s"
	spam_check = true
	send_raw   = false

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}
`, hostname, url)
}

func testAccCheckSendgridParseWebhookConfigWithWebhookSecurityPolicy(hostname, url, policyId string) string {
	return fmt.Sprintf(`
resource "sendgrid_parse_webhook" "security_policy" {
	hostname   = "%s"
	url        = "%s"
	spam_check = true
	send_raw   = false
	webhook_security_policy_id = "%s"
}
`, hostname, url, policyId)
}

func testAccCheckSendgridParseWebhookExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No parse webhook hostname set")
		}

		c := testAccProvider.Meta().(*sendgrid.Client)
		ctx := context.Background()

		_, err := c.ReadParseWebhook(ctx, rs.Primary.ID)
		if err.Err != nil {
			return fmt.Errorf("parse webhook not found: %s", rs.Primary.ID)
		}

		return nil
	}
}
