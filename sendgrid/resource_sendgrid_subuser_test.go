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

func TestAccSendgridSubuserBasic(t *testing.T) {
	username := "terraform-subuser-" + acctest.RandString(10)
	email := username + "@example.com"
	password := "TerraformTest123!"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridSubuserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridSubuserConfigBasic(username, email, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSubuserExists("sendgrid_subuser.test"),
					resource.TestCheckResourceAttr("sendgrid_subuser.test", "username", username),
					resource.TestCheckResourceAttr("sendgrid_subuser.test", "email", email),
					resource.TestCheckResourceAttr("sendgrid_subuser.test", "disabled", "false"),
				),
			},
		},
	})
}

func TestAccSendgridSubuserWithIps(t *testing.T) {
	username := "terraform-subuser-ips-" + acctest.RandString(10)
	email := username + "@example.com"
	password := "TerraformTest123!"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridSubuserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridSubuserConfigWithIps(username, email, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSubuserExists("sendgrid_subuser.ips"),
					resource.TestCheckResourceAttr("sendgrid_subuser.ips", "username", username),
					resource.TestCheckResourceAttr("sendgrid_subuser.ips", "email", email),
					resource.TestCheckResourceAttr("sendgrid_subuser.ips", "ips.#", "1"),
				),
			},
		},
	})
}

func TestAccSendgridSubuserUpdate(t *testing.T) {
	username := "terraform-subuser-update-" + acctest.RandString(10)
	email := username + "@example.com"
	password := "TerraformTest123!"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridSubuserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridSubuserConfigBasic(username, email, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSubuserExists("sendgrid_subuser.test"),
					resource.TestCheckResourceAttr("sendgrid_subuser.test", "disabled", "false"),
				),
			},
			{
				Config: testAccCheckSendgridSubuserConfigDisabled(username, email, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSubuserExists("sendgrid_subuser.test"),
					resource.TestCheckResourceAttr("sendgrid_subuser.test", "disabled", "true"),
				),
			},
		},
	})
}

func TestAccSendgridSubuserWithRateLimiting(t *testing.T) {
	username := "terraform-subuser-rate-" + acctest.RandString(10)
	email := username + "@example.com"
	password := "TerraformTest123!"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckSendgridSubuserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridSubuserConfigWithTimeouts(username, email, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridSubuserExists("sendgrid_subuser.rate_limit"),
					resource.TestCheckResourceAttr("sendgrid_subuser.rate_limit", "username", username),
				),
			},
		},
	})
}

func testAccCheckSendgridSubuserDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sendgrid_subuser" {
			continue
		}

		username := rs.Primary.ID
		ctx := context.Background()

		_, err := c.ReadSubUser(ctx, username)
		if err.StatusCode != 404 {
			return fmt.Errorf("subuser still exists: %s", username)
		}
	}

	return nil
}

func testAccCheckSendgridSubuserConfigBasic(username, email, password string) string {
	return fmt.Sprintf(`
resource "sendgrid_subuser" "test" {
	username = "%s"
	email    = "%s"
	password = "%s"
	disabled = false
}
`, username, email, password)
}

func testAccCheckSendgridSubuserConfigWithIps(username, email, password string) string {
	return fmt.Sprintf(`
resource "sendgrid_subuser" "ips" {
	username = "%s"
	email    = "%s"
	password = "%s"
	disabled = false
	ips      = ["192.168.1.1"]
}
`, username, email, password)
}

func testAccCheckSendgridSubuserConfigDisabled(username, email, password string) string {
	return fmt.Sprintf(`
resource "sendgrid_subuser" "test" {
	username = "%s"
	email    = "%s"
	password = "%s"
	disabled = true
}
`, username, email, password)
}

func testAccCheckSendgridSubuserConfigWithTimeouts(username, email, password string) string {
	return fmt.Sprintf(`
resource "sendgrid_subuser" "rate_limit" {
	username = "%s"
	email    = "%s"
	password = "%s"
	disabled = false

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}
`, username, email, password)
}

func testAccCheckSendgridSubuserExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No subuser username set")
		}

		config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")
		ctx := context.Background()

		_, err := c.ReadSubUser(ctx, rs.Primary.ID)
		if err.Err != nil {
			return fmt.Errorf("subuser not found: %s", rs.Primary.ID)
		}

		return nil
	}
}
