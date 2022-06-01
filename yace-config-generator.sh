#!/bin/bash
set -x
#mv /etc/yace/config.yml /etc/yace/config_$(date +%d-%m-%Y).yml
RES_FILE="/root/dockprom/resource-in-use.txt"
TEMP_FILE="/tmp/yace-config.yml"
YACE_CONF_FILE="/etc/yace/config.yml"
aws resourcegroupstaggingapi get-resources --region $(curl http://169.254.169.254/latest/dynamic/instance-identity/document|grep region|awk -F\" '{print $4}') --query 'ResourceTagMappingList[].ResourceARN'| awk -F ':' '{print $3}' | sort | uniq > $RES_FILE
LIST="$(< $RES_FILE)"
array=("alb" "apigateway" "asg" "athena" "cloudfront" "ebs" "ec" "ec2" "ecs-containerinsights" "ecs-svc" "efs" "elb" "firehose" "kinesis" "lambda" "ngw" "nlb" "rds" "redshift" "route53" "
s3" "shield" "sns" "sqs" "vpn")
if [ -f ${RES_FILE} ]
then
    if [ -s ${RES_FILE} ]
    then
      echo "File exists and is not empty"
      cat /root/dockprom/yace/config.yml >> $TEMP_FILE
      for i in $LIST; do
        if [[ " ${array[*]} " =~ $i ]]; then
          cat /root/dockprom/yace/$i.yml >> $TEMP_FILE;
        fi
      done
    else
      echo "File exists but empty"
    fi
else
    echo "File not exists"
fi
if cmp -s "$TEMP_FILE" "$YACE_CONF_FILE"; then
 echo "Files are different therefore updating configuration"   
 mv $YACE_CONF_FILE /etc/yace/config_$(date +"%Y-%m-%d-%H-%M-%S").yml
 cp $TEMP_FILE $YACE_CONF_FILE
 rm -rf $TEMP_FILE
 echo "Updated yace config for new aws resources"
 systemctl restart yace.service
else
 echo "No new resources have been added"
fi
