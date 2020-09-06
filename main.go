package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/kradalby/terraform-provider-opnsense/opnsense"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: providerFunc,
	})
}

func providerFunc() *schema.Provider {
	return opnsense.Provider()
}
