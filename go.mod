module github.com/kradalby/terraform-provider-opnsense

go 1.15

// replace github.com/kradalby/opnsense-go => ../opnsense-go

require (
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.2.0
	github.com/kradalby/opnsense-go v0.0.0-20200916130135-67bf4464de4c
	github.com/mitchellh/mapstructure v1.3.3
	github.com/satori/go.uuid v1.2.0
)
