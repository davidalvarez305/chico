package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/aws/aws-sdk-go-v2/service/route53domains/types"
)

func ChangeNameservers(domain, zoneId string) error {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return err
	}

	client := route53.NewFromConfig(cfg)

	input := route53.GetHostedZoneInput{
		Id: &zoneId,
	}

	output, err := client.GetHostedZone(ctx, &input)

	if err != nil {
		return err
	}

	var nameServers []types.Nameserver
	for i := 0; i < len(output.DelegationSet.NameServers); i++ {
		var name types.Nameserver
		name.Name = &output.DelegationSet.NameServers[i]
		nameServers = append(nameServers, name)
	}

	domainClient := route53domains.NewFromConfig(cfg)

	nameserversInput := route53domains.UpdateDomainNameserversInput{
		DomainName:  &domain,
		Nameservers: nameServers,
	}

	r, err := domainClient.UpdateDomainNameservers(ctx, &nameserversInput)

	if err != nil {
		return err
	}

	fmt.Printf("Successfully changed nameservers %+v", r)

	return nil
}

func CreateKeyPair(domain string) (string, error) {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return "", err
	}

	client := ec2.NewFromConfig(cfg)

	keyInput := ec2.CreateKeyPairInput{
		KeyName: &domain,
	}

	key, err := client.CreateKeyPair(ctx, &keyInput)

	if err != nil {
		return "", err
	}

	v := reflect.ValueOf(key.KeyMaterial).Elem()
	fmt.Printf("Key Material: %v\n", v)

	keyName := strings.Split(domain, ".")[0] + ".pem"

	err = ioutil.WriteFile(keyName, []byte(*key.KeyMaterial), 400)

	if err != nil {
		return "", err
	}

	keysDir, err := user.Current()

	if err != nil {
		return "", err
	}

	dst := fmt.Sprintf(keysDir.HomeDir + "/keys/" + keyName)

	err = CopyFile(keyName, dst)

	if err != nil {
		return "", err
	}

	return keyName, nil
}

func CreateEC2Instance(key string) (string, error) {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return "", err
	}

	client := ec2.NewFromConfig(cfg)

	imageId := "09d56f8956ab235b3"
	var count int32 = 1

	ins := ec2.RunInstancesInput{
		MaxCount:     &count,
		MinCount:     &count,
		ImageId:      &imageId,
		InstanceType: "t3a.small",
		KeyName:      &key,
	}

	out, err := client.RunInstances(ctx, &ins)

	if err != nil {
		return "", err
	}

	var insId string
	for i := 0; i < len(out.Instances); i++ {
		insId = *out.Instances[i].InstanceId
	}

	input := ec2.DescribeInstancesInput{
		InstanceIds: []string{insId},
	}

	o, err := client.DescribeInstances(ctx, &input)

	if err != nil {
		return "", err
	}

	var publicIpAddress string
	for i := 0; i < len(o.Reservations); i++ {
		for l := 0; l < len(o.Reservations[i].Instances); l++ {
			publicIpAddress = *o.Reservations[i].Instances[i].PublicIpAddress
		}
	}

	return publicIpAddress, nil
}

func ChangeRecordSets(zoneId, domain, instanceId string) error {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return err
	}

	client := route53.NewFromConfig(cfg)

	content, err := ioutil.ReadFile("./change-hosted-zone.json")

	if err != nil {
		log.Fatal("Error loading Change Hosted Zones file.")
	}

	var input route53types.ChangeBatch
	err = json.Unmarshal(content, &input)

	if err != nil {
		log.Fatal("Error trying to parse JSON contents into struct.")
	}

	input.Changes[0].ResourceRecordSet.Name = &domain
	input.Changes[0].ResourceRecordSet.ResourceRecords[0].Value = &instanceId

	in := route53.ChangeResourceRecordSetsInput{
		HostedZoneId: &zoneId,
		ChangeBatch:  &input,
	}

	_, err = client.ChangeResourceRecordSets(ctx, &in)

	if err != nil {
		return err
	}

	return nil
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)

	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)

	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
