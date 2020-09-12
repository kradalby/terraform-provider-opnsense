package opnsense

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"opnsense": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("OPNSENSE_ADDRESS"); v == "" {
		t.Fatal("OPNSENSE_ADDRESS must be set for acceptance tests")
	}
	if v := os.Getenv("OPNSENSE_KEY"); v == "" {
		t.Fatal("OPNSENSE_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("OPNSENSE_SECRET"); v == "" {
		t.Fatal("OPNSENSE_SECRET must be set for acceptance tests")
	}
}
