package opnsense

import (
	"github.com/cdeconinck/opnsense-go/opnsense"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/satori/go.uuid"
	"log"
	"strings"
)

func resourceFirewallAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallAliasCreate,
		Read:   resourceFirewallAliasRead,
		Update: resourceFirewallAliasUpdate,
		Delete: resourceFirewallAliasDelete,

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
				Description: "Enable the alias",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the alias",
				Required:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of the alias",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the alias",
				Optional:    true,
			},
			"content": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			// TODO add other fields (like proto)
		},
	}
}

func resourceFirewallAliasRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] Getting OPNsense client from meta")
	c := meta.(*opnsense.Client)

	log.Printf("[TRACE] Converting ID to UUID")
	uuid, err := uuid.FromString(d.Id())
	if err != nil {
		log.Printf("[ERROR] Failed to parse ID")
		return err
	}

	log.Printf("[TRACE] Fetching client configuration from OPNsense")
	client, err := c.AliasGet(uuid)
	if err != nil {
		// temporary fix for the internal error API when we try to get an unreferenced UIID
		if err.Error() == "Internal Error status code received" {
			d.SetId("")
			return nil
		}
		log.Printf("ERROR: \n%#v", err)
		log.Printf("[ERROR] Failed to fetch uuid: %s", uuid)
		return err
	}

	log.Printf("[DEBUG] Configuration from OPNsense: \n")
	log.Printf("[DEBUG] %#v \n", client)

	if client.Enabled == "1" {
		d.Set("enabled", true)
	} else {
		d.Set("enabled", false)
	}
	d.Set("name", client.Name)
	d.Set("type", client.Type)
	d.Set("description", client.Description)

	content := make([]string, 0)
	if client.Content != nil {
		for _, v := range client.Content {
			if v.Selected == 1 {
				content = append(content, v.Value)
			}
		}
	}

	d.Set("content", content)

	return nil
}

func resourceFirewallAliasCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)
	client := opnsense.AliasSet{}

	err := prepareFirewallAliasConfiguration(d, &client)
	if err != nil {
		return err
	}

	uuid, err := c.AliasAdd(client)
	if err != nil {
		return err
	}

	d.SetId(uuid.String())
	resourceFirewallAliasRead(d, meta)

	return nil
}

func resourceFirewallAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	uuid, err := uuid.FromString(d.Id())
	if err != nil {
		return err
	}

	client := opnsense.AliasSet{}

	err = prepareFirewallAliasConfiguration(d, &client)
	if err != nil {
		return err
	}

	_, err = c.AliasUpdate(uuid, client)
	if err != nil {
		return err
	}

	d.SetId(uuid.String())
	resourceFirewallAliasRead(d, meta)

	return nil
}

func resourceFirewallAliasDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	uuid, err := uuid.FromString(d.Id())
	if err != nil {
		return err
	}

	_, err = c.AliasDelete(uuid)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func prepareFirewallAliasConfiguration(d *schema.ResourceData, client *opnsense.AliasSet) error {
	if d.Get("enabled").(bool) {
		client.Enabled = "1"
	} else {
		client.Enabled = "0"
	}
	client.Name = d.Get("name").(string)
	client.Description = d.Get("description").(string)
	client.Type = d.Get("type").(string)
	content_list := d.Get("content").(*schema.Set).List()

	content_list_str := make([]string, len(content_list))
	for i := range content_list {
		content_list_str[i] = content_list[i].(string)
	}
	client.Content = strings.Join(content_list_str, "\n")

	return nil
}