module github.com/kradalby/terraform-provider-opnsense

go 1.15

// replace github.com/kradalby/opnsense-go => ../opnsense-go

require (
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.4
	github.com/kradalby/opnsense-go v0.0.0-20200916123608-6df8ebb1a878
	github.com/mitchellh/mapstructure v1.3.3
	github.com/satori/go.uuid v1.2.0
)
