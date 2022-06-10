package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/davidalvarez305/chico/utils"
)

func PurchaseDomain(domain string) {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		log.Fatalf("Couldn't load AWS SDK config, %v", err)
	}

	client := route53domains.NewFromConfig(cfg)

	path, err := utils.ResolvePath("register-domain.json")

	if err != nil {
		log.Fatal("Failed resolving path to file\n", err)
	}

	content, err := ioutil.ReadFile(path)

	if err != nil {
		log.Fatal("Error loading Register Domain file.")
	}

	var input route53domains.RegisterDomainInput
	err = json.Unmarshal(content, &input)

	if err != nil {
		log.Fatal("Error trying to parse JSON contents into struct.")
	}

	input.DomainName = aws.String(domain)
	v := reflect.ValueOf(input.DomainName).Elem()
	fmt.Printf("Domain to purchase: %v\n", v)

	in := route53domains.CheckDomainAvailabilityInput{
		DomainName: aws.String(domain),
	}

	check, err := client.CheckDomainAvailability(ctx, &in)
	fmt.Printf("Checking availability of %v...\n", v)

	if err != nil {
		log.Fatal("Error loading Register Domain file.")
	}

	if check.Availability == "AVAILABLE" {
		fmt.Printf("%v is %s!\n", v, check.Availability)
		_, err := client.RegisterDomain(ctx, &input)

		if err != nil {
			log.Fatal("Failed to register domain: %v\n", err)
		}
	} else {
		log.Fatal("Domain is not available.")
	}
}

func LaunchServer(domain, db, siteName string) {

	zoneId, err := utils.GetZoneId(domain)

	if err != nil {
		log.Fatalf("Failed Getting Zone ID: %v\n", err)
	}

	fmt.Printf("Zone ID: %s\n", zoneId)

	err = utils.ChangeNameservers(domain, zoneId)

	if err != nil {
		log.Fatalf("Failed getting nameservers: %v\n", err)
	}

	key, err := utils.CreateKeyPair(domain)

	if err != nil {
		log.Fatalf("Failed Creating Key Pair: %v\n", err)
	}

	publicIp, err := utils.CreateEC2Instance(key)

	if err != nil {
		log.Fatalf("Failed Purchasing EC2 Instance: %v\n", err)
	}

	err = utils.ChangeRecordSets(zoneId, domain, publicIp)

	if err != nil {
		log.Fatalf("Failed Changing Record Sets: %v\n", err)
	}

	err = utils.PrepareServer(key, publicIp, domain, db, siteName)

	if err != nil {
		log.Fatalf("Failed Changing Record Sets: %v\n", err)
	}
}
