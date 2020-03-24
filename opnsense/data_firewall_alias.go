package opnsense

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/kradalby/opnsense-go/opnsense"
	uuid "github.com/satori/go.uuid"
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

	wantedName := d.Get("name")

	// list all alias
	aliasList, err := c.AliasGetList()
	if err != nil {
		// temporary fix for the internal error API when we try to get an unreferenced UIID
		if err.Error() == apiInternalErrorMsg {
			d.SetId("")
			return nil
		}

		log.Printf("ERROR: \n%#v", err)

		return err
	}

	for _, alias := range aliasList.Rows {
		if alias.Name == wantedName {
			wantedUUID, err := uuid.FromString(alias.UUID)
			if err != nil {
				log.Printf("[ERROR] dataFirewallAliasRead -  Failed to parse ID")
				return err
			}

			wantedAlias, err := c.AliasGet(wantedUUID)
			if err != nil {
				return err
			}

			d.SetId(wantedAlias.UUID.String())
			d.Set("Name", wantedAlias.Name)
			d.Set("Enabled", wantedAlias.Enabled)
			d.Set("Description", wantedAlias.Description)
			d.Set("Type", wantedAlias.Type)
			d.Set("Content", wantedAlias.Content)

			break
		}
	}

	return nil
}
