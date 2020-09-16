package opnsense

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kradalby/opnsense-go/opnsense"
)

func resourceFirmware() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFirmwareCreate,
		ReadContext:   resourceFirmwareRead,
		UpdateContext: resourceFirmwareUpdate,
		DeleteContext: resourceFirmwareDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"plugin": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "A plugin installed to OPNsense",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"installed": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"comment": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"flatsize": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"locked": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"license": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"repository": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"origin": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"provided": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"configured": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceFirmwareRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] Getting OPNsense client from meta")

	var diags diag.Diagnostics

	c := meta.(*opnsense.Client)

	installedPlugins, err := c.FirmwareInstalledPluginsList()
	if err != nil {
		log.Printf("[DEBUG]: \n%#v", err)
		log.Println("[ERROR] Failed to fetch information")

		return diag.FromErr(err)
	}

	installedPluginMaps := make([]map[string]interface{}, len(installedPlugins))

	for index, plugin := range installedPlugins {
		installedPluginMaps[index] = opnsense.StructToMap(plugin)
	}

	if err := d.Set("plugin", installedPluginMaps); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("firmware")

	return diags
}

func resourceFirmwareCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*opnsense.Client)

	added := d.Get("plugin").(*schema.Set)

	diags := installPlugins(ctx, d, c, added)
	if diags.HasError() {
		return diags
	}

	return resourceFirmwareRead(ctx, d, meta)
}

func resourceFirmwareUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*opnsense.Client)

	if d.HasChange("plugin") {
		oldRaw, newRaw := d.GetChange("plugin")
		old := oldRaw.(*schema.Set)
		new := newRaw.(*schema.Set)

		added := new.Difference(old)
		removed := old.Difference(new)

		diags := installPlugins(ctx, d, c, added)
		if diags.HasError() {
			return diags
		}

		diags = removePlugins(ctx, d, c, removed)
		if diags.HasError() {
			return diags
		}
	}

	return resourceFirmwareRead(ctx, d, meta)
}

func resourceFirmwareDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*opnsense.Client)

	removed := d.Get("plugin").(*schema.Set)

	diags := removePlugins(ctx, d, c, removed)
	if diags.HasError() {
		return diags
	}

	resourceFirmwareRead(ctx, d, meta)
	d.SetId("")

	return diags
}

func statusStateConf(d *schema.ResourceData, client *opnsense.Client) *resource.StateChangeConf {
	createStateConf := &resource.StateChangeConf{
		Pending: []string{
			opnsense.StatusRunning,
		},
		Target: []string{
			opnsense.StatusDone,
		},
		Refresh: func() (interface{}, string, error) {
			resp, err := client.FirmwareUpgradeStatus()
			if err != nil {
				return 0, "", err
			}

			return resp, resp.Status, nil
		},
		Timeout: d.Timeout(schema.TimeoutCreate),

		Delay:                     10 * time.Second,
		MinTimeout:                5 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	return createStateConf
}

func installPlugins(ctx context.Context,
	d *schema.ResourceData,
	c *opnsense.Client,
	added *schema.Set) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, plug := range added.List() {
		plugin := plug.(map[string]interface{})

		name := plugin["name"].(string)

		err := c.FirmwareInstall(name)
		if err != nil {
			return diag.FromErr(err)
		}

		upgradeChecker := statusStateConf(d, c)

		_, err = upgradeChecker.WaitForStateContext(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func removePlugins(ctx context.Context,
	d *schema.ResourceData,
	c *opnsense.Client,
	removed *schema.Set) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, plug := range removed.List() {
		plugin := plug.(map[string]interface{})

		name := plugin["name"].(string)

		err := c.FirmwareRemove(name)
		if err != nil {
			return diag.FromErr(err)
		}

		upgradeChecker := statusStateConf(d, c)

		_, err = upgradeChecker.WaitForStateContext(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
