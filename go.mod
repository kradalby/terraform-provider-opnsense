module github.com/kradalby/terraform-provider-opnsense

go 1.15

replace github.com/kradalby/opnsense-go => ../opnsense-go

require (
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.2
	github.com/kradalby/opnsense-go v0.0.0-20200906095057-b3959c4e0c07
	github.com/mitchellh/mapstructure v1.1.2
	github.com/satori/go.uuid v1.2.0
)
