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

func TestAccSendgridEventWebhookBasic(t *testing.T) {
	url := "https://example-" + acctest.RandString(10) + ".com/webhook"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridEventWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridEventWebhookConfigBasic(url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridEventWebhookExists("sendgrid_event_webhook.test"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.test", "url", url),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("sendgrid_event_webhook.test", "id"),
				),
			},
			// Import test
			{
				ResourceName:            "sendgrid_event_webhook.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"oauth_client_secret"},
			},
		},
	})
}

func TestAccSendgridEventWebhookWithEvents(t *testing.T) {
	url := "https://events-" + acctest.RandString(10) + ".com/webhook"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridEventWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridEventWebhookConfigWithEvents(url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridEventWebhookExists("sendgrid_event_webhook.events"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.events", "url", url),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.events", "enabled", "true"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.events", "bounce", "true"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.events", "click", "true"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.events", "deferred", "false"),
				),
			},
		},
	})
}

func TestAccSendgridEventWebhookUpdate(t *testing.T) {
	url := "https://update-" + acctest.RandString(10) + ".com/webhook"
	urlUpdated := "https://updated-" + acctest.RandString(10) + ".com/webhook"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridEventWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridEventWebhookConfigBasic(url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridEventWebhookExists("sendgrid_event_webhook.test"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.test", "url", url),
				),
			},
			{
				Config: testAccCheckSendgridEventWebhookConfigBasic(urlUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridEventWebhookExists("sendgrid_event_webhook.test"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.test", "url", urlUpdated),
				),
			},
		},
	})
}

func TestAccSendgridEventWebhookWithRateLimiting(t *testing.T) {
	url := "https://rate-limit-" + acctest.RandString(10) + ".com/webhook"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridEventWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridEventWebhookConfigWithTimeouts(url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridEventWebhookExists("sendgrid_event_webhook.rate_limit"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.rate_limit", "url", url),
				),
			},
		},
	})
}

func TestAccSendgridEventWebhookWithFriendlyName(t *testing.T) {
	url := "https://friendly-" + acctest.RandString(10) + ".com/webhook"
	friendlyName := "Test Webhook " + acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridEventWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridEventWebhookConfigWithFriendlyName(url, friendlyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridEventWebhookExists("sendgrid_event_webhook.friendly"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.friendly", "url", url),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.friendly", "enabled", "true"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.friendly", "friendly_name", friendlyName),
				),
			},
		},
	})
}

func TestAccSendgridEventWebhookFriendlyNameUpdate(t *testing.T) {
	url := "https://update-name-" + acctest.RandString(10) + ".com/webhook"
	friendlyName := "Original Name " + acctest.RandString(5)
	friendlyNameUpdated := "Updated Name " + acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridEventWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridEventWebhookConfigWithFriendlyName(url, friendlyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridEventWebhookExists("sendgrid_event_webhook.friendly"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.friendly", "friendly_name", friendlyName),
				),
			},
			{
				Config: testAccCheckSendgridEventWebhookConfigWithFriendlyName(url, friendlyNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridEventWebhookExists("sendgrid_event_webhook.friendly"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.friendly", "friendly_name", friendlyNameUpdated),
				),
			},
		},
	})
}

func TestAccSendgridEventWebhookMultiple(t *testing.T) {
	urlDev := "https://dev-" + acctest.RandString(10) + ".com/webhook"
	urlProd := "https://prod-" + acctest.RandString(10) + ".com/webhook"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridEventWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridEventWebhookConfigMultiple(urlDev, urlProd),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridEventWebhookExists("sendgrid_event_webhook.dev"),
					testAccCheckSendgridEventWebhookExists("sendgrid_event_webhook.prod"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.dev", "url", urlDev),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.prod", "url", urlProd),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.dev", "friendly_name", "Development Webhook"),
					resource.TestCheckResourceAttr("sendgrid_event_webhook.prod", "friendly_name", "Production Webhook"),
					resource.TestCheckResourceAttrSet("sendgrid_event_webhook.dev", "id"),
					resource.TestCheckResourceAttrSet("sendgrid_event_webhook.prod", "id"),
					// Verify IDs are different
					testAccCheckSendgridEventWebhookIDsDifferent("sendgrid_event_webhook.dev", "sendgrid_event_webhook.prod"),
				),
			},
		},
	})
}

func testAccCheckSendgridEventWebhookDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sendgrid_event_webhook" {
			continue
		}

		ctx := context.Background()
		_, err := c.ReadEventWebhook(ctx, rs.Primary.ID)
		if err.StatusCode != 404 && err.Err == nil {
			return fmt.Errorf("event webhook %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckSendgridEventWebhookConfigBasic(url string) string {
	return fmt.Sprintf(`
resource "sendgrid_event_webhook" "test" {
	url     = "%s"
	enabled = true
}
`, url)
}

func testAccCheckSendgridEventWebhookConfigWithEvents(url string) string {
	return fmt.Sprintf(`
resource "sendgrid_event_webhook" "events" {
	url      = "%s"
	enabled  = true
	bounce   = true
	click    = true
	deferred = false
	delivered = true
	dropped  = false
}
`, url)
}

func testAccCheckSendgridEventWebhookConfigWithTimeouts(url string) string {
	return fmt.Sprintf(`
resource "sendgrid_event_webhook" "rate_limit" {
	url     = "%s"
	enabled = true

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}
`, url)
}

func testAccCheckSendgridEventWebhookConfigWithFriendlyName(url, friendlyName string) string {
	return fmt.Sprintf(`
resource "sendgrid_event_webhook" "friendly" {
	url           = "%s"
	enabled       = true
	friendly_name = "%s"

	bounce      = true
	delivered   = true
	open        = true
	click       = true
}
`, url, friendlyName)
}

func testAccCheckSendgridEventWebhookConfigMultiple(urlDev, urlProd string) string {
	return fmt.Sprintf(`
resource "sendgrid_event_webhook" "dev" {
	url           = "%s"
	enabled       = true
	friendly_name = "Development Webhook"

	bounce      = true
	delivered   = true
	open        = true
	click       = true
}

resource "sendgrid_event_webhook" "prod" {
	url           = "%s"
	enabled       = true
	friendly_name = "Production Webhook"

	bounce      = true
	delivered   = true
	open        = true
	click       = true
	spam_report = true
}
`, urlDev, urlProd)
}

func testAccCheckSendgridEventWebhookExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No event webhook ID set")
		}

		config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")
		ctx := context.Background()

		_, err := c.ReadEventWebhook(ctx, rs.Primary.ID)
		if err.Err != nil {
			return fmt.Errorf("event webhook not found: %v", err.Err)
		}

		return nil
	}
}

func testAccCheckSendgridEventWebhookIDsDifferent(name1, name2 string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs1, ok := s.RootModule().Resources[name1]
		if !ok {
			return fmt.Errorf("Not found: %s", name1)
		}

		rs2, ok := s.RootModule().Resources[name2]
		if !ok {
			return fmt.Errorf("Not found: %s", name2)
		}

		if rs1.Primary.ID == rs2.Primary.ID {
			return fmt.Errorf("Expected different IDs for %s and %s, but both have ID %s", name1, name2, rs1.Primary.ID)
		}

		return nil
	}
}
