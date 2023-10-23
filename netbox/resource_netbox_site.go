package netbox

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/dcim"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxSiteStatusOptions = []string{"planned", "staging", "active", "decommissioning", "retired"}

func resourceNetboxSite() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxSiteCreate,
		ReadContext:   resourceNetboxSiteRead,
		UpdateContext: resourceNetboxSiteUpdate,
		Delete:        resourceNetboxSiteDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/sites-and-racks/#sites):
 
> How you choose to employ sites when modeling your network may vary depending on the nature of your organization, but generally a site will equate to a building or campus. For example, a chain of banks might create a site to represent each of its branches, a site for its corporate headquarters, and two additional sites for its presence in two colocation facilities.
>
> Each site must be assigned a unique name and may optionally be assigned to a region and/or tenant.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice(resourceNetboxSiteStatusOptions, false),
				Description:  buildValidValueDescription(resourceNetboxSiteStatusOptions),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"facility": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
			"longitude": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"latitude": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"physical_address": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"shipping_address": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"region_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"group_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			tagsKey: tagsSchema,
			"timezone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"asn_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNetboxSiteCreate(ctx context.Context, d *schema.ResourceData, m any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, fmt.Sprintf("Created Logsss"))
	api := m.(*client.NetBoxAPI)

	data := models.WritableSite{}

	name := d.Get("name").(string)
	data.Name = &name

	slugValue, slugOk := d.GetOk("slug")
	// Default slug to generated slug if not given
	if !slugOk {
		data.Slug = strToPtr(getSlug(name))
	} else {
		data.Slug = strToPtr(slugValue.(string))
	}

	data.Status = d.Get("status").(string)

	if description, ok := d.GetOk("description"); ok {
		data.Description = description.(string)
	}

	if facility, ok := d.GetOk("facility"); ok {
		data.Facility = facility.(string)
	}

	latitudeValue, ok := d.GetOk("latitude")
	if ok {
		data.Latitude = float64ToPtr(float64(latitudeValue.(float64)))
	}

	longitudeValue, ok := d.GetOk("longitude")
	if ok {
		data.Longitude = float64ToPtr(float64(longitudeValue.(float64)))
	}

	physicalAddressValue, ok := d.GetOk("physical_address")
	if ok {
		data.PhysicalAddress = physicalAddressValue.(string)
	}

	shippingAddressValue, ok := d.GetOk("shipping_address")
	if ok {
		data.ShippingAddress = shippingAddressValue.(string)
	}

	regionIDValue, ok := d.GetOk("region_id")
	if ok {
		data.Region = int64ToPtr(int64(regionIDValue.(int)))
	}

	groupIDValue, ok := d.GetOk("group_id")
	if ok {
		data.Group = int64ToPtr(int64(groupIDValue.(int)))
	}

	tenantIDValue, ok := d.GetOk("tenant_id")
	if ok {
		data.Tenant = int64ToPtr(int64(tenantIDValue.(int)))
	}

	if timezone, ok := d.GetOk("timezone"); ok {
		data.TimeZone = strToPtr(timezone.(string))
	}

	data.Asns = []int64{}
	if asnsValue, ok := d.GetOk("asn_ids"); ok {
		data.Asns = toInt64List(asnsValue)
	}

	data.Tags, _ = getNestedTagListFromResourceDataSet(api, d.Get(tagsKey))

	ct, ok := d.GetOk(customFieldsKey)
	if ok {
		data.CustomFields = ct
	}

	params := dcim.NewDcimSitesCreateParams().WithData(&data)
	fmt.Printf("\n Created params for site update with data: %+v", data)
	log.Printf("[DEBUG] Created params for site update with data: %+v", data)
	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Created params for site update with data: %+v", data))
	fmt.Printf("\n The Params %v", params)

	// res, err := api.Dcim.DcimSitesCreate(params, nil)
	// if err != nil {
	// 	return err
	// }

	res, err := api.Dcim.DcimSitesCreate(params, nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error creating the site",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(res.GetPayload().ID, 10))

	return resourceNetboxSiteRead(ctx, d, m)
}

