package opnsense

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/kradalby/opnsense-go/opnsense"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OPNSENSE_URL", nil),
				Description: "The OPNsense url to connect to",
			},
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OPNSENSE_KEY", nil),
				Description: "The OPNsense API key",
			},
			"secret": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("OPNSENSE_SECRET", nil),
				Description: "The OPNsense API secret",
			},
			"allow_unverified_tls": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("OPNSENSE_ALLOW_UNVERIFIED_TLS", false),
				Description: "Allow connection to a OPNsense server without verified TLS",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"opnsense_wireguard_client":    resourceWireGuardClient(),
			"opnsense_wireguard_server":    resourceWireGuardServer(),
			"opnsense_firewall_alias":      resourceFirewallAlias(),
			"opnsense_firewall_alias_util": resourceFirewallAliasUtil(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"opnsense_firewall_alias": dataFirewallAlias(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	url := d.Get("url").(string)
	key := d.Get("key").(string)
	secret := d.Get("secret").(string)
	skipTLS := d.Get("allow_unverified_tls").(bool)

	log.Printf("[TRACE] Creating OPNsense client\n")

	c, err := opnsense.NewClient(url, key, secret, skipTLS)
	if err != nil {
		log.Printf("[ERROR] Could not create OPNsense client: %#v\n", err)

		return nil, err
	}

	return c, nil
}
