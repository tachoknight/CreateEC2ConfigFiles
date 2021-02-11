package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// The fields we want to capture for writing out to the
// config file
type configData struct {
	keyName          string
	hostName         string
	publicIPAddress  string
	privateIPAddress string
	user             string
}

func main() {
	// We take two command-line arguments, "key" and "value" as
	// these are going to be passed to the EC2 filter for querying
	keyPtr := flag.String("key", "xyzzy", "EC2 Tag 'Key' (Note: case sensitive)")
	valuePtr := flag.String("value", "xyzzy", "EC2 Tag 'Value' (Note: case sensitive)")
	regionPtr := flag.String("region", "xyzzy", "EC2 region")
	flag.Parse()

	// Standard AWS SDK stuff here...
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(*regionPtr)},
	)

	svc := ec2.New(sess)
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(fmt.Sprintf("tag:%s", *keyPtr)),
				Values: []*string{
					aws.String(*valuePtr),
				},
			},
		},
	}

	result, err := svc.DescribeInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return
	}

	var instances []configData
	var publicServerName string

	for _, r := range result.Reservations {
		// Get the individual EC2 instance...
		var i = *&r.Instances[0]
		// And the fields we need
		var cd configData

		// The key needed to log into the instance
		cd.keyName = *i.KeyName
		// The name of the instance is hidden in the tags array
		cd.hostName = func(tags []*ec2.Tag) string {
			for _, v := range tags {
				if *v.Key == "Name" {
					return *v.Value
				}
			}
			return "NAME TAG MISSING"
		}(*&i.Tags)
		// Instances always have a private IP
		cd.privateIPAddress = *i.PrivateIpAddress
		// Does it have a public IP?
		if len(*i.PublicDnsName) > 0 {
			// yes it does!
			cd.publicIPAddress = *i.PublicIpAddress
			// This is the public instance that all
			// the other instances are hidden behind;
			// we store this off now so we don't have to
			// go looking for it later
			publicServerName = cd.hostName
		}

		instances = append(instances, cd)
	}

	///////////////////////////////////////////////////////////////////////////
	// Now let's put together the config file. We are making a few
	// assumptions:
	//		1. The user is ec2-user; if you use a non-Amazon AMI, it'll probably
	//		   be something like "fedora" or "ubuntu".
	//		2. If the instance has a public IP, it's a bastion server and all the
	//		   other instances will use that as a proxy
	//		3. The key resides wherever keyLocation is set to

	keyLocation := "~/.ssh/"

	hostNames := make(map[string]int)

	for _, i := range instances {
		// First let's see if there's already a key in the dictionary
		// with this hostname, in which case we need to append the
		// dictionary value + 1
		num, ok := hostNames[i.hostName]
		num = num + 1
		if ok {
			fmt.Printf("Host \"%s\"\n", fmt.Sprintf("%s-%d", i.hostName, num))
			hostNames[i.hostName] = num
		} else {
			fmt.Printf("Host \"%s\"\n", i.hostName)
			hostNames[i.hostName] = num
		}

		if len(i.publicIPAddress) > 0 {
			fmt.Printf("\tHostName %s\n", i.publicIPAddress)
		} else {
			fmt.Printf("\tHostName %s\n", i.privateIPAddress)
		}
		fmt.Printf("\tUser %s\n", "ec2-user")
		fmt.Printf("\tStrictHostKeyChecking no\n")
		fmt.Printf("\tUserKnownHostsFile=/dev/null\n")
		fmt.Printf("\tIdentityFile %s%s.pem\n", keyLocation, i.keyName)
		if len(i.publicIPAddress) == 0 {
			fmt.Printf("\tProxyCommand ssh -W %%h:%%p %s 2> /dev/null\n", publicServerName)
		}

		fmt.Printf("\n")
	}
}
