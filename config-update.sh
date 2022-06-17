#!/bin/bash
set -x
# Set location variables:
tmp_file_location=/root/config-tmp
prom_config_location=/etc/config/prometheus
yace_config_location=/root/dockprom/yace
# Create new directory to store today's messages:
date=$(date +"%Y-%m-%d-%H-%M-%S")
[[ -d ${tmp_file_location}/$date ]] || mkdir ${tmp_file_location}/$date
# Retrieve new messages from S3 and save to tmpemails/ directory:
aws s3 cp --recursive s3://config-test-metrics/ $tmp_file_location/$date
# Checking configuration for prometheus
cd $tmp_file_location/$date/prometheus
prom_list=`find . -type f -printf "%f\n"`
for a in $prom_list; do
   if [ ! -f "$prom_config_location/$a" ]; then
        cp $a $prom_config_location/$a
        curl -s -XPOST localhost:9090/-/reload
        echo "Config reloaded after new file"
      continue
   fi
   if [ -f "$prom_config_location/$a" ]; then
	   diff $a $prom_config_location/$a > /dev/null
  	   if [[ "$?" == "1" ]]; then
    	    # File exists but is different so copy changed file
            if [[ "$a" == "prometheus.yml" ]]; then
                promtool check config $prom_config_location/prometheus.yml | grep -i "FAILED" 
                if [[ "$?" == "0" ]]; then
                    echo "Configuration file is invalid"
                fi  
            fi 
            break    
            mv "$prom_config_location/$a" "$prom_config_location/$a"_bkp_"$date"
            cp $a $prom_config_location 
            curl -s -XPOST localhost:9090/-/reload 
            echo "Config reloaded successfully"
    	fi
   continue      
   fi	
done
# # Checking configuration for alertmanager
# cd $tmp_file_location/$date/alertmanager
# diff alertmanager.yml $prom_config_location/alertmanager.yml > /dev/null
#    if [[ "$?" == "1" ]]; then
#         # File exists but is different so copy changed file
#         mv $prom_config_location/alertmanager.yml $prom_config_location/alertmanager.yml_bkp_$date
#         cp alertmanager.yml $prom_config_location
# 		  amtool check-config /etc/config/prometheus/alertmanager.yml | grep -i 'SUCCESS'
#         if [[ "$?" == "1" ]]; then
#         	echo "Configuration file is invalid"
#         fi	
#         systemctl restart alertmanager 
#         echo "Alertmanager Service has been restarted"
#     continue  
#    fi		
# # Checking configuration for yace
# cd $tmp_file_location/$date/yace
# yace_list=`find . -type f -printf "%f\n"`
# for i in $yace_list; do
#    if [ ! -f "$yace_config_location/$i" ]; then
#         cp $i $yace_config_location
#       continue
#    fi
#    if [ -f "$yace_config_location/$i" ]; then
# 	   diff $i $yace_config_location/$i > /dev/null
#   	   if [[ "$?" == "1" ]]; then
#     	    # File exists but is different so copy changed file
#         	mv $yace_config_location/$i $yace_config_location/$i_bkp_$date
#         	cp $i $yace_config_location       	
#         	echo "Config will be reloaded on next 10 min"
#    	   fi
#     continue   
#    fi	
# done
