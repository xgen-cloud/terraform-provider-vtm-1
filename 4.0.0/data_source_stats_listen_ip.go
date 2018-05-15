// Copyright (C) 2018, Pulse Secure, LLC. 
// Licensed under the terms of the MPL 2.0. See LICENSE file for details.

// Data Source Object ListenIp
package main

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	vtm "github.com/pulse-vadc/go-vtm/4.0"
)

func dataSourceListenIpStatistics() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceListenIpStatisticsRead,
		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			// Bytes sent to this listening IP.
			"bytes_in": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			// Bytes sent from this listening IP.
			"bytes_out": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			// TCP connections currently established to this listening IP.
			"current_conn": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			// Maximum number of simultaneous TCP connections this listening
			//  IP has processed at any one time.
			"max_conn": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			// Requests sent to this listening IP.
			"total_requests": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func dataSourceListenIpStatisticsRead(d *schema.ResourceData, tm interface{}) error {
	objectName := d.Get("name").(string)
	object, err := tm.(*vtm.VirtualTrafficManager).GetListenIpStatistics(objectName)
	if err != nil {
		if err.ErrorId == "resource.not_found" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Failed to read vtm_listen_ips '%v': %v", objectName, err.ErrorText)
	}
	d.Set("bytes_in", int(*object.Statistics.BytesIn))
	d.Set("bytes_out", int(*object.Statistics.BytesOut))
	d.Set("current_conn", int(*object.Statistics.CurrentConn))
	d.Set("max_conn", int(*object.Statistics.MaxConn))
	d.Set("total_requests", int(*object.Statistics.TotalRequests))
	d.SetId(objectName)
	return nil
}
