package sendgrid_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	provider "github.com/arslanbekov/terraform-provider-sendgrid/sendgrid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccSendgridTeammate_basic(t *testing.T) {
	email := "terraform-test-" + acctest.RandString(10) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridTeammateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridTeammateConfigBasic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridTeammateExists("sendgrid_teammate.test"),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "email", email),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "is_admin", "false"),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "is_sso", "false"),
				),
			},
		},
	})
}

func TestAccSendgridTeammate_admin(t *testing.T) {
	email := "terraform-admin-" + acctest.RandString(10) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridTeammateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridTeammateConfigAdmin(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridTeammateExists("sendgrid_teammate.test"),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "email", email),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "is_admin", "true"),
				),
			},
		},
	})
}

func TestAccSendgridTeammate_withValidScopes(t *testing.T) {
	email := "terraform-scopes-" + acctest.RandString(10) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridTeammateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridTeammateConfigWithScopes(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridTeammateExists("sendgrid_teammate.test"),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "email", email),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "is_admin", "false"),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "scopes.#", "3"),
				),
			},
		},
	})
}

func TestAccSendgridTeammate_invalidScopes(t *testing.T) {
	email := "terraform-invalid-" + acctest.RandString(10) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckSendgridTeammateConfigInvalidScopes(email),
				ExpectError: regexp.MustCompile("the following scopes are not valid or assignable"),
			},
		},
	})
}

func TestAccSendgridTeammate_automaticScopes(t *testing.T) {
	email := "terraform-auto-" + acctest.RandString(10) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckSendgridTeammateConfigAutomaticScopes(email),
				ExpectError: regexp.MustCompile("set automatically by SendGrid and cannot be manually assigned"),
			},
		},
	})
}

func TestAccSendgridTeammatePendingUser(t *testing.T) {
	email := "terraform-teammate-pending-test-" + acctest.RandString(10) + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridTeammateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridTeammateConfigBasic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridTeammateExists("sendgrid_teammate.test"),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "email", email),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "is_admin", "false"),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "is_sso", "false"),
					// For non-SSO users, status should be pending initially
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "user_status", "pending"),
				),
			},
			// Test that we can update a pending user
			{
				Config: testAccCheckSendgridTeammateConfigWithScopes(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridTeammateExists("sendgrid_teammate.test"),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "email", email),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "user_status", "pending"),
					resource.TestCheckResourceAttr("sendgrid_teammate.test", "scopes.#", "3"),
				),
			},
		},
	})
}

func testAccCheckSendgridTeammateDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sendgrid_teammate" {
			continue
		}

		email := rs.Primary.ID
		ctx := context.Background()

		_, err := c.ReadUser(ctx, email)
		if err.StatusCode != 404 {
			return fmt.Errorf("teammate still exists: %s", email)
		}
	}

	return nil
}

func testAccCheckSendgridTeammateExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No teammate email set")
		}

		config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")
		ctx := context.Background()

		_, err := c.ReadUser(ctx, rs.Primary.ID)
		if err.Err != nil {
			return fmt.Errorf("teammate not found: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSendgridTeammateConfigBasic(email string) string {
	return fmt.Sprintf(`
resource "sendgrid_teammate" "test" {
	email    = "%s"
	is_admin = false
	is_sso   = false
	scopes   = ["mail.send"]
}
`, email)
}

func testAccCheckSendgridTeammateConfigAdmin(email string) string {
	return fmt.Sprintf(`
resource "sendgrid_teammate" "test" {
	email    = "%s"
	is_admin = true
	is_sso   = false
}
`, email)
}

func testAccCheckSendgridTeammateConfigWithScopes(email string) string {
	return fmt.Sprintf(`
resource "sendgrid_teammate" "test" {
	email    = "%s"
	is_admin = false
	is_sso   = false
	scopes   = [
		"mail.send",
		"templates.read",
		"stats.read"
	]
}
`, email)
}

func testAccCheckSendgridTeammateConfigInvalidScopes(email string) string {
	return fmt.Sprintf(`
resource "sendgrid_teammate" "test" {
	email    = "%s"
	is_admin = false
	is_sso   = false
	scopes   = [
		"mail.send",
		"invalid.scope",
		"another.invalid"
	]
}
`, email)
}

func testAccCheckSendgridTeammateConfigAutomaticScopes(email string) string {
	return fmt.Sprintf(`
resource "sendgrid_teammate" "test" {
	email    = "%s"
	is_admin = false
	is_sso   = false
	scopes   = [
		"mail.send",
		"2fa_exempt"
	]
}
`, email)
}
