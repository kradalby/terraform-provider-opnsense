package opnsense

import (
	"github.com/cdeconinck/opnsense-go/opnsense"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/satori/go.uuid"
	"log"
)

func dataFirewallAlias() *schema.Resource {
	return &schema.Resource{
		Read: dataFirewallAliasRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			//TODO add other fields
		},
	}
}

// Read will fetch the data of a resource.
func dataFirewallAliasRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] Getting OPNsense client from meta")
	c := meta.(*opnsense.Client)
	wanted_name := d.Get("name")

	// list all alias
	alias_list, err_list := c.AliasGetList()
	if err_list != nil {
		// temporary fix for the internal error API when we try to get an unreferenced UIID
		if err_list.Error() == "Internal Error status code received" {
			d.SetId("")
			return nil
		}
		log.Printf("ERROR: \n%#v", err_list)
		return err_list
	}

	for _, alias := range *alias_list {
		if alias.Name == wanted_name {
			wanted_uuid, err_uuid := uuid.FromString(alias.UUID)
			if err_uuid != nil {
				log.Printf("[ERROR]dataFirewallAliasRead -  Failed to parse ID")
				return err_uuid
			}

			wanted_alias, err_get := c.AliasGet(wanted_uuid)
			if err_get != nil {
				return err_get
			}

			d.SetId(wanted_alias.UUID.String())
			d.Set("Name", wanted_alias.Name)
			d.Set("Enabled", wanted_alias.Enabled)
			d.Set("Description", wanted_alias.Description)
			d.Set("Type", wanted_alias.Type)
			d.Set("Content", wanted_alias.Content)
			break
		}
	}

	return nil
}
