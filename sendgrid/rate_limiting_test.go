package sendgrid_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSendgridRateLimitingAPIKey(t *testing.T) {
	// Test creating multiple API keys in sequence to trigger rate limiting
	keys := make([]string, 5)
	for i := 0; i < 5; i++ {
		keys[i] = "terraform-rate-test-" + acctest.RandString(10)
	}

	var configs []string
	for i, name := range keys {
		configs = append(configs, fmt.Sprintf(`
resource "sendgrid_api_key" "rate_test_%d" {
	name   = "%s"
	scopes = ["mail.send"]

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}`, i, name))
	}

	// Concatenate all configs
	allConfigs := ""
	for _, cfg := range configs {
		allConfigs += cfg
	}

	config := fmt.Sprintf(`
%s

# Output all API key IDs
output "api_key_ids" {
	value = [%s]
}
`,
		allConfigs,
		"sendgrid_api_key.rate_test_0.id, sendgrid_api_key.rate_test_1.id, sendgrid_api_key.rate_test_2.id, sendgrid_api_key.rate_test_3.id, sendgrid_api_key.rate_test_4.id")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_api_key.rate_test_0", "name", keys[0]),
					resource.TestCheckResourceAttr("sendgrid_api_key.rate_test_1", "name", keys[1]),
					resource.TestCheckResourceAttr("sendgrid_api_key.rate_test_2", "name", keys[2]),
					resource.TestCheckResourceAttr("sendgrid_api_key.rate_test_3", "name", keys[3]),
					resource.TestCheckResourceAttr("sendgrid_api_key.rate_test_4", "name", keys[4]),
				),
			},
		},
	})
}

func TestAccSendgridRateLimitingTemplate(t *testing.T) {
	// Test creating multiple templates to trigger rate limiting
	templates := make([]string, 3)
	for i := 0; i < 3; i++ {
		templates[i] = "terraform-template-rate-test-" + acctest.RandString(10)
	}

	config := fmt.Sprintf(`
resource "sendgrid_template" "rate_test_0" {
	name       = "%s"
	generation = "dynamic"

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_template" "rate_test_1" {
	name       = "%s"
	generation = "dynamic"

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_template" "rate_test_2" {
	name       = "%s"
	generation = "dynamic"

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}
`, templates[0], templates[1], templates[2])

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_template.rate_test_0", "name", templates[0]),
					resource.TestCheckResourceAttr("sendgrid_template.rate_test_1", "name", templates[1]),
					resource.TestCheckResourceAttr("sendgrid_template.rate_test_2", "name", templates[2]),
				),
			},
		},
	})
}

func TestAccSendgridRateLimitingTeammate(t *testing.T) {
	// Test creating multiple teammates to trigger rate limiting
	emails := make([]string, 3)
	for i := 0; i < 3; i++ {
		emails[i] = "terraform-teammate-rate-test-" + acctest.RandString(10) + "@example.com"
	}

	config := fmt.Sprintf(`
resource "sendgrid_teammate" "rate_test_0" {
	email    = "%s"
	is_admin = false
	is_sso   = false
	scopes   = ["mail.send"]

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_teammate" "rate_test_1" {
	email    = "%s"
	is_admin = false
	is_sso   = false
	scopes   = ["mail.send"]

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}

resource "sendgrid_teammate" "rate_test_2" {
	email    = "%s"
	is_admin = false
	is_sso   = false
	scopes   = ["mail.send"]

	timeouts {
		create = "30m"
		update = "30m"
		delete = "30m"
	}
}
`, emails[0], emails[1], emails[2])

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sendgrid_teammate.rate_test_0", "email", emails[0]),
					resource.TestCheckResourceAttr("sendgrid_teammate.rate_test_1", "email", emails[1]),
					resource.TestCheckResourceAttr("sendgrid_teammate.rate_test_2", "email", emails[2]),
				),
			},
		},
	})
}