func resourceNetboxSiteRead(ctx context.Context, d *schema.ResourceData, m any) (diags diag.Diagnostics) {
	api := m.(*client.NetBoxAPI)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := dcim.NewDcimSitesReadParams().WithID(id)

	res, err := api.Dcim.DcimSitesRead(params, nil)

	if err != nil {
		return diag.FromErr(
			fmt.Errorf(
				"Error"),
		)
		// if errresp, ok := err.(*dcim.DcimSitesReadDefault); ok {
		// 	errorcode := errresp.Code()
		// 	if errorcode == 404 {
		// 		// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
		// 		d.SetId("")
		// 		return nil
		// 	}
		// }
		// log.Printf("Error encountered: %v", err)
		// return err
	}

	site := res.GetPayload()

	d.Set("name", site.Name)
	d.Set("slug", site.Slug)
	d.Set("status", site.Status.Value)
	d.Set("description", site.Description)
	d.Set("facility", site.Facility)
	d.Set("longitude", site.Longitude)
	fmt.Printf("\n Read Longitude %v", site.Longitude)
	log.Printf("[DEBUG] Read Longitude %v", site.Longitude)
	tflog.Debug(ctx, fmt.Sprintf("Read Longitude %v", site.Longitude))
	d.Set("latitude", site.Latitude)
	fmt.Printf("\n Read Longitude %v", site.Latitude)
	log.Printf("[DEBUG] Read Longitude %v", site.Latitude)
	tflog.Debug(ctx, fmt.Sprintf("Read Longitude %v", site.Latitude))
	d.Set("physical_address", site.PhysicalAddress)
	d.Set("shipping_address", site.ShippingAddress)
	d.Set("timezone", site.TimeZone)
	d.Set("asn_ids", getIDsFromNestedASNList(site.Asns))

	if res.GetPayload().Region != nil {
		d.Set("region_id", res.GetPayload().Region.ID)
	} else {
		d.Set("region_id", nil)
	}

	if res.GetPayload().Group != nil {
		d.Set("group_id", res.GetPayload().Group.ID)
	} else {
		d.Set("group_id", nil)
	}

	if res.GetPayload().Tenant != nil {
		d.Set("tenant_id", res.GetPayload().Tenant.ID)
	} else {
		d.Set("tenant_id", nil)
	}

	cf := getCustomFields(res.GetPayload().CustomFields)
	if cf != nil {
		d.Set(customFieldsKey, cf)
	}
	d.Set(tagsKey, getTagListFromNestedTagList(res.GetPayload().Tags))

	return nil
}

func resourceNetboxSiteUpdate(ctx context.Context, d *schema.ResourceData, m any) (diags diag.Diagnostics) {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	data := models.WritableSite{}

	name := d.Get("name").(string)
	data.Name = &name

	slugValue, slugOk := d.GetOk("slug")
	// Default slug to generated slug if not given
	if !slugOk {
		data.Slug = strToPtr(getSlug(name))
	} else {
		data.Slug = strToPtr(slugValue.(string))
	}

	data.Status = d.Get("status").(string)

	if description, ok := d.GetOk("description"); ok {
		data.Description = description.(string)
	} else if d.HasChange("description") {
		// If GetOK returned unset description and its value changed, set it as a space string to delete it ...
		data.Description = " "
	}

	if facility, ok := d.GetOk("facility"); ok {
		data.Facility = facility.(string)
	}

	latitudeValue, ok := d.GetOk("latitude")
	if ok {
		data.Latitude = float64ToPtr(float64(latitudeValue.(float64)))
	}

	longitudeValue, ok := d.GetOk("longitude")
	if ok {
		data.Longitude = float64ToPtr(float64(longitudeValue.(float64)))
	}

	physicalAddressValue, ok := d.GetOk("physical_address")
	if ok {
		data.PhysicalAddress = physicalAddressValue.(string)
	} else if d.HasChange("physical_address") {
		// If GetOK returned unset description and its value changed, set it as a space string to delete it ...
		data.PhysicalAddress = " "
	}

	shippingAddressValue, ok := d.GetOk("shipping_address")
	if ok {
		data.ShippingAddress = shippingAddressValue.(string)
	} else if d.HasChange("shipping_address") {
		// If GetOK returned unset description and its value changed, set it as a space string to delete it ...
		data.ShippingAddress = " "
	}

	regionIDValue, ok := d.GetOk("region_id")
	if ok {
		data.Region = int64ToPtr(int64(regionIDValue.(int)))
	}

	groupIDValue, ok := d.GetOk("group_id")
	if ok {
		data.Group = int64ToPtr(int64(groupIDValue.(int)))
	}

	tenantIDValue, ok := d.GetOk("tenant_id")
	if ok {
		data.Tenant = int64ToPtr(int64(tenantIDValue.(int)))
	}

	if timezone, ok := d.GetOk("timezone"); ok {
		data.TimeZone = strToPtr(timezone.(string))
	}

	data.Asns = []int64{}
	if asnsValue, ok := d.GetOk("asn_ids"); ok {
		data.Asns = toInt64List(asnsValue)
	}

	data.Tags, _ = getNestedTagListFromResourceDataSet(api, d.Get(tagsKey))

	cf, ok := d.GetOk(customFieldsKey)
	if ok {
		data.CustomFields = cf
	}

	params := dcim.NewDcimSitesPartialUpdateParams().WithID(id).WithData(&data)

	fmt.Printf("\n Updated params for site update with ID %d and data: %+v", id, data)
	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Updated params for site update with data: %+v", data))
	tflog.Debug(ctx, fmt.Sprintf("\n The Params %v", params))
	// _, err := api.Dcim.DcimSitesPartialUpdate(params, nil)
	// if err != nil {
	// 	return err
	// }

	_, err := api.Dcim.DcimSitesPartialUpdate(params, nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error creating the site",
			Detail:   err.Error(),
		})
		return diags
	}

	return resourceNetboxSiteRead(ctx, d, m)
}

func resourceNetboxSiteDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := dcim.NewDcimSitesDeleteParams().WithID(id)

	_, err := api.Dcim.DcimSitesDelete(params, nil)
	if err != nil {
		if errresp, ok := err.(*dcim.DcimSitesDeleteDefault); ok {
			if errresp.Code() == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}
	return nil
}

func getIDsFromNestedASNList(nestedASNs []*models.NestedASN) []int64 {
	var asns []int64
	for _, asn := range nestedASNs {
		asns = append(asns, asn.ID)
	}
	return asns
}
