package opnsense

import (
	//"fmt"
	"github.com/cdeconinck/opnsense-go/opnsense"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func resourceFirewallAliasPush() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallAliasPushCreate,
		Read:   resourceFirewallAliasPushRead,
		Update: resourceFirewallAliasPushUpdate,
		Delete: resourceFirewallAliasPushDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the alias",
				Required:    true,
			},
			"address": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

// By disabling this function, we prevent terraform to update his state whith all the address set in alias
// this allow us to not track addresses already set
func resourceFirewallAliasPushRead(d *schema.ResourceData, meta interface{}) error {
	/*log.Printf("[TRACE] Getting OPNsense client from meta")
	c := meta.(*opnsense.Client)

	log.Printf("[TRACE] Fetching client configuration from OPNsense")
	client, err := c.AliasPushGet(d.Id())
	if err != nil {
		// temporary fix for the internal error API when we try to get an unreferenced UIID
		if err.Error() == "Internal Error status code received" {
			d.SetId("")
			return nil
		}
		log.Printf("ERROR: \n%#v", err)
		log.Printf("[ERROR] Failed to fetch name : %s", d.Id())
		return err
	}

	log.Printf("[DEBUG] Configuration from OPNsense: \n")
	log.Printf("[DEBUG] %#v \n", client)

	d.Set("name", client.Name)
	d.SetId(client.Name)

	address_list := make([]string, 0)
	if client.Rows != nil {
		for _, v := range client.Rows {
			address_list = append(address_list, v.Address)
		}
	}

	d.Set("address", address_list)
	*/
	return nil
}

func resourceFirewallAliasPushCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	name := d.Get("name").(string)
	address_list := d.Get("address").(*schema.Set).List()

	err := del_list(c, d.Id(), address_list)
	if err != nil {
		return err
	}

	d.SetId(name)
	resourceFirewallAliasPushRead(d, meta)

	return nil
}

func resourceFirewallAliasPushUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)
	var name = d.Get("name").(string)

	if d.HasChange("address") {
		var old_address, new_address = d.GetChange("address")

		old_set := old_address.(*schema.Set)
		new_set := new_address.(*schema.Set)

		list_to_del := old_set.Difference(new_set).List()
		list_to_add := new_set.Difference(old_set).List()

		err_add := add_list(c, name, list_to_add)
		if err_add != nil {
			return err_add
		}

		err_del := del_list(c, name, list_to_del)
		if err_del != nil {
			return err_del
		}
	}

	d.SetId(name)

	return resourceFirewallAliasPushRead(d, meta)
}

func resourceFirewallAliasPushDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)

	address_list := d.Get("address").(*schema.Set).List()
	// remove one by one the address
	err := del_list(c, d.Id(), address_list)
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func add_list(client *opnsense.Client, name string, list []interface{}) error {
	conf := opnsense.AliasPushSet{}

	for _, v := range list {
		conf.Address = v.(string)
		log.Println("[TRACE] adding :", conf.Address)

		_, err := client.AliasPushAdd(name, conf)

		if err != nil {
			return err
		}
	}

	return nil
}

func del_list(client *opnsense.Client, name string, list []interface{}) error {
	conf := opnsense.AliasPushSet{}

	for _, v := range list {
		conf.Address = v.(string)
		log.Println("[TRACE] removing :", conf.Address)

		_, err := client.AliasPushDel(name, conf)

		if err != nil {
			return err
		}
	}

	return nil
}
