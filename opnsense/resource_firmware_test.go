package opnsense

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/kradalby/opnsense-go/opnsense"
)

func testFirmwarePluginResource(name string, plugins []string) string {
	return fmt.Sprintf(`
locals {
  plugins = ["%s"]
}

resource "opnsense_firmware" "%s" {
  dynamic "plugin" {
    for_each = local.plugins

    content {
      name      = plugin.value
      installed = true

    }
  }
}
`, strings.Join(plugins, `", "`), name)

}

func testAccFirmwarePluginResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*opnsense.Client)

	installedPlugins, err := c.FirmwareInstalledPluginsList()
	if err != nil {
		return err
	}

	if len(installedPlugins) != 0 {
		return fmt.Errorf("All plugins are not uninstalled, %d", len(installedPlugins))
	}

	return nil
}

func TestFirmwarePlugin_basic(t *testing.T) {
	rName := fmt.Sprintf("a%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccFirmwarePluginResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testFirmwarePluginResource(rName,
					[]string{"os-vmware"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						fmt.Sprintf("opnsense_firmware.%s", rName),
						"plugin.#",
						"1",
					),
					// resource.TestCheckResourceAttr(
					// 	fmt.Sprintf("opnsense_firmware.%s", rName),
					// 	"plugin.0.name",
					// 	"os-wireguard",
					// ),
				),
			},
			{
				Config: testFirmwarePluginResource(rName,
					[]string{"os-iperf", "os-wireguard", "os-firewall"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						fmt.Sprintf("opnsense_firmware.%s", rName),
						"plugin.#",
						"3",
					),
					// resource.TestCheckResourceAttr(
					// fmt.Sprintf("opnsense_firmware.%s", rName),
					// "plugin.0.name",
					// "os-iperf",
					// ),
					// resource.TestCheckResourceAttr(
					// fmt.Sprintf("opnsense_firmware.%s", rName),
					// "plugin.1.name",
					// "os-wireguard",
					// ),
					// resource.TestCheckResourceAttr(
					// fmt.Sprintf("opnsense_firmware.%s", rName),
					// "plugin.2.name",
					// "os-firewall",
					// ),
				),
			},
			{
				Config: testFirmwarePluginResource(rName,
					[]string{"os-iperf", "os-firewall"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						fmt.Sprintf("opnsense_firmware.%s", rName),
						"plugin.#",
						"2",
					),
					// resource.TestCheckResourceAttr(
					// 	fmt.Sprintf("opnsense_firmware.%s", rName),
					// 	"plugin.0.name",
					// 	"os-iperf",
					// ),
					// resource.TestCheckResourceAttr(
					// 	fmt.Sprintf("opnsense_firmware.%s", rName),
					// 	"plugin.1.name",
					// 	"os-firewall",
					// ),
				),
			},
			{
				Config: testFirmwarePluginResource(rName,
					[]string{"os-iperf", "os-wireguard"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						fmt.Sprintf("opnsense_firmware.%s", rName),
						"plugin.#",
						"2",
					),
					// resource.TestCheckResourceAttr(
					// 	fmt.Sprintf("opnsense_firmware.%s", rName),
					// 	"plugin.0.name",
					// 	"os-iperf",
					// ),
					// resource.TestCheckResourceAttr(
					// 	fmt.Sprintf("opnsense_firmware.%s", rName),
					// 	"plugin.1.name",
					// 	"os-wireguard",
					// ),
				),
			},
			{
				Config: testFirmwarePluginResource(rName,
					[]string{"os-wireguard"}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						fmt.Sprintf("opnsense_firmware.%s", rName),
						"plugin.#",
						"1",
					),
					// resource.TestCheckResourceAttr(
					// 	fmt.Sprintf("opnsense_firmware.%s", rName),
					// 	"plugin.0.name",
					// 	"os-wireguard",
					// ),
				),
			},
		},
	})
}
