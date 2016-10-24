/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	ecc "github.com/ernestio/ernest-config-client"
	"github.com/nats-io/nats"
)

var nc *nats.Conn
var natsErr error

func eventHandler(m *nats.Msg) {
	var n Event

	err := n.Process(m.Data)
	if err != nil {
		return
	}

	if err = n.Validate(); err != nil {
		n.Error(err)
		return
	}

	err = deleteNetwork(&n)
	if err != nil {
		n.Error(err)
		return
	}

	n.Complete()
}

func getNetworkInterfaces(svc *ec2.EC2, networkID string) (*ec2.DescribeNetworkInterfacesOutput, error) {
	f := []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("subnet-id"),
			Values: []*string{aws.String(networkID)},
		},
	}

	req := ec2.DescribeNetworkInterfacesInput{
		Filters: f,
	}

	return svc.DescribeNetworkInterfaces(&req)
}

func waitForInterfaceRemoval(svc *ec2.EC2, networkID string) error {
	for {
		resp, err := getNetworkInterfaces(svc, networkID)
		if err != nil {
			return err
		}

		if len(resp.NetworkInterfaces) == 0 {
			return nil
		}

		time.Sleep(time.Second)
	}
}

func deleteNetwork(ev *Event) error {
	creds := credentials.NewStaticCredentials(ev.DatacenterAccessKey, ev.DatacenterAccessToken, "")
	svc := ec2.New(session.New(), &aws.Config{
		Region:      aws.String(ev.DatacenterRegion),
		Credentials: creds,
	})

	err := waitForInterfaceRemoval(svc, ev.NetworkAWSID)
	if err != nil {
		return err
	}

	req := ec2.DeleteSubnetInput{
		SubnetId: aws.String(ev.NetworkAWSID),
	}

	_, err = svc.DeleteSubnet(&req)

	return err
}

func main() {
	nc = ecc.NewConfig(os.Getenv("NATS_URI")).Nats()

	fmt.Println("listening for network.delete.aws")
	nc.Subscribe("network.delete.aws", eventHandler)

	runtime.Goexit()
}
