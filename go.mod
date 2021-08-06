module github.com/kradalby/terraform-provider-opnsense

go 1.15

// replace github.com/kradalby/opnsense-go => ../opnsense-go

require (
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.7.0
	github.com/kradalby/opnsense-go v0.0.0-20210802162605-47fb9b5e0bbb
	github.com/mitchellh/mapstructure v1.4.1
	github.com/rogpeppe/go-internal v1.6.2 // indirect
	github.com/satori/go.uuid v1.2.0
)
