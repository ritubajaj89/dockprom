#!/bin/bash
set -x
ENV=$1
echo "Environment $ENV"
LEANIX_ID=$2
echo "LeanIX ID is $LEANIX_ID"
if [ $ENV == "staging" ]; then
    CONF="--kafka-brokers=kafka1.stg.techmon.emea.kaas.3stripes.net:9093,kafka2.stg.techmon.emea.kaas.3stripes.net:9093 --kafka-topic=$2.metrics --httpListenAddr=:9201 --ca-file=/etc/config/prom2kafka/certs/ca.crt --key-file=/etc/config/prom2kafka/certs/client.key --cert-file=/etc/config/prom2kafka/certs/cert.pem --batch-size=1000"   
echo OPTIONS=$CONF > /etc/sysconfig/prom2kafka
fi
if [ $ENV == "production" ]; then
    CONF="--kafka-brokers=kafka1.pro.techmon.emea.kaas.3stripes.net:9093,kafka2.pro.techmon.emea.kaas.3stripes.net:9093 --kafka-topic=$2.metrics --httpListenAddr=:9201 --ca-file=/etc/config/prom2kafka/certs/ca.crt --key-file=/etc/config/prom2kafka/certs/client.key --cert-file=/etc/config/prom2kafka/certs/cert.pem --batch-size=1000" 
echo OPTIONS=$CONF > /etc/sysconfig/prom2kafka
fi 
