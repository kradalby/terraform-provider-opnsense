package opnsense

import (
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/kradalby/opnsense-go/opnsense"
	uuid "github.com/satori/go.uuid"
)

func resourceWireGuardClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceWireGuardClientCreate,
		Read:   resourceWireGuardClientRead,
		Update: resourceWireGuardClientUpdate,
		Delete: resourceWireGuardClientDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			// "uuid": {
			// 	Type:        schema.TypeString,
			// 	Description: "UUID assigned to client by OPNsense",
			// 	Computed:    true,
			// },
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Enable the client",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the client",
				Required:    true,
			},
			"tunnel_address": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of Tunnel addresses",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					Description:  "Tunnel address for the client, e.g. 10.0.0.1/32",
					ValidateFunc: validation.CIDRNetwork(0, 32),
				},
			},
			"public_key": {
				Type:        schema.TypeString,
				Description: "Public key of the client",
				Required:    true,
			},
			"shared_key": {
				Type:        schema.TypeString,
				Description: "Shared key of the client",
				Optional:    true,
			},
			"endpoint_address": {
				Type:         schema.TypeString,
				Description:  "IP or CNAME of remote endpoint",
				Optional:     true,
				ValidateFunc: validation.SingleIP(),
			},
			"endpoint_port": {
				Type:         schema.TypeInt,
				Description:  "Port of remote endpoint",
				Optional:     true,
				ValidateFunc: validation.IntBetween(10, 65535),
			},
			"keep_alive": {
				Type:         schema.TypeInt,
				Description:  "Connection keep alive",
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
		},
	}
}

func resourceWireGuardClientRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] Getting OPNsense client from meta")

	c := meta.(*opnsense.Client)

	log.Printf("[TRACE] Converting ID to UUID")

	uuid, err := uuid.FromString(d.Id())
	if err != nil {
		log.Printf("[ERROR] Failed to parse ID")

		return err
	}

	log.Printf("[TRACE] Fetching client configuration from OPNsense")

	client, err := c.WireGuardClientGet(uuid)
	if err != nil {
		if err.Error() == "found empty array, most likely 404" {
			d.SetId("")

			return nil
		}

		log.Printf("[DEBUG]: \n%#v", err)
		log.Printf("[ERROR] Failed to fetch uuid: %s", uuid)

		return err
	}

	log.Printf("[DEBUG] Configuration from OPNsense: \n")
	log.Printf("[DEBUG] %#v \n", client)

	d.Set("enabled", client.Enabled)
	d.Set("name", client.Name)
	d.Set("public_key", client.PubKey)
	d.Set("shared_key", client.Psk)
	d.Set("endpoint_address", client.ServerAddress)

	tunnelAddressList := opnsense.ListSelectedValues(client.TunnelAddress)
	d.Set("tunnel_address", tunnelAddressList)

	if client.ServerPort != "" {
		serverPort, err := strconv.Atoi(client.ServerPort)
		if err != nil {
			log.Printf("[ERROR] Failed to convert ServerPort to int: %s", client.ServerPort)

			return err
		}

		d.Set("endpoint_port", serverPort)
	}

	keepAlive, err := strconv.Atoi(client.KeepAlive)
	if err != nil {
		log.Printf("[ERROR] Failed to convert KeepAlive to int: %s", client.KeepAlive)

		return err
	}

	d.Set("keep_alive", keepAlive)

	return nil
}

func resourceWireGuardClientCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	client := opnsense.WireGuardClientSet{}

	err := prepareClientConfiguration(d, &client)
	if err != nil {
		return err
	}

	uuid, err := c.WireGuardClientAdd(client)
	if err != nil {
		return err
	}

	d.SetId(uuid.String())
	err = resourceWireGuardClientRead(d, meta)

	return err
}

func resourceWireGuardClientUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	uuid, err := uuid.FromString(d.Id())
	if err != nil {
		return err
	}

	client := opnsense.WireGuardClientSet{}

	err = prepareClientConfiguration(d, &client)
	if err != nil {
		return err
	}

	_, err = c.WireGuardClientSet(uuid, client)
	if err != nil {
		return err
	}

	d.SetId(uuid.String())
	err = resourceWireGuardClientRead(d, meta)

	return err
}

func resourceWireGuardClientDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	uuid, err := uuid.FromString(d.Id())
	if err != nil {
		return err
	}

	_, err = c.WireGuardClientDelete(uuid)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func prepareClientConfiguration(d *schema.ResourceData, client *opnsense.WireGuardClientSet) error {
	client.Enabled = d.Get("enabled").(opnsense.Bool)
	client.Name = d.Get("name").(string)
	client.PubKey = d.Get("public_key").(string)
	client.Psk = d.Get("shared_key").(string)
	client.ServerAddress = d.Get("endpoint_address").(string)

	if endpointPort := d.Get("endpoint_port").(int); endpointPort != 0 {
		log.Printf("[TRACE] ENDPOINT_PORT: %d", endpointPort)
		client.ServerPort = strconv.Itoa(endpointPort)
	}

	client.KeepAlive = strconv.Itoa(d.Get("keep_alive").(int))

	tunnelAddressList := d.Get("tunnel_address").(*schema.Set).List()
	tunnelAddressStringList := make([]string, len(tunnelAddressList))

	for index := range tunnelAddressList {
		tunnelAddressStringList[index] = tunnelAddressList[index].(string)
	}

	client.TunnelAddress = strings.Join(tunnelAddressStringList, ",")

	return nil
}
