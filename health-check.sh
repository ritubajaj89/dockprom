#!/bin/bash
#YACE_STATUS="$(systemctl is-active yace.service)"
#PROM_STATUS="$(systemctl is-active prometheus.service)"
#ALERT_STATUS="$(systemctl is-active alertmanager.service)"
#P2K_STATUS="$(systemctl is-active prom2kafka.service)"
export INSTANCE=$(curl http://169.254.169.254/latest/meta-data/instance-id)
export REGION=$(curl http://169.254.169.254/latest/dynamic/instance-identity/document|grep region|awk -F\" '{print $4}')
echo "$(date)"
echo $INSTANCE
if [ "$(systemctl is-active yace.service)" != "active" ]; then	
    systemctl restart yace.service
    echo "Restarting yace service"
    sleep 5s
fi
if [ "$(systemctl is-active prometheus.service)" != "active" ]; then    
    systemctl restart prometheus.service
    echo "Restarting prometheus service"
    sleep 5s
fi
if [ "$(systemctl is-active alertmanager.service)" != "active" ]; then
    systemctl restart alertmanager.service
    echo "Restarting alermanager service"
    sleep 5s
fi
if [ "$(systemctl is-active prom2kafka.service)" != "active" ]; then
    systemctl restart prom2kafka.service
    echo "Restarting prom2kafka service"
    sleep 5s
fi
if [ "$(systemctl is-active yace.service)" = "active" ] && [ "$(systemctl is-active prometheus.service)" = "active" ] && [ "$(systemctl is-active alertmanager.service)" = "active" ] && [ "$(systemctl is-active prom2kafka.service)" = "active" ] ; then
    echo "All services are running"
else    
    echo " Service not running.... so exiting "
    aws autoscaling set-instance-health --region $REGION --instance-id $INSTANCE --health-status Unhealthy
fi
