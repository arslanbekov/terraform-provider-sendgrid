package sendgrid_test

import (
	"os"
	"testing"

	"github.com/arslanbekov/terraform-provider-sendgrid/sendgrid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviderFactories map[string]func() (*schema.Provider, error)

var testAccProvider *schema.Provider

func init() {
	testAccProvider = sendgrid.Provider()
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"sendgrid": func() (*schema.Provider, error) { return testAccProvider, nil },
	}
}

func TestProvider(t *testing.T) {
	if err := sendgrid.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	_ = sendgrid.Provider()
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	if err := os.Getenv("SENDGRID_API_KEY"); err == "" {
		t.Fatal("SENDGRID_API_KEY must be set for acceptance tests")
	}
}
