package main

import (
	"github.com/cdeconinck/terraform-provider-opnsense/opnsense"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: opnsense.Provider})
}
