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

func TestAccSendgridUnsubscribeGroupBasic(t *testing.T) {
	name := "terraform-unsubscribe-" + acctest.RandString(10)
	description := "Test unsubscribe group created by Terraform"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridUnsubscribeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridUnsubscribeGroupConfigBasic(name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridUnsubscribeGroupExists("sendgrid_unsubscribe_group.test"),
					resource.TestCheckResourceAttr("sendgrid_unsubscribe_group.test", "name", name),
					resource.TestCheckResourceAttr("sendgrid_unsubscribe_group.test", "description", description),
					resource.TestCheckResourceAttr("sendgrid_unsubscribe_group.test", "is_default", "false"),
				),
			},
		},
	})
}

func TestAccSendgridUnsubscribeGroupUpdate(t *testing.T) {
	name := "terraform-unsubscribe-" + acctest.RandString(10)
	description := "Test unsubscribe group created by Terraform"
	nameUpdated := "terraform-unsubscribe-updated-" + acctest.RandString(10)
	descriptionUpdated := "Updated test unsubscribe group"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridUnsubscribeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridUnsubscribeGroupConfigBasic(name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridUnsubscribeGroupExists("sendgrid_unsubscribe_group.test"),
					resource.TestCheckResourceAttr("sendgrid_unsubscribe_group.test", "name", name),
					resource.TestCheckResourceAttr("sendgrid_unsubscribe_group.test", "description", description),
				),
			},
			{
				Config: testAccCheckSendgridUnsubscribeGroupConfigBasic(nameUpdated, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridUnsubscribeGroupExists("sendgrid_unsubscribe_group.test"),
					resource.TestCheckResourceAttr("sendgrid_unsubscribe_group.test", "name", nameUpdated),
					resource.TestCheckResourceAttr("sendgrid_unsubscribe_group.test", "description", descriptionUpdated),
				),
			},
		},
	})
}

func TestAccSendgridUnsubscribeGroupWithRateLimiting(t *testing.T) {
	name := "terraform-rate-limit-" + acctest.RandString(10)
	description := "Test unsubscribe group with rate limiting"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSendgridUnsubscribeGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSendgridUnsubscribeGroupConfigWithTimeouts(name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSendgridUnsubscribeGroupExists("sendgrid_unsubscribe_group.rate_limit"),
					resource.TestCheckResourceAttr("sendgrid_unsubscribe_group.rate_limit", "name", name),
					resource.TestCheckResourceAttr("sendgrid_unsubscribe_group.rate_limit", "description", description),
				),
			},
		},
	})
}

func testAccCheckSendgridUnsubscribeGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sendgrid_unsubscribe_group" {
			continue
		}

		groupID := rs.Primary.ID
		ctx := context.Background()

		_, err := c.ReadUnsubscribeGroup(ctx, groupID)
		if err.StatusCode != 404 {
			return fmt.Errorf("unsubscribe group still exists: %s", groupID)
		}
	}

	return nil
}

func testAccCheckSendgridUnsubscribeGroupConfigBasic(name, description string) string {
	return fmt.Sprintf(`
resource "sendgrid_unsubscribe_group" "test" {
	name        = "%s"
	description = "%s"
	is_default  = false
}
`, name, description)
}

func testAccCheckSendgridUnsubscribeGroupConfigWithTimeouts(name, description string) string {
	return fmt.Sprintf(`
resource "sendgrid_unsubscribe_group" "rate_limit" {
	name        = "%s"
	description = "%s"
	is_default  = false

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}
`, name, description)
}

func testAccCheckSendgridUnsubscribeGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No unsubscribe group ID set")
		}

		config := testAccProvider.Meta().(*provider.Config)
	c := config.NewClient("")
		ctx := context.Background()

		_, err := c.ReadUnsubscribeGroup(ctx, rs.Primary.ID)
		if err.Err != nil {
			return fmt.Errorf("unsubscribe group not found: %s", rs.Primary.ID)
		}

		return nil
	}
}
