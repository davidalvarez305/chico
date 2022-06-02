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

	content, err := ioutil.ReadFile("./register-domain.json")

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

func LaunchServer(domain, zoneId string) {
	err := utils.ChangeNameservers(domain, zoneId)

	if err != nil {
		log.Fatalf("Failed getting nameservers: %v\n", err)
	}

	key, err := utils.CreateKeyPair(domain)

	if err != nil {
		log.Fatalf("Failed creating key pair: %v\n", err)
	}

	instanceId, err := utils.CreateEC2Instance(key)

	if err != nil {
		log.Fatalf("Failed creating key pair: %v\n", err)
	}

	fmt.Println(instanceId)
}
