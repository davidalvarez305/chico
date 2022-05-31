package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
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

	fmt.Printf("%s", string(content))
	var input route53domains.RegisterDomainInput
	err2 := json.Unmarshal(content, &input)

	if err2 != nil {
		log.Fatal("Error trying to parse JSON contents into struct.")
	}

	in := route53domains.CheckDomainAvailabilityInput{
		DomainName: aws.String(domain),
	}

	check, err := client.CheckDomainAvailability(ctx, &in)

	if err != nil {
		log.Fatal("Error loading Register Domain file.")
	}

	if check.Availability == "AVAILABLE" {
		// client.RegisterDomain(ctx, input)
	} else {
		log.Fatal("Domain is not available.")
	}
}
