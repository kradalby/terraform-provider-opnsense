package opnsense

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/kradalby/opnsense-go/opnsense"
)

func testWireguardServerResource(name string) string {
	return fmt.Sprintf(`
resource "opnsense_firmware" "%s" {
    plugin {
      name      = "os-wireguard"
      installed = true
    }
}

resource "opnsense_wireguard_server" "%s" {
  enabled  = true
  name     = "tjoda"
  tunnel_address = ["10.10.10.0/24"]
  port = 51820
  disable_routes = false
  dns = ["1.1.1.1"]
  peers = []
}
`, name, name)
}

func testAccWireguardServerResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*opnsense.Client)

	servers, err := c.WireGuardServerList()
	if err != nil {
		return err
	}

	if len(servers) != 0 {
		return fmt.Errorf("All servers are not removed, %d", len(servers))
	}

	return nil
}

func TestWireguardServer_basic(t *testing.T) {
	rName := fmt.Sprintf("a%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccWireguardServerResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWireguardServerResource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						fmt.Sprintf("opnsense_firmware.%s", rName),
						"plugin.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("opnsense_wireguard_server.%s", rName),
						"name",
						"tjoda",
					),
				),
			},
		},
	})
}
