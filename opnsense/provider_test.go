package opnsense

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kradalby/opnsense-go/opnsense"
)

var (
	testAccProviders map[string]*schema.Provider
	testAccProvider  *schema.Provider
)

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"opnsense": testAccProvider,
	}
}

func setupPlugins(plugins []string) error {
	c, err := opnsense.NewClient("", "", "", true)
	if err != nil {
		return err
	}

	for _, plugin := range plugins {
		err := c.FirmwareInstall(plugin)
		if err != nil {
			return err
		}
	}

	return nil
}

func tearDownPlugins(plugins []string) error {
	c, err := opnsense.NewClient("", "", "", true)
	if err != nil {
		return err
	}

	for _, plugin := range plugins {
		err := c.FirmwareRemove(plugin)
		if err != nil {
			return err
		}
	}

	return nil
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("OPNSENSE_URL"); v == "" {
		t.Fatal("OPNSENSE_URL must be set for acceptance tests")
	}

	if v := os.Getenv("OPNSENSE_KEY"); v == "" {
		t.Fatal("OPNSENSE_KEY must be set for acceptance tests")
	}

	if v := os.Getenv("OPNSENSE_SECRET"); v == "" {
		t.Fatal("OPNSENSE_SECRET must be set for acceptance tests")
	}
}
