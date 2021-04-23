#!/usr/bin/env bash

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

NS=${NAMESPACE:-lightbend}

# wait es_services are ready. polling es-prom-server/es-console/es-grafana
CURL='curl --connect-timeout 1 --max-time 1 --output /dev/null --silent'
ES_CONSOLE_URL=$(minikube service --url expose-es-console -n "$NS")
echo "ES-Console URL: ${ES_CONSOLE_URL}"

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

polling_service $ES_CONSOLE_URL
echo "poll services takes $seconds sec"
