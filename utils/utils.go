package utils

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

func GetNameServers(zoneId string) (string, error) {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		log.Fatalf("Couldn't load AWS SDK config, %v\n", err)
		return "", err
	}

	client := route53.NewFromConfig(cfg)

	input := route53.GetHostedZoneInput{
		Id: &zoneId,
	}

	output, err := client.GetHostedZone(ctx, &input)

	if err != nil {
		log.Fatalf("Couldn't get hosted zone: %v\n", err)
		return "", err
	}

	var nameServers string
	for i := 0; i < len(output.DelegationSet.NameServers); i++ {
		nameServers += "Name=" + output.DelegationSet.NameServers[i] + " "
	}
	return nameServers, nil
}
