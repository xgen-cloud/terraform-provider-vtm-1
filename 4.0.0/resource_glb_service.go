// Copyright (C) 2018, Pulse Secure, LLC. 
// Licensed under the terms of the MPL 2.0. See LICENSE file for details.

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	vtm "github.com/pulse-vadc/go-vtm/4.0"
)

func resourceGlbService() *schema.Resource {
	return &schema.Resource{
		Read:   resourceGlbServiceRead,
		Exists: resourceGlbServiceExists,
		Create: resourceGlbServiceCreate,
		Update: resourceGlbServiceUpdate,
		Delete: resourceGlbServiceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			// Defines the global load balancing algorithm to be used.
			"algorithm": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"chained", "geo", "hybrid", "load", "round_robin", "weighted_random"}, false),
				Default:      "hybrid",
			},

			// Are all the monitors required to be working in a location to
			//  mark this service as alive?
			"all_monitors_needed": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			// The last location to fail will be available as soon as it recovers.
			"autorecovery": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			// Enable/Disable automatic failback mode.
			"chained_auto_failback": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			// The locations this service operates for and defines the order
			//  in which locations fail.
			"chained_location_order": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			// Locations recovering from a failure will become disabled.
			"disable_on_failure": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			// A table mapping domains to the private keys that authenticate
			//  them
			"dnssec_keys": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{

						// domain
						"domain": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						// ssl_key
						"ssl_key": &schema.Schema{
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			// JSON representation of dnssec_keys
			"dnssec_keys_json": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateJsonString,
			},

			// The domains shown here should be a list of Fully Qualified Domain
			//  Names that you would like to balance globally. Responses from
			//  the back end DNS servers for queries that do not match this list
			//  will be forwarded to the client unmodified. Note: "*" may be
			//  used as a wild card.
			"domains": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			// Enable/Disable our response manipulation of DNS.
			"enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			// How much should the locality of visitors affect the choice of
			//  location used? This value is a percentage, 0% means that no locality
			//  information will be used, and 100% means that locality will always
			//  control which location is used. Values between the two extremes
			//  will act accordingly.
			"geo_effect": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
				Default:      50,
			},

			// The response to be sent in case there are no locations available.
			"last_resort_response": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			// This is the list of locations for which this service is draining.
			//  A location that is draining will never serve any of its service
			//  IP addresses for this domain. This can be used to take a location
			//  off-line.
			"location_draining": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			// Table containing location specific settings.
			"location_settings": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{

						// ips
						"ips": &schema.Schema{
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},

						// location
						"location": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						// monitors
						"monitors": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Default:  nil,
						},

						// weight
						"weight": &schema.Schema{
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 100),
							Default:      1,
						},
					},
				},
			},

			// JSON representation of location_settings
			"location_settings_json": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateJsonString,
			},

			// Return all or none of the IPs under complete failure.
			"return_ips_on_fail": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			// Response rules to be applied in the context of the service, in
			//  order, comma separated.
			"rules": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			// The TTL for the DNS resource records handled by the GLB service.
			"ttl": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
			},

			// Log connections to this GLB service?
			"log_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			// The filename the verbose query information should be logged to.
			//  Appliances will ignore this.
			"log_filename": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "%zeushome%/zxtm/log/services/%g.log",
			},

			// The format of the log lines.
			"log_format": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "%t, %s, %l, %q, %g, %n, %d, %a",
			},
		},
	}
}

