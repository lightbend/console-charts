#!/usr/bin/env bash

# wait es_services are ready. polling es-prom-server/es-console/es-grafana
SERVICES=$(minikube service list)
CURL='curl --connect-timeout 1 --max-time 1 --output /dev/null --silent'
PROMETHEUS_URL=`echo "$SERVICES" | grep expose-prometheus | awk  '{print $6}'`
GRAFANA_URL=`echo "$SERVICES" | grep expose-grafana | awk  '{print $6}'`
ES_CONSOLE_URL=`echo "$SERVICES" | grep expose-es-console | awk  '{print $6}'`


# polling es-prom-server
seconds=0
limit=360

function polling_service {
  until $($CURL $1); do
    printf '.'
    sleep 2
    seconds=$((seconds+2))

    if [ $seconds == $limit ]; then
      echo "poll $1 timeout"
      exit 1
    fi
  done
  echo "$1 is up"
}

polling_service $PROMETHEUS_URL
polling_service $ES_CONSOLE_URL
polling_service $GRAFANA_URL
echo "poll services takes $seconds sec"
