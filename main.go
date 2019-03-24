package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/kradalby/terraform-provider-opnsense/opnsense"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: opnsense.Provider})
}
