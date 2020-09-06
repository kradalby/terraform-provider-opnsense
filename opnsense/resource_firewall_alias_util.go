package opnsense

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kradalby/opnsense-go/opnsense"
)

func resourceFirewallAliasUtil() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallAliasUtilCreate,
		Read:   resourceFirewallAliasUtilRead,
		Update: resourceFirewallAliasUtilUpdate,
		Delete: resourceFirewallAliasUtilDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the alias",
				Required:    true,
			},
			"address": {
				Type:        schema.TypeString,
				Description: "IP or CIDR address to add in the alias",
				Required:    true,
			},
		},
	}
}

func resourceFirewallAliasUtilRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	name := d.Get("name").(string)

	alias, err := c.AliasUtilsGet(name)
	if err != nil {
		d.SetId("")
		// fix for the internal error API received when we try to get an unreferenced alias
		if err.Error() == apiInternalErrorMsg {
			return nil
		}

		return fmt.Errorf("list of address used in the alias '%s' could not be retreived : %w", name, err)
	}

	// We don't want to retrieved all addresse present in the alias because this resource will manage only one address
	// instead we just check if the address is present or not in the alias
	addressInState := d.Get("address").(string)
	addressFound := false

	if alias.Rows != nil {
		for _, v := range alias.Rows {
			if v.Address == addressInState {
				addressFound = true

				break
			}
		}
	}

	if addressFound {
		err = d.Set("address", addressInState)
		if err != nil {
			return err
		}
	} else {
		// address no found in the alias, we tell Terraform that the resource no longer exists
		d.SetId("")
	}

	return nil
}

func resourceFirewallAliasUtilCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	name := d.Get("name").(string)
	address := d.Get("address").(string)
	conf := opnsense.AliasUtilsSet{
		Address: address,
	}

	_, err := c.AliasUtilsAdd(name, conf)
	if err != nil {
		return fmt.Errorf("failed to add '%s' from alias '%s' : %w", conf.Address, name, err)
	}

	d.SetId(address)

	return resourceFirewallAliasUtilRead(d, meta)
}

func resourceFirewallAliasUtilUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)
	conf := opnsense.AliasUtilsSet{}

	oldAddress := d.Get("address")
	newAddress := d.Get("address")
	oldName := d.Get("name")
	newName := d.Get("name")

	if d.HasChange("name") {
		oldName, newName = d.GetChange("name")
	}

	if d.HasChange("address") {
		oldAddress, newAddress = d.GetChange("address")
	}

	// We always try to add the new address before removing the previous one because this
	// could result in an unsafe state where  we have deleted the address without added the new one
	conf.Address = newAddress.(string)

	_, err := c.AliasUtilsAdd(newName.(string), conf)
	if err != nil {
		return fmt.Errorf("failed to add '%s' in alias '%s' : %w", conf.Address, newName.(string), err)
	}

	conf.Address = oldAddress.(string)
	_, err = c.AliasUtilsDel(oldName.(string), conf)

	if err != nil {
		return fmt.Errorf("failed to remove '%s' in alias '%s' : %w", conf.Address, oldName.(string), err)
	}

	d.SetId(newAddress.(string))

	return resourceFirewallAliasUtilRead(d, meta)
}

func resourceFirewallAliasUtilDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	name := d.Get("name").(string)
	conf := opnsense.AliasUtilsSet{
		Address: d.Get("address").(string),
	}

	_, err := c.AliasUtilsDel(name, conf)
	if err != nil {
		// normally we should handle here the case when the API return an error and check if the
		// resource might already be destroyed (manually for example) but the API doesn't return
		// an error when we try to delete a non existing address in the alias so we can skip
		// this logic and always return the error send by opnsense
		return fmt.Errorf("failed to remove '%s' from alias '%s' : %w", conf.Address, name, err)
	}

	return nil
}
