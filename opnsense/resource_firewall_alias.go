package opnsense

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/kradalby/opnsense-go/opnsense"
	uuid "github.com/satori/go.uuid"
)

func resourceFirewallAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallAliasCreate,
		Read:   resourceFirewallAliasRead,
		Update: resourceFirewallAliasUpdate,
		Delete: resourceFirewallAliasDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"parent": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "UIID of other alias who will contain this alias (nested)",
				Optional:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "State of the alias",
				Optional:    true,
				Default:     true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the alias",
				Required:    true,
			},
			"type": {
				Type:         schema.TypeString,
				Description:  "Type of the alias",
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"host", "network", "port", "url"}, false),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the alias",
				Optional:    true,
				Default:     "",
			},
			"content": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The content of this alias (IP, Cidr, url, ...)",

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

	log.Printf("[TRACE] Fetching alias configuration from OPNsense")

	alias, err := c.AliasGet(uuid)
	if err != nil {
		// temporary fix for the internal error API when we try to get an unreferenced UIID
		if err.Error() == apiInternalErrorMsg {
			d.SetId("")

			return nil
		}

		log.Printf("[ERROR] Failed to fetch uuid: %s", uuid)

		return err
	}

	log.Printf("[DEBUG] Configuration from OPNsense: \n")
	log.Printf("[DEBUG] %#v \n", alias)

	d.SetId(alias.UUID.String())

	err = d.Set("enabled", alias.Enabled)
	if err != nil {
		return err
	}

	err = d.Set("name", alias.Name)
	if err != nil {
		return err
	}

	err = d.Set("type", alias.Type)
	if err != nil {
		return err
	}

	err = d.Set("description", alias.Description)
	if err != nil {
		return err
	}

	err = d.Set("content", alias.Content)
	if err != nil {
		return err
	}

	parents := []string{}

	// check if this alias is a member of another alias (nested)
	aliasList, err := c.AliasGetList()
	if err != nil {
		log.Printf("[ERROR]: %v", err)

		return err
	}

	for _, nestedAlias := range aliasList.Rows {
		if strings.Contains(nestedAlias.Content, alias.Name) {
			parents = append(parents, nestedAlias.UUID)
		}
	}

	err = d.Set("parent", parents)
	if err != nil {
		return err
	}

	return nil
}

func resourceFirewallAliasCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)
	alias := opnsense.AliasFormat{}

	err := prepareFirewallAliasConfiguration(d, &alias)
	if err != nil {
		return err
	}

	// create the alias
	createdUUID, err := c.AliasAdd(alias)
	if err != nil {
		return err
	}

	// add the alias to his parent if necessary
	parent := d.Get("parent")
	if parent != nil {
		parentList := parent.(*schema.Set).List()
		if len(parentList) > 0 {
			err = addNestedAlias(c, parentList, alias.Name)
			if err != nil {
				return err
			}
		}
	}

	// apply configuration change
	_, err = c.AliasReconfigure()
	if err != nil {
		return err
	}

	d.SetId(createdUUID.String())
	err = resourceFirewallAliasRead(d, meta)

	return err
}

func resourceFirewallAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	// TODO don"t update the alias if only the parent field is modified
	c := meta.(*opnsense.Client)

	elmUUID, err := uuid.FromString(d.Id())
	if err != nil {
		return err
	}

	alias := opnsense.AliasFormat{}

	err = prepareFirewallAliasConfiguration(d, &alias)
	if err != nil {
		return err
	}

	_, err = c.AliasUpdate(elmUUID, alias)
	if err != nil {
		return err
	}

	if d.HasChange("parent") {
		oldParent, newParent := d.GetChange("parent")
		log.Println("[TRACE] OLD Parent : ", oldParent)
		log.Println("[TRACE] NEW Parent : ", newParent)

		oldParentSet := oldParent.(*schema.Set)
		newParentSet := newParent.(*schema.Set)

		listToDel := oldParentSet.Difference(newParentSet).List()
		listToAdd := newParentSet.Difference(oldParentSet).List()

		// remove this alias from the previous nested alias
		if len(listToDel) > 0 {
			err = removeNestedAlias(c, listToDel, alias.Name)
			if err != nil {
				return err
			}
		}

		if len(listToAdd) > 0 {
			err = addNestedAlias(c, listToAdd, alias.Name)
			if err != nil {
				return err
			}
		}
	}

	// apply configuration change
	_, err = c.AliasReconfigure()
	if err != nil {
		return err
	}

	d.SetId(elmUUID.String())

	err = resourceFirewallAliasRead(d, meta)

	return err
}

func resourceFirewallAliasDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	uuid, err := uuid.FromString(d.Id())
	if err != nil {
		return err
	}

	// if this alias is nested, we need to delete this ressource in the parent before deleting this alias
	parent := d.Get("parent")
	if parent != nil {
		parentList := parent.(*schema.Set).List()
		if len(parentList) > 0 {
			err = removeNestedAlias(c, parentList, d.Get("name").(string))
			if err != nil {
				return err
			}
		}
	}

	_, err = c.AliasDelete(uuid)
	if err != nil {
		return err
	}

	// apply configuration change
	_, err = c.AliasReconfigure()
	if err != nil {
		return err
	}

	d.SetId("")

	return err
}

func prepareFirewallAliasConfiguration(d *schema.ResourceData, conf *opnsense.AliasFormat) error {
	conf.Enabled = d.Get("enabled").(bool)
	conf.Name = d.Get("name").(string)
	conf.Description = d.Get("description").(string)
	conf.Type = d.Get("type").(string)

	contentList := d.Get("content").(*schema.Set).List()
	contentListStr := make([]string, len(contentList))

	for i := range contentList {
		contentListStr[i] = contentList[i].(string)
	}

	conf.Content = contentListStr

	return nil
}

func removeInList(slice []string, elm string) ([]string, bool) {
	for k, v := range slice {
		if v == elm {
			return append(slice[:k], slice[k+1:]...), true
		}
	}

	return slice, false
}

func removeNestedAlias(c *opnsense.Client, parentUUIDList []interface{}, name string) error {
	for _, parentUUIDStr := range parentUUIDList {
		parentUUID, err := uuid.FromString(parentUUIDStr.(string))
		if err != nil {
			return fmt.Errorf("[ERROR] Failed to parse ID: %w", err)
		}

		parentAlias, err := c.AliasGet(parentUUID)
		if err != nil {
			return fmt.Errorf("[ERROR] Something went wrong while retrieving parent alias for: %w", err)
		}

		parentAlias.Content, _ = removeInList(parentAlias.Content, name)

		_, err = c.AliasUpdate(parentUUID, *parentAlias)
		if err != nil {
			return err
		}
	}

	return nil
}

func addNestedAlias(c *opnsense.Client, parentUUIDList []interface{}, name string) error {
	for _, parentUUIDStr := range parentUUIDList {
		parentUUID, err := uuid.FromString(parentUUIDStr.(string))
		if err != nil {
			return fmt.Errorf("[ERROR] Failed to parse ID: %w", err)
		}

		parentAlias, err := c.AliasGet(parentUUID)
		if err != nil {
			return fmt.Errorf("[ERROR] Something went wrong while retrieving parent alias for: %w", err)
		}

		parentAlias.Content = append(parentAlias.Content, name)
		_, err = c.AliasUpdate(parentUUID, *parentAlias)

		if err != nil {
			return err
		}
	}

	return nil
}
