package opnsense

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			// TODO add other fields
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
				log.Printf("[ERROR] dataFirewallAliasRead - Failed to parse ID")

				return err
			}

			wantedAlias, err := c.AliasGet(wantedUUID)
			if err != nil {
				return err
			}

			d.SetId(wantedAlias.UUID.String())

			err = d.Set("Name", wantedAlias.Name)
			if err != nil {
				return err
			}

			err = d.Set("Enabled", wantedAlias.Enabled)
			if err != nil {
				return err
			}

			err = d.Set("Description", wantedAlias.Description)
			if err != nil {
				return err
			}

			err = d.Set("Type", wantedAlias.Type)
			if err != nil {
				return err
			}

			err = d.Set("Content", wantedAlias.Content)
			if err != nil {
				return err
			}

			break
		}
	}

	return nil
}
