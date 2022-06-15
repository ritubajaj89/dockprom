#!/bin/bash
set -x
# Set location variables:
tmp_file_location=/root/config-tmp
prom_config_location=/etc/config/prometheus
am_config_location=/etc/config/alertmanager
yace_config_location=/root/dockprom/yace
# Create new directory to store today's messages:
date=$(date +"%Y-%m-%d-%H-%M-%S")
[[ -d ${tmp_file_location}/$date ]] || mkdir ${tmp_file_location}/$date
# Retrieve new messages from S3 and save to tmpemails/ directory:
aws s3 cp --recursive s3://config-test-metrics/ $tmp_file_location/$date
cd $tmp_file_location/$date/prometheus
prom_list=`find . -type f -printf "%f\n"`

for a in $prom_list; do
   if [ ! -f "$prom_config_location/$a" ]; then
        cp $a $prom_config_location/$a
        curl -s -XPOST localhost:9090/-/reload
        echo "Config reloaded after new file"
      continue
   fi
   diff $a $prom_config_location/$a > /dev/null
   if [[ "$?" == "1" ]]; then
        # File exists but is different so copy changed file
        mv $prom_config_location/$a $prom_config_location/$a_bkp_$date
        cp $a $prom_config_location
        promtool check config $prom_config_location/prometheus.yml | grep -i "FAILED" 
        if [[ "$?" == "1" ]]; then
        	echo "Configuration file is invalid"
        fi	
        curl -s -XPOST localhost:9090/-/reload 
        echo "Config reloaded after prm"
   fi
done