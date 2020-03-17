package opnsense

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/kradalby/opnsense-go/opnsense"
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
		if err.Error() == "Internal Error status code received" {
			d.SetId("")
			return nil
		}
		log.Printf("ERROR: \n%#v", err)
		log.Printf("[ERROR] Failed to fetch uuid: %s", uuid)
		return err
	}

	log.Printf("[DEBUG] Configuration from OPNsense: \n")
	log.Printf("[DEBUG] %#v \n", alias)

	d.SetId(alias.UUID.String())
	d.Set("enabled", alias.Enabled)
	d.Set("name", alias.Name)
	d.Set("type", alias.Type)
	d.Set("description", alias.Description)
	d.Set("content", alias.Content)
	parents := []string{}

	// check if this alias is a member of another alias (nested)
	alias_list, err := c.AliasGetList()
	if err != nil {
		log.Printf("[ERROR]: %v", err)
		return err
	}

	for _, nested_alias := range alias_list.Rows {
		if strings.Contains(nested_alias.Content, alias.Name) {
			parents = append(parents, nested_alias.UUID)
		}
	}

	d.Set("parent", parents)

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
	uuid_created, err := c.AliasAdd(alias)
	if err != nil {
		return err
	}

	// add the alias to his parent if necessary
	parent := d.Get("parent")
	if parent != nil {
		parent_list := parent.(*schema.Set).List()
		if len(parent_list) > 0 {
			addNestedAlias(c, parent_list, alias.Name)
		}
	}

	// apply configuration change
	_, err_apply := c.AliasReconfigure()
	if err_apply != nil {
		return err_apply
	}

	d.SetId(uuid_created.String())
	resourceFirewallAliasRead(d, meta)

	return nil
}

func resourceFirewallAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	// TODO don"t update the alias if only the parent field is modified
	c := meta.(*opnsense.Client)

	elm_uuid, err := uuid.FromString(d.Id())
	if err != nil {
		return err
	}

	alias := opnsense.AliasFormat{}

	err = prepareFirewallAliasConfiguration(d, &alias)
	if err != nil {
		return err
	}

	_, err = c.AliasUpdate(elm_uuid, alias)
	if err != nil {
		return err
	}

	if d.HasChange("parent") {
		old_parent, new_parent := d.GetChange("parent")
		log.Println("[ERROR] OLD Parent : ", old_parent)
		log.Println("[ERROR] NEW Parent : ", new_parent)

		old_parent_set := old_parent.(*schema.Set)
		new_parent_set := new_parent.(*schema.Set)

		list_to_del := old_parent_set.Difference(new_parent_set).List()
		list_to_add := new_parent_set.Difference(old_parent_set).List()

		// remove this alias from the previous nested alias
		if len(list_to_del) > 0 {
			removeNestedAlias(c, list_to_del, alias.Name)
		}

		if len(list_to_add) > 0 {
			addNestedAlias(c, list_to_add, alias.Name)
		}
	}

	// apply configuration change
	_, err_apply := c.AliasReconfigure()
	if err_apply != nil {
		return err_apply
	}

	d.SetId(elm_uuid.String())
	resourceFirewallAliasRead(d, meta)

	return nil
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
		alias_name := d.Get("name").(string)
		parent_list := parent.(*schema.Set).List()
		if len(parent_list) > 0 {
			removeNestedAlias(c, parent_list, alias_name)
		}
	}

	_, err_del := c.AliasDelete(uuid)
	if err_del != nil {
		return err_del
	}

	// apply configuration change
	_, err_apply := c.AliasReconfigure()
	if err_apply != nil {
		return err_apply
	}

	d.SetId("")

	return nil
}

func prepareFirewallAliasConfiguration(d *schema.ResourceData, conf *opnsense.AliasFormat) error {
	conf.Enabled = d.Get("enabled").(bool)
	conf.Name = d.Get("name").(string)
	conf.Description = d.Get("description").(string)
	conf.Type = d.Get("type").(string)

	content_list := d.Get("content").(*schema.Set).List()
	content_list_str := make([]string, len(content_list))
	for i := range content_list {
		content_list_str[i] = content_list[i].(string)
	}
	conf.Content = content_list_str

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

func removeNestedAlias(c *opnsense.Client, parent_uuid_list []interface{}, name string) error {
	for _, parent_uuid_str := range parent_uuid_list {
		parent_uuid, err := uuid.FromString(parent_uuid_str.(string))
		if err != nil {
			return fmt.Errorf("[ERROR] Failed to parse ID")
		}

		parent_alias, err_get := c.AliasGet(parent_uuid)
		if err_get != nil {
			return fmt.Errorf("Something went wrong while retrieving parent alias for: %s", err_get)
		}

		parent_alias.Content, _ = removeInList(parent_alias.Content, name)
		_, err_update := c.AliasUpdate(parent_uuid, *parent_alias)

		if err_update != nil {
			return err_update
		}
	}

	return nil
}

func addNestedAlias(c *opnsense.Client, parent_uuid_list []interface{}, name string) error {
	for _, parent_uuid_str := range parent_uuid_list {
		parent_uuid, err := uuid.FromString(parent_uuid_str.(string))
		if err != nil {
			return fmt.Errorf("[ERROR] Failed to parse ID")
		}

		parent_alias, err_get := c.AliasGet(parent_uuid)
		if err_get != nil {
			return fmt.Errorf("Something went wrong while retrieving parent alias for: %s", err_get)
		}

		parent_alias.Content = append(parent_alias.Content, name)
		_, err_update := c.AliasUpdate(parent_uuid, *parent_alias)

		if err_update != nil {
			return err_update
		}
	}

	return nil
}
