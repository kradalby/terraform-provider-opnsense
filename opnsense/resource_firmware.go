package opnsense

import (
	"context"
	"fmt"
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
				Description: "A specification for a virtual disk device on this virtual machine.",
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

	log.Printf("[TRACE] Converting ID to UUID")

	info, err := c.GetInformation()
	if err != nil {
		if err.Error() == "found empty array, most likely 404" {
			d.SetId("")

			return diags
		}

		log.Printf("[DEBUG]: \n%#v", err)
		log.Println("[ERROR] Failed to fetch information")

		return diag.FromErr(err)
	}

	pluginsInOpnsense := make([]map[string]interface{}, 0)
	// pluginsInTerraform := make(map[string]bool)

	for _, plugin := range info.Plugin {
		if bool(plugin.Installed) {
			pluginsInOpnsense = append(pluginsInOpnsense,
				map[string]interface{}{
					"name":       plugin.Name,
					"version":    plugin.Version,
					"comment":    plugin.Comment,
					"flatsize":   plugin.Flatsize,
					"locked":     plugin.Locked,
					"license":    plugin.License,
					"repository": plugin.Repository,
					"origin":     plugin.Origin,
					"provided":   plugin.Provided,
					"installed":  plugin.Installed,
					"path":       plugin.Path,
					"configured": plugin.Configured,
				},
			)
		}
	}

	fmt.Println("[DEBUG]", pluginsInOpnsense)

	if err := d.Set("plugin", pluginsInOpnsense); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceFirmwareCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceFirmwareRead(ctx, d, meta)
}

func resourceFirmwareUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*opnsense.Client)

	if d.HasChange("plugin") {
		oldRaw, newRaw := d.GetChange("plugin")
		old := oldRaw.(*schema.Set)
		new := newRaw.(*schema.Set)

		removed := old.Difference(new)
		added := new.Difference(old)

		fmt.Println("[DEBUG] Added:", added)
		fmt.Println("[DEBUG] Removed:", removed)

		for _, plug := range added.List() {
			plugin := plug.(map[string]interface{})

			name := plugin["name"].(string)

			resp, err := c.PackageInstall(name)
			if err != nil {
				return diag.FromErr(err)
			}

			if resp.Status != "ok" {
				return diag.FromErr(fmt.Errorf("status: %s, err: %w", resp.Status, ErrStatusNotOk))
			}

			upgradeChecker := statusStateConf(d, c)

			_, err = upgradeChecker.WaitForStateContext(ctx)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		for _, plug := range removed.List() {
			plugin := plug.(map[string]interface{})

			name := plugin["name"].(string)

			resp, err := c.PackageRemove(name)
			if err != nil {
				return diag.FromErr(err)
			}

			if resp.Status != "ok" {
				return diag.FromErr(fmt.Errorf("status: %s, err: %w", resp.Status, ErrStatusNotOk))
			}

			upgradeChecker := statusStateConf(d, c)

			_, err = upgradeChecker.WaitForStateContext(ctx)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceFirmwareRead(ctx, d, meta)
}

func resourceFirmwareDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

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
			resp, err := client.GetUpgradeStatus()
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
