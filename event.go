/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"errors"
	"log"
)

var (
	ErrDatacenterIDInvalid          = errors.New("Datacenter VPC ID invalid")
	ErrDatacenterRegionInvalid      = errors.New("Datacenter Region invalid")
	ErrDatacenterCredentialsInvalid = errors.New("Datacenter credentials invalid")
	ErrNetworkSubnetInvalid         = errors.New("Network subnet invalid")
)

// Event stores the network create data
type Event struct {
	ID                    string `json:"id"`
	DatacenterVPCID       string `json:"datacenter_vpc_id"`
	DatacenterRegion      string `json:"datacenter_region"`
	DatacenterAccessKey   string `json:"datacenter_access_key"`
	DatacenterAccessToken string `json:"datacenter_access_token"`
	NetworkAWSID          string `json:"network_aws_id,omitempty"`
	NetworkSubnet         string `json:"network_subnet"`
	ErrorMessage          string `json:"error,omitempty"`
}

// Validate checks if all criteria are met
func (ev *Event) Validate() error {
	if ev.DatacenterVPCID == "" {
		return ErrDatacenterIDInvalid
	}

	if ev.DatacenterRegion == "" {
		return ErrDatacenterRegionInvalid
	}

	if ev.DatacenterAccessKey == "" || ev.DatacenterAccessToken == "" {
		return ErrDatacenterCredentialsInvalid
	}

	if ev.NetworkSubnet == "" {
		return ErrNetworkSubnetInvalid
	}

	return nil
}

// Error the request
func (ev *Event) Error(err error) {
	log.Printf("Error: %s", err.Error())
	ev.ErrorMessage = err.Error()

	data, err := json.Marshal(ev)
	if err != nil {
		log.Panic(err)
	}
	nc.Publish("network.delete.aws.error", data)
}

// Complete the request
func (ev *Event) Complete() {
	data, err := json.Marshal(ev)
	if err != nil {
		ev.Error(err)
	}
	nc.Publish("network.delete.aws.done", data)
}