func resourceGlbServiceRead(d *schema.ResourceData, tm interface{}) error {
	objectName := d.Get("name").(string)
	object, err := tm.(*vtm.VirtualTrafficManager).GetGlbService(objectName)
	if err != nil {
		if err.ErrorId == "resource.not_found" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Failed to read vtm_glb_service '%v': %v", objectName, err.ErrorText)
	}
	d.Set("algorithm", string(*object.Basic.Algorithm))
	d.Set("all_monitors_needed", bool(*object.Basic.AllMonitorsNeeded))
	d.Set("autorecovery", bool(*object.Basic.Autorecovery))
	d.Set("chained_auto_failback", bool(*object.Basic.ChainedAutoFailback))
	d.Set("chained_location_order", []string(*object.Basic.ChainedLocationOrder))
	d.Set("disable_on_failure", bool(*object.Basic.DisableOnFailure))

	dnssecKeys := make([]map[string]interface{}, 0, len(*object.Basic.DnssecKeys))
	for _, item := range *object.Basic.DnssecKeys {
		itemTerraform := make(map[string]interface{})
		if item.Domain != nil {
			itemTerraform["domain"] = string(*item.Domain)
		}
		if item.SslKey != nil {
			itemTerraform["ssl_key"] = []string(*item.SslKey)
		}
		dnssecKeys = append(dnssecKeys, itemTerraform)
	}
	d.Set("dnssec_keys", dnssecKeys)
	dnssecKeysJson, _ := json.Marshal(dnssecKeys)
	d.Set("dnssec_keys_json", dnssecKeysJson)
	d.Set("domains", []string(*object.Basic.Domains))
	d.Set("enabled", bool(*object.Basic.Enabled))
	d.Set("geo_effect", int(*object.Basic.GeoEffect))
	d.Set("last_resort_response", []string(*object.Basic.LastResortResponse))
	d.Set("location_draining", []string(*object.Basic.LocationDraining))

	locationSettings := make([]map[string]interface{}, 0, len(*object.Basic.LocationSettings))
	for _, item := range *object.Basic.LocationSettings {
		itemTerraform := make(map[string]interface{})
		if item.Ips != nil {
			itemTerraform["ips"] = []string(*item.Ips)
		}
		if item.Location != nil {
			itemTerraform["location"] = string(*item.Location)
		}
		if item.Monitors != nil {
			itemTerraform["monitors"] = []string(*item.Monitors)
		}
		if item.Weight != nil {
			itemTerraform["weight"] = int(*item.Weight)
		}
		locationSettings = append(locationSettings, itemTerraform)
	}
	d.Set("location_settings", locationSettings)
	locationSettingsJson, _ := json.Marshal(locationSettings)
	d.Set("location_settings_json", locationSettingsJson)
	d.Set("return_ips_on_fail", bool(*object.Basic.ReturnIpsOnFail))
	d.Set("rules", []string(*object.Basic.Rules))
	d.Set("ttl", int(*object.Basic.Ttl))
	d.Set("log_enabled", bool(*object.Log.Enabled))
	d.Set("log_filename", string(*object.Log.Filename))
	d.Set("log_format", string(*object.Log.Format))

	d.SetId(objectName)
	return nil
}

func resourceGlbServiceExists(d *schema.ResourceData, tm interface{}) (bool, error) {
	objectName := d.Get("name").(string)
	_, err := tm.(*vtm.VirtualTrafficManager).GetGlbService(objectName)
	if err != nil {
		if err.ErrorId == "resource.not_found" {
			return false, nil
		}
		return false, fmt.Errorf("%v", err.ErrorText)
	}
	return true, nil
}

