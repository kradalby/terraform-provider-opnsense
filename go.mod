module github.com/kradalby/terraform-provider-opnsense

go 1.15

replace github.com/kradalby/opnsense-go => ../opnsense-go

require (
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.2
	github.com/kradalby/opnsense-go v0.0.0-20200906081803-a4d782408c43
	github.com/satori/go.uuid v1.2.0
)
