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

func TestAccSendgridWebhookSecurityPolicyOidcOnly(t *testing.T) {
	client_id := "client-" + acctest.RandString(10)
	client_secret := "secret" + acctest.RandString(10)
	token_url := "http://oauth." + acctest.RandString(10) + ".example.com/user/123/token"
	name := "OIDC Only Policy" + acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridWebhookSecurityPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridWebhookSecurityPolicyOidcOnly(name, client_id, client_secret, token_url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridWebhookSecurityPolicyExists("sendgrid_webhook_security_policy.oidc_only"),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.oidc_only", "name", name),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.oidc_only", "oauth.0.client_id", client_id),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.oidc_only", "oauth.0.client_secret", client_secret),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.oidc_only", "oauth.0.token_url", token_url),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.oidc_only", "oauth.0.scopes.0", "webhooks:read"),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.oidc_only", "oauth.0.scopes.1", "webhooks:write"),
				),
			},
		},
	})
}

func testAccCheckSendgridWebhookSecurityPolicyOidcOnly(name, client_id, client_secret, token_url string) string {
	return fmt.Sprintf(`
resource "sendgrid_webhook_security_policy" "oidc_only" {
	name = "%s"

	oauth {
	  client_id     = "%s"
	  client_secret = "%s"
	  token_url     = "%s"
	  scopes        = ["webhooks:read", "webhooks:write"]
	}
}
`, name, client_id, client_secret, token_url)
}

func testAccCheckSendgridWebhookSecurityPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No webhook security policy ID set")
		}

		config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")
		ctx := context.Background()

		_, err := c.ReadWebhookSecurityPolicy(ctx, rs.Primary.ID)
		if err.Err != nil {
			return fmt.Errorf("webhook security policy not found: %s", rs.Primary.ID)
		}

		return nil
	}
}

func TestAccSendgridWebhookSecurityPolicySignatureOnly(t *testing.T) {
	name := "Signature Only Policy" + acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridWebhookSecurityPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridWebhookSecurityPolicySignatureOnly(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridWebhookSecurityPolicyExists("sendgrid_webhook_security_policy.signature_only"),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.signature_only", "name", name),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.signature_only", "signature.0.enabled", "true"),
					testAccCheckSendgridWebhookSecurityPolicyPublicKeyExists("sendgrid_webhook_security_policy.signature_only"),
				),
			},
		},
	})
}

func testAccCheckSendgridWebhookSecurityPolicyPublicKeyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		// Check if public_key attribute exists and is not empty
		publicKey, ok := rs.Primary.Attributes["signature.0.public_key"]
		if !ok {
			return fmt.Errorf("public_key attribute not found in state")
		}

		if publicKey == "" {
			return fmt.Errorf("public_key is empty, expected a non-empty value")
		}

		return nil
	}
}

func testAccCheckSendgridWebhookSecurityPolicySignatureOnly(name string) string {
	return fmt.Sprintf(`
resource "sendgrid_webhook_security_policy" "signature_only" {
	name = "%s"

	signature {
	  enabled = true
	}
}
`, name)
}

func TestAccSendgridWebhookSecurityPolicyOidcAndSignature(t *testing.T) {
	client_id := "client-" + acctest.RandString(10)
	client_secret := "secret" + acctest.RandString(10)
	token_url := "http://oauth." + acctest.RandString(10) + ".example.com/user/123/token"
	name := "OIDC and Signature Policy" + acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridWebhookSecurityPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridWebhookSecurityPolicySignatureAndOidc(name, client_id, client_secret, token_url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridWebhookSecurityPolicyExists("sendgrid_webhook_security_policy.signature_and_oidc"),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.signature_and_oidc", "name", name),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.signature_and_oidc", "oauth.0.client_id", client_id),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.signature_and_oidc", "oauth.0.client_secret", client_secret),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.signature_and_oidc", "oauth.0.token_url", token_url),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.signature_and_oidc", "oauth.0.scopes.0", "webhooks:read"),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.signature_and_oidc", "oauth.0.scopes.1", "webhooks:write"),
					testAccCheckSendgridWebhookSecurityPolicyPublicKeyExists("sendgrid_webhook_security_policy.signature_and_oidc"),
				),
			},
		},
	})
}

func testAccCheckSendgridWebhookSecurityPolicySignatureAndOidc(name, client_id, client_secret, token_url string) string {
	return fmt.Sprintf(`
resource "sendgrid_webhook_security_policy" "signature_and_oidc" {
	name = "%s"

	oauth {
	  client_id     = "%s"
	  client_secret = "%s"
	  token_url     = "%s"
	  scopes        = ["webhooks:read", "webhooks:write"]
	}

	signature {
	  enabled = true
	}
}
`, name, client_id, client_secret, token_url)
}

func TestAccSendgridWebhookSecurityPolicyUpdate(t *testing.T) {
	name := "Signature Policy " + acctest.RandString(10)
	nameUpdated := "Updated Signature Policy " + acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridWebhookSecurityPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridWebhookSecurityPolicySignatureUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridWebhookSecurityPolicyExists("sendgrid_webhook_security_policy.update_test"),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.update_test", "name", name),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.update_test", "signature.0.enabled", "true"),
				),
			},
			{
				Config: testAccCheckSendgridWebhookSecurityPolicySignatureUpdate(nameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridWebhookSecurityPolicyExists("sendgrid_webhook_security_policy.update_test"),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.update_test", "name", nameUpdated),
					resource.TestCheckResourceAttr("sendgrid_webhook_security_policy.update_test", "signature.0.enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckSendgridWebhookSecurityPolicySignatureUpdate(name string) string {
	return fmt.Sprintf(`
resource "sendgrid_webhook_security_policy" "update_test" {
	name = "%s"

	signature {
	  enabled = true
	}
}
`, name)
}

func testAccCheckSendgridWebhookSecurityPolicyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sendgrid_webhook_security_policy" {
			continue
		}

		policyId := rs.Primary.ID
		ctx := context.Background()

		_, err := c.ReadWebhookSecurityPolicy(ctx, policyId)
		if err.StatusCode != 404 {
			return fmt.Errorf("webhook security policy still exists: %s", policyId)
		}
	}

	return nil
}
