#!/bin/bash

DATE=$(date '+%Y-%m-%d-%H-%M-%S')
DOMAIN=$2
REGION=$4
AVAILABILITY_ZONE="$REGION" + "a"
AMI_ID=ami-09d56f8956ab235b3
INSTANCE_SIZE=t3a.small
KEY_NAME=$6

if [[ $1 == "-d" && $3 == "-r" && $5 == "-k" ]];
then
    echo "Registering domain..."
    # Register Domain
    Content="$(jq --arg domain "$DOMAIN" '.DomainName = $domain' register-domain.json)" && echo -E "${Content}" > register-domain.json

    aws route53domains register-domain --region $REGION --cli-input-json file://register-domain.json

    echo "Creating hosted zone..."
    # Create Hosted Zone for Registered Domain
    ZONE_ID=$(aws route53 create-hosted-zone --name $DOMAIN --caller-reference $DATE | jq '.HostedZone' | grep Id | grep -Eoh "[A-Z0-9]{2,}")

    echo "Changing name servers..."
    RAW_NS=$(aws route53 get-hosted-zone --id $ZONE_ID | jq '.DelegationSet' | grep -Eoh "ns-[0-9]+.awsdns-[0-9]+.[a-z]+" | cut -d " " -f 1)

    nameservers=""

    for ns in $RAW_NS
    do
        nameservers+="Name=$ns "
    done

    # Point Domain to Hosted Zone
    aws route53domains update-domain-nameservers --region $REGION --domain-name $DOMAIN --nameservers $nameservers

    echo "Creating key pairs for EC2..."
    # Create EC2 Key Pair
    aws ec2 create-key-pair --key-name $KEY_NAME --query 'KeyMaterial' --output text > $KEY_NAME.pem

    # Change Key Permissions
    sudo chmod 400 $KEY_NAME.pem

    echo "Creating EC2 Instance..."
    # Create EC2 Instance
    INSTANCE_ID=$(aws ec2 run-instances --image-id $AMI_ID --instance-type $INSTANCE_SIZE \
        --count 1 --associate-public-ip-address \
        --key-name $KEY_NAME.pem | grep InstanceId | grep -Eoh "i-[a-z0-9]+")

    # Get Instance Public Id
    EC2_PUBLIC_ID=$(aws ec2 describe-instances --instance-ids $INSTANCE_ID | grep PublicIpAddress | grep -Eoh "[0-9.]+")

    # Update Hosted Zone A Record to EC2 Public Id
    echo "Updating A Record to Point to EC2 Instance..."
    Text="$(jq \
        --arg ip "$EC2_PUBLIC_ID" \
        --arg dns "$DOMAIN" \
        '.Changes[].ResourceRecordSet.ResourceRecords = [{ Value: $ip }] | .Changes[].ResourceRecordSet.Name = $dns' \
        change-hosted-zone.json)" && echo -E "${Text}" > change-hosted-zone.json

    aws route53 change-resource-record-sets --hosted-zone-id $ZONE_ID --change-batch file://change-hosted-zone.json

    echo "Give it a few minutes and your server will be fully ready."

else
    echo "Missing either -d or -r or -k flag"
    exit 1
fi