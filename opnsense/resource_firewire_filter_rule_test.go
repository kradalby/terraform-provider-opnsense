package opnsense

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/kradalby/opnsense-go/opnsense"
)

func testFirewallFilterRuleResource(name string) string {
	return fmt.Sprintf(`
resource "opnsense_firmware" "%s" {
    plugin {
      name      = "os-firewall"
      installed = true
    }
}

resource "opnsense_firewall_filter_rule" "%s" {
	enabled = true
	action = "pass"
	quick = true
	interface = "wan"
	source_net = "192.168.0.0/24"
	source_port = 8000
	destination_net = "192.168.0.0/24"
	destination_port = 8000
}
`, name, name)
}

func testAccFirewallFilterRuleResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*opnsense.Client)

	rules, err := c.FirewallFilterRuleSearch()
	if err != nil {
		return err
	}

	if len(rules) != 0 {
		return fmt.Errorf("All plugins are not uninstalled, %d", len(rules))
	}

	return nil
}

func TestFirewallFilterRule_basic(t *testing.T) {
	rName := fmt.Sprintf("a%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccFirewallFilterRuleResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testFirewallFilterRuleResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						fmt.Sprintf("opnsense_firmware.%s", rName),
						"plugin.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("opnsense_firewall_filter_rule.%s", rName),
						"action",
						"pass",
					),
				),
			},
		},
	})
}