func resourceGlbServiceCreate(d *schema.ResourceData, tm interface{}) error {
	objectName := d.Get("name").(string)
	object := tm.(*vtm.VirtualTrafficManager).NewGlbService(objectName)
	setString(&object.Basic.Algorithm, d, "algorithm")
	setBool(&object.Basic.AllMonitorsNeeded, d, "all_monitors_needed")
	setBool(&object.Basic.Autorecovery, d, "autorecovery")
	setBool(&object.Basic.ChainedAutoFailback, d, "chained_auto_failback")

	if _, ok := d.GetOk("chained_location_order"); ok {
		setStringList(&object.Basic.ChainedLocationOrder, d, "chained_location_order")
	} else {
		object.Basic.ChainedLocationOrder = &[]string{}
		d.Set("chained_location_order", []string(*object.Basic.ChainedLocationOrder))
	}
	setBool(&object.Basic.DisableOnFailure, d, "disable_on_failure")

	if _, ok := d.GetOk("domains"); ok {
		setStringList(&object.Basic.Domains, d, "domains")
	} else {
		object.Basic.Domains = &[]string{}
		d.Set("domains", []string(*object.Basic.Domains))
	}
	setBool(&object.Basic.Enabled, d, "enabled")
	setInt(&object.Basic.GeoEffect, d, "geo_effect")

	if _, ok := d.GetOk("last_resort_response"); ok {
		setStringList(&object.Basic.LastResortResponse, d, "last_resort_response")
	} else {
		object.Basic.LastResortResponse = &[]string{}
		d.Set("last_resort_response", []string(*object.Basic.LastResortResponse))
	}

	if _, ok := d.GetOk("location_draining"); ok {
		setStringList(&object.Basic.LocationDraining, d, "location_draining")
	} else {
		object.Basic.LocationDraining = &[]string{}
		d.Set("location_draining", []string(*object.Basic.LocationDraining))
	}
	setBool(&object.Basic.ReturnIpsOnFail, d, "return_ips_on_fail")

	if _, ok := d.GetOk("rules"); ok {
		setStringList(&object.Basic.Rules, d, "rules")
	} else {
		object.Basic.Rules = &[]string{}
		d.Set("rules", []string(*object.Basic.Rules))
	}
	setInt(&object.Basic.Ttl, d, "ttl")

	object.Basic.DnssecKeys = &vtm.GlbServiceDnssecKeysTable{}
	if dnssecKeysJson, ok := d.GetOk("dnssec_keys_json"); ok {
		_ = json.Unmarshal([]byte(dnssecKeysJson.(string)), object.Basic.DnssecKeys)
	} else if dnssecKeys, ok := d.GetOk("dnssec_keys"); ok {
		for _, row := range dnssecKeys.(*schema.Set).List() { // VTM-37687: dnssecKeys.([]interface{}) {
			itemTerraform := row.(map[string]interface{})
			VtmObject := vtm.GlbServiceDnssecKeys{}
			VtmObject.Domain = getStringAddr(itemTerraform["domain"].(string))
			VtmObject.SslKey = getStringListAddr(expandStringList(itemTerraform["ssl_key"].([]interface{})))
			*object.Basic.DnssecKeys = append(*object.Basic.DnssecKeys, VtmObject)
		}
		d.Set("dnssec_keys", dnssecKeys)
	} else {
		d.Set("dnssec_keys", make([]map[string]interface{}, 0, len(*object.Basic.DnssecKeys)))
	}

	object.Basic.LocationSettings = &vtm.GlbServiceLocationSettingsTable{}
	if locationSettingsJson, ok := d.GetOk("location_settings_json"); ok {
		_ = json.Unmarshal([]byte(locationSettingsJson.(string)), object.Basic.LocationSettings)
	} else if locationSettings, ok := d.GetOk("location_settings"); ok {
		for _, row := range locationSettings.(*schema.Set).List() { // VTM-37687: locationSettings.([]interface{}) {
			itemTerraform := row.(map[string]interface{})
			VtmObject := vtm.GlbServiceLocationSettings{}
			VtmObject.Ips = getStringListAddr(expandStringList(itemTerraform["ips"].([]interface{})))
			VtmObject.Location = getStringAddr(itemTerraform["location"].(string))
			VtmObject.Monitors = getStringListAddr(expandStringList(itemTerraform["monitors"].([]interface{})))
			VtmObject.Weight = getIntAddr(itemTerraform["weight"].(int))
			*object.Basic.LocationSettings = append(*object.Basic.LocationSettings, VtmObject)
		}
		d.Set("location_settings", locationSettings)
	} else {
		d.Set("location_settings", make([]map[string]interface{}, 0, len(*object.Basic.LocationSettings)))
	}
	setBool(&object.Log.Enabled, d, "log_enabled")
	setString(&object.Log.Filename, d, "log_filename")
	setString(&object.Log.Format, d, "log_format")

	object.Apply()
	d.SetId(objectName)
	return nil
}

