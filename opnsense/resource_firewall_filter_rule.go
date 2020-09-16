package opnsense

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/kradalby/opnsense-go/opnsense"
	"github.com/mitchellh/mapstructure"
	uuid "github.com/satori/go.uuid"
)

func resourceFirewallFilterRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFirewallFilterRuleCreate,
		ReadContext:   resourceFirewallFilterRuleRead,
		UpdateContext: resourceFirewallFilterRuleUpdate,
		DeleteContext: resourceFirewallFilterRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"uuid": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Unique ID",
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"sequence": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "pass",
				ValidateFunc: validation.StringInSlice([]string{"pass", "block", "reject"}, false),
			},
			"quick": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"interface": {
				Type:     schema.TypeString,
				Required: true,
			},
			"direction": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "in",
				ValidateFunc: validation.StringInSlice([]string{"in", "out"}, false),
			},
			"ipprotocol": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "ipv4",
				ValidateFunc: validation.StringInSlice([]string{"ipv4", "ipv6"}, false),
			},
			"protocol": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "any",
			},
			"source_net": {
				Type:     schema.TypeString,
				Required: true,
			},
			"source_not": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"source_port": {
				Type:     schema.TypeString,
				Required: true,
			},
			"destination_net": {
				Type:     schema.TypeString,
				Required: true,
			},
			"destination_not": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"destination_port": {
				Type:     schema.TypeString,
				Required: true,
			},
			"gateway": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"log": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceFirewallFilterRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Getting OPNsense client from meta")

	var diags diag.Diagnostics

	c := meta.(*opnsense.Client)

	log.Printf("[TRACE] Converting ID to UUID")

	uuid, err := uuid.FromString(d.Id())
	if err != nil {
		log.Printf("[ERROR] Failed to parse ID")

		return diag.FromErr(err)
	}

	rule, err := c.FirewallFilterRuleGet(uuid)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get rule from OPNsense",
			Detail: fmt.Sprintf(
				"When attempting to fetch the rule %s, the API returned %s",
				uuid, err,
			),
		})

		return diags

	}

	ruleMap := opnsense.StructToMap(rule)

	for k, v := range ruleMap {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(uuid.String())

	return diags
}

func resourceFirewallFilterRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*opnsense.Client)

	rule := opnsense.FilterRule{}
	ruleMap := make(map[string]interface{})

	for _, field := range opnsense.JSONFields(rule) {
		ruleMap[field] = d.Get(field)
	}

	err := mapstructure.Decode(ruleMap, &rule)
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.FirewallFilterRuleAdd(&rule)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFirewallFilterRuleRead(ctx, d, meta)
}

func resourceFirewallFilterRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*opnsense.Client)

	rule := opnsense.FilterRule{}
	ruleMap := make(map[string]interface{})

	for _, field := range opnsense.JSONFields(rule) {
		ruleMap[field] = d.Get(field)
	}

	err := mapstructure.Decode(ruleMap, &rule)
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.FirewallFilterRuleSet(&rule)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFirewallFilterRuleRead(ctx, d, meta)
}

func resourceFirewallFilterRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*opnsense.Client)

	var diags diag.Diagnostics

	uuid, err := uuid.FromString(d.Id())
	if err != nil {
		log.Printf("[ERROR] Failed to parse ID")

		return diag.FromErr(err)
	}

	err = c.FirewallFilterRuleDelete(uuid)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

// func statusStateConf(d *schema.ResourceData, client *opnsense.Client) *resource.StateChangeConf {
// 	createStateConf := &resource.StateChangeConf{
// 		Pending: []string{
// 			opnsense.StatusRunning,
// 		},
// 		Target: []string{
// 			opnsense.StatusDone,
// 		},
// 		Refresh: func() (interface{}, string, error) {
// 			resp, err := client.FirmwareUpgradeStatus()
// 			if err != nil {
// 				return 0, "", err
// 			}

// 			return resp, resp.Status, nil
// 		},
// 		Timeout: d.Timeout(schema.TimeoutCreate),

// 		Delay:                     10 * time.Second,
// 		MinTimeout:                5 * time.Second,
// 		ContinuousTargetOccurence: 2,
// 	}

// 	return createStateConf
// }
