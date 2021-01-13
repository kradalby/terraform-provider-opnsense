package opnsense

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/kradalby/opnsense-go/opnsense"
)

func testWireguardClientResource(name string) string {
	return fmt.Sprintf(`
// resource "opnsense_firmware" "%s" {
//     plugin {
//       name      = "os-wireguard"
//       installed = true
//     }
// }

resource "opnsense_wireguard_client" "%s" {
  enabled  = true
  name     = "tjoda"
  tunnel_address = ["10.10.10.0/24"]
  public_key    = "sDoPaHLw1efsq78fDaOtzPHmqAWnZImeKTfdJT3Cfk8="
  endpoint_port = 51820
  keep_alive    = 21
  shared_key    = ""
}
`, name, name)
}

func testAccWireguardClientResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*opnsense.Client)

	clients, err := c.WireGuardClientList()
	if err != nil {
		return err
	}

	if len(clients) != 0 {
		return fmt.Errorf("All clients are not removed, %d", len(clients))
	}

	return nil
}

func TestWireguardClient_basic(t *testing.T) {
	err := setupPlugins([]string{"os-wireguard"})
	if err != nil {
		t.Errorf("Setup failed, manual cleanup steps might be needed: %s", err)
	}

	rName := fmt.Sprintf("a%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccWireguardClientResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testWireguardClientResource(rName),
				Check: resource.ComposeTestCheckFunc(
					// resource.TestCheckResourceAttr(
					// 	fmt.Sprintf("opnsense_firmware.%s", rName),
					// 	"plugin.#",
					// 	"1",
					// ),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("opnsense_wireguard_client.%s", rName),
						"name",
						"tjoda",
					),
				),
			},
		},
	})

	err = tearDownPlugins([]string{"os-wireguard"})
	if err != nil {
		t.Errorf("Setup failed, manual cleanup steps might be needed: %s", err)
	}
}
