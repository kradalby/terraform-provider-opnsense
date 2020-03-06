package opnsense

import (
	//"github.com/cdeconinck/opnsense-go/opnsense"
	"github.com/hashicorp/terraform/helper/schema"
	//"github.com/satori/go.uuid"
	//"log"
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
		},
	}
}

// Read will fetch the data of a resource.
func dataFirewallAliasRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}
