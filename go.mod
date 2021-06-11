module github.com/kradalby/terraform-provider-opnsense

go 1.15

// replace github.com/kradalby/opnsense-go => ../opnsense-go

require (
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.6.1
	github.com/kradalby/opnsense-go v0.0.0-20210528195939-75c226e47325
	github.com/mitchellh/mapstructure v1.4.1
	github.com/rogpeppe/go-internal v1.6.2 // indirect
	github.com/satori/go.uuid v1.2.0
)
