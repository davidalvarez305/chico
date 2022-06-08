package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/aws/aws-sdk-go-v2/service/route53domains/types"
	project "github.com/davidalvarez305/chico/types"
)

func GetZoneId(domain string) (string, error) {
	var zoneId string
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return zoneId, err
	}

	client := route53.NewFromConfig(cfg)

	var count int32 = 100

	input := route53.ListHostedZonesInput{
		MaxItems: &count,
	}

	o, err := client.ListHostedZones(ctx, &input)

	for i := 0; i < len(o.HostedZones); i++ {
		if strings.Contains(*o.HostedZones[i].Name, domain) {
			zoneId = *o.HostedZones[i].Id
		}
	}

	return zoneId, nil
}

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
	var keyName string
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return keyName, err
	}

	client := ec2.NewFromConfig(cfg)

	keyName = strings.Split(domain, ".")[0] + ".pem"

	keyInput := ec2.CreateKeyPairInput{
		KeyName: &keyName,
	}

	key, err := client.CreateKeyPair(ctx, &keyInput)

	if err != nil {
		return keyName, err
	}

	v := reflect.ValueOf(key.KeyMaterial).Elem()
	fmt.Printf("Key Material: %v\n", v)

	err = ioutil.WriteFile(keyName, []byte(*key.KeyMaterial), 400)

	if err != nil {
		return keyName, err
	}

	keysDir, err := user.Current()

	if err != nil {
		return keyName, err
	}

	src := fmt.Sprintf(keysDir.HomeDir + "/chico/" + keyName)
	dst := fmt.Sprintf(keysDir.HomeDir + "/keys/" + keyName)

	err = CopyFile(src, dst)

	if err != nil {
		return keyName, err
	}

	return keyName, nil
}

func CreateEC2Instance(key string) (string, error) {
	var publicIpAddress string
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return publicIpAddress, err
	}

	client := ec2.NewFromConfig(cfg)

	imageId := "ami-09d56f8956ab235b3"
	var count int32 = 1

	ins := ec2.RunInstancesInput{
		MaxCount:     &count,
		MinCount:     &count,
		ImageId:      &imageId,
		InstanceType: "t3a.small",
		KeyName:      &key,
	}

	fmt.Println("Purchasing EC2 Instance...")

	out, err := client.RunInstances(ctx, &ins)

	if err != nil {
		return publicIpAddress, err
	}
	fmt.Println("Successfully Purchased EC2 Instance...")

	var insId string
	for i := 0; i < len(out.Instances); i++ {
		insId = *out.Instances[i].InstanceId
	}

	input := ec2.DescribeInstancesInput{
		InstanceIds: []string{insId},
	}

	time.Sleep(180 * time.Second)

	o, err := client.DescribeInstances(ctx, &input)

	if err != nil {
		return publicIpAddress, err
	}

	publicIpAddress = *o.Reservations[0].Instances[0].PublicIpAddress
	return publicIpAddress, nil
}

func ChangeRecordSets(zoneId, domain, publicIp string) error {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return err
	}

	client := route53.NewFromConfig(cfg)

	comp, err := user.Current()

	if err != nil {
		return err
	}

	file := fmt.Sprintf(comp.HomeDir + "/chico/change-hosted-zone.json")

	content, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal("Error loading Change Hosted Zones file.")
	}

	var input route53types.ChangeBatch
	err = json.Unmarshal(content, &input)

	if err != nil {
		log.Fatal("Error trying to parse JSON contents into struct.")
	}

	subDomain := "www." + domain
	input.Changes[0].ResourceRecordSet.Name = &domain
	input.Changes[1].ResourceRecordSet.Name = &subDomain
	input.Changes[0].ResourceRecordSet.ResourceRecords[0].Value = &publicIp
	input.Changes[1].ResourceRecordSet.ResourceRecords[0].Value = &publicIp

	fmt.Println("Successfully changed record sets on JSON File.")

	in := route53.ChangeResourceRecordSetsInput{
		HostedZoneId: &zoneId,
		ChangeBatch:  &input,
	}

	_, err = client.ChangeResourceRecordSets(ctx, &in)

	if err != nil {
		return err
	}

	fmt.Println("Successfully Changed Record Sets for Hosted Zone.")

	return nil
}

func transformSiteName(siteName string) string {
	return strings.Join(strings.Split(siteName, "-"), " ")
}

func PrepareServer(key, publicId, domain, db, siteName string) error {
	user := os.Getenv("DB_USER")
	s3Bucket := os.Getenv("AWS_S3_BUCKET")
	websiteName := transformSiteName(siteName)
	fmt.Println("Copying key to server...")
	copyKeyCmd := fmt.Sprintf("scp -r -i %s ./cli/prep ubuntu@%s:/home/ubuntu/", key, publicId)
	_, err := exec.Command("/bin/bash", "-c", copyKeyCmd).Output()

	if err != nil {
		return nil
	}

	fmt.Println("Preparing server...")
	prepareServerCmd := fmt.Sprintf(`ssh -i %s ubuntu@%s "chmod +x ./cli/prep/server.sh && sudo ./cli/prep/server.sh %s %s %s %s %s %s"`, key, publicId, db, user, s3Bucket, domain, websiteName, publicId)
	_, err = exec.Command("/bin/bash", "-c", prepareServerCmd).Output()

	if err != nil {
		return nil
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

func SecureCopy(keyName, ip, projectName string) {
	var username = os.Getenv("SERVER_USER")
	var keysFolder = os.Getenv("KEYS_FOLDER")
	var folder = os.Getenv("ENV_FOLDER")
	var prepFolder = os.Getenv("PREP_FOLDER")

	dirPath := folder + "/" + projectName
	scpCmd := fmt.Sprintf("scp -r -i %s %s@%s:%s %s", keysFolder+keyName, username, ip, prepFolder, dirPath)

	_, err := exec.Command("/bin/bash", "-c", scpCmd).Output()

	if err != nil {
		fmt.Println("Error while copying file ", err)
	}
}

func DeployProject(project project.Project) error {
	username := os.Getenv("SERVER_USER")
	startDocker := "sudo docker-compose -f docker-compose.yml down && sudo docker-compose -f ~/client_template/docker-compose.yml up --build"
	cloneGithubRepo := "sudo rm -r soflo_node && git clone " + project.Repo
	keysFolder := os.Getenv("KEYS_FOLDER")

	cmd := fmt.Sprintf(`ssh -i %s %s@%s "cd && %s && %s"`, keysFolder+project.Key, project.IP, username, cloneGithubRepo, startDocker)
	_, err := exec.Command("/bin/bash", "-c", cmd).Output()

	if err != nil {
		return err
	}

	return nil
}