func resourceGlbServiceUpdate(d *schema.ResourceData, tm interface{}) error {
	objectName := d.Get("name").(string)
	object, err := tm.(*vtm.VirtualTrafficManager).GetGlbService(objectName)
	if err != nil {
		return fmt.Errorf("Failed to update vtm_glb_service '%v': %v", objectName, err)
	}
	setString(&object.Basic.Algorithm, d, "algorithm")
	setBool(&object.Basic.AllMonitorsNeeded, d, "all_monitors_needed")
	setBool(&object.Basic.Autorecovery, d, "autorecovery")
	setBool(&object.Basic.ChainedAutoFailback, d, "chained_auto_failback")

	if _, ok := d.GetOk("chained_location_order"); ok {
		setStringList(&object.Basic.ChainedLocationOrder, d, "chained_location_order")
	} else {
		object.Basic.ChainedLocationOrder = &[]string{}
		d.Set("chained_location_order", []string(*object.Basic.ChainedLocationOrder))
	}
	setBool(&object.Basic.DisableOnFailure, d, "disable_on_failure")

	if _, ok := d.GetOk("domains"); ok {
		setStringList(&object.Basic.Domains, d, "domains")
	} else {
		object.Basic.Domains = &[]string{}
		d.Set("domains", []string(*object.Basic.Domains))
	}
	setBool(&object.Basic.Enabled, d, "enabled")
	setInt(&object.Basic.GeoEffect, d, "geo_effect")

	if _, ok := d.GetOk("last_resort_response"); ok {
		setStringList(&object.Basic.LastResortResponse, d, "last_resort_response")
	} else {
		object.Basic.LastResortResponse = &[]string{}
		d.Set("last_resort_response", []string(*object.Basic.LastResortResponse))
	}

	if _, ok := d.GetOk("location_draining"); ok {
		setStringList(&object.Basic.LocationDraining, d, "location_draining")
	} else {
		object.Basic.LocationDraining = &[]string{}
		d.Set("location_draining", []string(*object.Basic.LocationDraining))
	}
	setBool(&object.Basic.ReturnIpsOnFail, d, "return_ips_on_fail")

	if _, ok := d.GetOk("rules"); ok {
		setStringList(&object.Basic.Rules, d, "rules")
	} else {
		object.Basic.Rules = &[]string{}
		d.Set("rules", []string(*object.Basic.Rules))
	}
	setInt(&object.Basic.Ttl, d, "ttl")

	object.Basic.DnssecKeys = &vtm.GlbServiceDnssecKeysTable{}
	if dnssecKeysJson, ok := d.GetOk("dnssec_keys_json"); ok {
		_ = json.Unmarshal([]byte(dnssecKeysJson.(string)), object.Basic.DnssecKeys)
	} else if dnssecKeys, ok := d.GetOk("dnssec_keys"); ok {
		for _, row := range dnssecKeys.(*schema.Set).List() { // VTM-37687: dnssecKeys.([]interface{}) {
			itemTerraform := row.(map[string]interface{})
			VtmObject := vtm.GlbServiceDnssecKeys{}
			VtmObject.Domain = getStringAddr(itemTerraform["domain"].(string))
			VtmObject.SslKey = getStringListAddr(expandStringList(itemTerraform["ssl_key"].([]interface{})))
			*object.Basic.DnssecKeys = append(*object.Basic.DnssecKeys, VtmObject)
		}
		d.Set("dnssec_keys", dnssecKeys)
	} else {
		d.Set("dnssec_keys", make([]map[string]interface{}, 0, len(*object.Basic.DnssecKeys)))
	}

	object.Basic.LocationSettings = &vtm.GlbServiceLocationSettingsTable{}
	if locationSettingsJson, ok := d.GetOk("location_settings_json"); ok {
		_ = json.Unmarshal([]byte(locationSettingsJson.(string)), object.Basic.LocationSettings)
	} else if locationSettings, ok := d.GetOk("location_settings"); ok {
		for _, row := range locationSettings.(*schema.Set).List() { // VTM-37687: locationSettings.([]interface{}) {
			itemTerraform := row.(map[string]interface{})
			VtmObject := vtm.GlbServiceLocationSettings{}
			VtmObject.Ips = getStringListAddr(expandStringList(itemTerraform["ips"].([]interface{})))
			VtmObject.Location = getStringAddr(itemTerraform["location"].(string))
			VtmObject.Monitors = getStringListAddr(expandStringList(itemTerraform["monitors"].([]interface{})))
			VtmObject.Weight = getIntAddr(itemTerraform["weight"].(int))
			*object.Basic.LocationSettings = append(*object.Basic.LocationSettings, VtmObject)
		}
		d.Set("location_settings", locationSettings)
	} else {
		d.Set("location_settings", make([]map[string]interface{}, 0, len(*object.Basic.LocationSettings)))
	}
	setBool(&object.Log.Enabled, d, "log_enabled")
	setString(&object.Log.Filename, d, "log_filename")
	setString(&object.Log.Format, d, "log_format")

	object.Apply()
	d.SetId(objectName)
	return nil
}

func resourceGlbServiceDelete(d *schema.ResourceData, tm interface{}) error {
	objectName := d.Get("name").(string)
	err := tm.(*vtm.VirtualTrafficManager).DeleteGlbService(objectName)
	if err != nil {
		return fmt.Errorf("Failed to delete vtm_glb_service '%v': %v", objectName, err.ErrorText)
	}
	d.SetId("")
	return nil
}
