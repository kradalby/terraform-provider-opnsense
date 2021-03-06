package opnsense

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/kradalby/opnsense-go/opnsense"
	uuid "github.com/satori/go.uuid"
)

func resourceWireGuardServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceWireGuardServerCreate,
		Read:   resourceWireGuardServerRead,
		Update: resourceWireGuardServerUpdate,
		Delete: resourceWireGuardServerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Enable the server",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the server",
				Required:    true,
			},
			"public_key": {
				Type:        schema.TypeString,
				Description: "Public key of the server",
				Computed:    true,
			},
			"private_key": {
				Type:        schema.TypeString,
				Description: "Public key of the server",
				Computed:    true,
				Sensitive:   true,
			},
			"port": {
				Type:         schema.TypeInt,
				Description:  "Listening port for WireGuard server",
				Required:     true,
				ValidateFunc: validation.IntBetween(10, 65535),
			},
			"mtu": {
				Type: schema.TypeInt,
				Description: "Set the interface MTU for this interface. Leaving empty uses the " +
					"MTU from main interface which is fine for most setups.",
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 16384),
			},
			"disable_routes": {
				Type:        schema.TypeBool,
				Description: "Prevent WireGuard from adding routes",
				Required:    true,
			},
			"tunnel_address": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of Tunnel addresses",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					Description:  "Tunnel address for the server, e.g. 10.0.0.1/32",
					ValidateFunc: validation.IsCIDRNetwork(0, 32),
				},
			},
			"dns": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of DNS addresses",
				Elem: &schema.Schema{
					Type:        schema.TypeString,
					Description: "DNS address, e.g. 10.0.0.1",
				},
			},
			"peers": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of UUIDs for clients",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					Description:  "UUIDs for clients",
					ValidateFunc: validation.IsUUID,
				},
			},
		},
	}
}

func resourceWireGuardServerRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] Getting OPNsense client from meta")

	c := meta.(*opnsense.Client)

	log.Printf("[TRACE] Converting ID to UUID")

	uuid, err := uuid.FromString(d.Id())
	if err != nil {
		log.Printf("[ERROR] Failed to parse ID")

		return err
	}

	log.Printf("[TRACE] Fetching server configuration from OPNsense")

	server, err := c.WireGuardServerGet(uuid)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch uuid: %s", uuid)

		return err
	}

	log.Printf("[DEBUG] Configuration from OPNsense: \n")
	log.Printf("[DEBUG] %#v \n", server)

	err = d.Set("enabled", server.Enabled)
	if err != nil {
		return err
	}

	err = d.Set("name", server.Name)
	if err != nil {
		return err
	}

	err = d.Set("public_key", server.PubKey)
	if err != nil {
		return err
	}

	err = d.Set("private_key", server.PrivKey)
	if err != nil {
		return err
	}

	err = d.Set("disable_routes", server.DisableRoutes)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(server.Port)
	if err != nil {
		log.Printf("[ERROR] Failed to convert ServerPort to int: %s", server.Port)

		return err
	}

	err = d.Set("port", port)
	if err != nil {
		return err
	}

	if server.MTU != "" {
		mtu, err := strconv.Atoi(server.MTU)
		if err != nil {
			log.Printf("[ERROR] Failed to convert MTU to int: %s", server.MTU)

			return err
		}

		err = d.Set("mtu", mtu)
		if err != nil {
			return err
		}
	}

	// TODO: Handle this map[string]interface
	// d.Set("tunnel_address", server.TunnelAddress)
	if server.TunnelAddress != nil {
		tunnelAddressList := opnsense.ListSelectedValues(server.TunnelAddress)

		err = d.Set("tunnel_address", tunnelAddressList)
		if err != nil {
			return err
		}
	}

	if server.DNS != nil {
		dnsAddressList := opnsense.ListSelectedValues(server.DNS)

		err = d.Set("dns", dnsAddressList)
		if err != nil {
			return err
		}
	}

	if server.Peers != nil {
		peerList := opnsense.ListSelectedKeys(server.Peers)

		err = d.Set("peers", peerList)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceWireGuardServerCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	server := opnsense.WireGuardServerSet{}

	err := prepareServerConfiguration(d, &server)
	if err != nil {
		return err
	}

	err = c.WireGuardServerAdd(server)
	if err != nil {
		return err
	}

	uuids, err := c.WireGuardServerFindUUIDByName(server.Name)
	if err != nil {
		return err
	}

	if len(uuids) != 1 {
		err := fmt.Errorf(
			"server returned %d UUIDs for the given server name %w",
			len(uuids), ErrMoreThanOneUUIDReturned,
		)
		log.Printf("[ERROR] %#v", err)

		return err
	}

	d.SetId(uuids[0].String())

	err = resourceWireGuardServerRead(d, meta)

	return err
}

func resourceWireGuardServerUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	uuid, err := uuid.FromString(d.Id())
	if err != nil {
		return err
	}

	server := opnsense.WireGuardServerSet{}

	err = prepareServerConfiguration(d, &server)
	if err != nil {
		return err
	}

	_, err = c.WireGuardServerSet(uuid, server)
	if err != nil {
		return err
	}

	d.SetId(uuid.String())
	err = resourceWireGuardServerRead(d, meta)

	return err
}

func resourceWireGuardServerDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	uuid, err := uuid.FromString(d.Id())
	if err != nil {
		return err
	}

	_, err = c.WireGuardServerDelete(uuid)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func prepareServerConfiguration(d *schema.ResourceData, server *opnsense.WireGuardServerSet) error {
	server.Enabled = opnsense.Bool(d.Get("enabled").(bool))
	server.Name = d.Get("name").(string)
	server.PubKey = d.Get("public_key").(string)
	server.PrivKey = d.Get("private_key").(string)
	server.DisableRoutes = opnsense.Bool(d.Get("disable_routes").(bool))

	server.Port = strconv.Itoa(d.Get("port").(int))

	if d.Get("MTU") != nil {
		server.MTU = strconv.Itoa(d.Get("MTU").(int))
	}

	tunnelAddressList := d.Get("tunnel_address").(*schema.Set).List()
	tunnelAddressStringList := make([]string, len(tunnelAddressList))

	for index := range tunnelAddressList {
		tunnelAddressStringList[index] = tunnelAddressList[index].(string)
	}

	server.TunnelAddress = strings.Join(tunnelAddressStringList, ",")

	dnsAddressList := d.Get("dns").(*schema.Set).List()
	dnsAddressStringList := make([]string, len(dnsAddressList))

	for index := range dnsAddressList {
		dnsAddressStringList[index] = dnsAddressList[index].(string)
	}

	server.DNS = strings.Join(dnsAddressStringList, ",")

	peerList := d.Get("peers").(*schema.Set).List()
	peerStringList := make([]string, len(peerList))

	for index := range peerList {
		peerStringList[index] = peerList[index].(string)
	}

	server.Peers = strings.Join(peerStringList, ",")

	return nil
}
