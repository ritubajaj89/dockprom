#!/bin/bash
set -x
aws resourcegroupstaggingapi get-resources --region eu-west-1 --query 'ResourceTagMappingList[].ResourceARN'| awk -F ':' '{print $3}' | sort | uniq > /home/ec2-user/resource-in-use.txt
LIST="$(< /home/ec2-user/resource-in-use.txt)"
array=("alb" "apigateway" "asg" "athena" "cloudfront" "ebs" "ec" "ec2" "ecs-containerinsights" "ecs-svc" "efs" "elb" "firehose" "kinesis" "lambda" "ngw" "nlb" "rds" "redshift" "route53" "
s3" "shield" "sns" "sqs" "vpn")
for i in $LIST; do
if [[ " ${array[*]} " =~ $i ]]; then
           cat /home/ec2-user/dockprom/yace/$i.yml >> /home/ec2-user/dockprom/yace/config.yml;
   echo "$i.yml has been added to config.yml";
        fi
done