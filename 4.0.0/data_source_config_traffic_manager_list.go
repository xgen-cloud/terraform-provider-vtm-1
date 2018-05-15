// Copyright (C) 2018, Pulse Secure, LLC. 
// Licensed under the terms of the MPL 2.0. See LICENSE file for details.

package main

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	vtm "github.com/pulse-vadc/go-vtm/4.0"
)

func dataSourceTrafficManagerList() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTrafficManagerListRead,

		Schema: map[string]*schema.Schema{
			"object_list": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"starts_with": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ends_with": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"contains": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"regex_match": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
		},
	}
}

func dataSourceTrafficManagerListRead(d *schema.ResourceData, tm interface{}) error {
	objectList, err := tm.(*vtm.VirtualTrafficManager).ListTrafficManagers()
	if err != nil {
		d.SetId("")
		return fmt.Errorf("Failed to read vtm_traffic_manager_list: %v", err.ErrorText)
	}

	if starts_with, ok := d.GetOk("starts_with"); ok {
		objectList = getStringListStartingWith(objectList, starts_with.(string))
	}
	if ends_with, ok := d.GetOk("ends_with"); ok {
		objectList = getStringListEndingWith(objectList, ends_with.(string))
	}
	if contains, ok := d.GetOk("contains"); ok {
		objectList = getStringListContaining(objectList, contains.(string))
	}
	var regexErr error
	if regex_match, ok := d.GetOk("regex_match"); ok {
		objectList, regexErr = getStringListMatchingRegex(objectList, regex_match.(string))
		if regexErr != nil {
			d.SetId("")
			return regexErr
		}
	}

	d.Set("object_list", objectList)
	d.SetId("traffic_manager_list")
	return nil
}
