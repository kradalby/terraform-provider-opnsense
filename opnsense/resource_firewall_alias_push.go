package opnsense

import (
	"fmt"
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
	log.Printf("[TRACE] Getting OPNsense client from meta")
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

	return nil
}

func resourceFirewallAliasPushCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)
	client := opnsense.AliasPushSet{}
	var name = d.Get("name").(string)
	var address_list = d.Get("address")
	if address_list == nil {
		return fmt.Errorf("Undefined conf")
	}

	list_set := address_list.(*schema.Set).List()

	// add one by one the address
	for _, v := range list_set {
		client.Address = v.(string)

		_, err := c.AliasPushAdd(name, client)

		if err != nil {
			return err
		}
	}

	d.SetId(name)
	resourceFirewallAliasPushRead(d, meta)

	return nil
}

func resourceFirewallAliasPushUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)
	client := opnsense.AliasPushSet{}
	var name = d.Get("name").(string)

	if d.HasChange("address") {
		// il éxiste surement un meilleur moyen pour trouver les différences...
		var old_address, new_address = d.GetChange("address")
		old_set := old_address.(*schema.Set).List()
		new_set := new_address.(*schema.Set).List()

		list_old := []string{}
		for _, v := range old_set {
			list_old = append(list_old, v.(string))
		}

		list_new := []string{}
		for _, v := range new_set {
			list_new = append(list_new, v.(string))
		}

		log.Println("[TRACE] old_address :", list_old)
		log.Println("[TRACE] new_address :", list_new)

		for _, v := range list_new {
			if !contains(list_old, v) {
				log.Println("[TRACE] to add :", v)
				client.Address = v

				_, err := c.AliasPushAdd(name, client)

				if err != nil {
					return err
				}
			}
		}

		for _, v := range list_old {
			if !contains(list_new, v) {
				log.Println("[TRACE] to del :", v)

				client.Address = v

				_, err := c.AliasPushDel(name, client)

				if err != nil {
					return err
				}
			}
		}
	}

	d.SetId(name)
	resourceFirewallAliasPushRead(d, meta)

	return nil
}

func resourceFirewallAliasPushDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*opnsense.Client)
	client := opnsense.AliasPushSet{}

	var address_list = d.Get("address")
	if address_list == nil {
		return fmt.Errorf("Undefined conf")
	}

	list_set := address_list.(*schema.Set).List()

	// add one by one the address
	for _, v := range list_set {
		client.Address = v.(string)

		_, err := c.AliasPushDel(d.Id(), client)

		if err != nil {
			return err
		}
	}

	d.SetId("")

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
