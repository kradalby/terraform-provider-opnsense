package opnsense

import (
	//"fmt"
	"github.com/cdeconinck/opnsense-go/opnsense"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/satori/go.uuid"
	"log"
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
			"parent": {
				Type:        schema.TypeString,
				Description: "Name assigned to the parent alias",
				Optional:    true,
			},
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
		log.Printf("[ERROR]resourceFirewallAliasRead -  Failed to parse ID")
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

	d.Set("enabled", client.Enabled)
	d.Set("name", client.Name)
	d.Set("type", client.Type)
	d.Set("description", client.Description)
	d.Set("content", client.Content)

	return nil
}

func resourceFirewallAliasCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)
	client := opnsense.AliasFormat{}

	err := prepareFirewallAliasConfiguration(d, &client)
	if err != nil {
		return err
	}

	// create the alias
	uuid_created, err := c.AliasAdd(client)
	if err != nil {
		return err
	}

	// add this alias to the parent if set
	/*parent := d.Get("parent")
	if parent != nil {
		parent_uuid, err := uuid.FromString(parent.(string))
		if err != nil {
			return fmt.Errorf("[ERROR] Failed to parse ID")
		}

		parent_alias, err_get := c.AliasGet(parent_uuid)
		if err_get != nil {
			return fmt.Errorf("Something went wrong while retrieving parent alias for: %s", err_get)
		}

		parent_alias.Content = append(parent_alias.Content, client.Name)

		_, err_update := c.AliasUpdate(parent_uuid, *parent_alias)

		if err_update != nil {
			log.Println(err_update)
		}
	}*/

	d.SetId(uuid_created.String())
	resourceFirewallAliasRead(d, meta)

	return nil
}

func resourceFirewallAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	uuid, err := uuid.FromString(d.Id())
	if err != nil {
		return err
	}

	client := opnsense.AliasFormat{}

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

func prepareFirewallAliasConfiguration(d *schema.ResourceData, client *opnsense.AliasFormat) error {
	client.Enabled = d.Get("enabled").(bool)
	client.Name = d.Get("name").(string)
	client.Description = d.Get("description").(string)
	client.Type = d.Get("type").(string)

	content_list := d.Get("content").(*schema.Set).List()
	content_list_str := make([]string, len(content_list))
	for i := range content_list {
		content_list_str[i] = content_list[i].(string)
	}
	client.Content = content_list_str

	return nil
}
