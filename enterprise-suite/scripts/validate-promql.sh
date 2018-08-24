#!/usr/bin/env bash

set -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

if ! command -v promtool > /dev/null; then
    echo "please install promtool"
    exit 1
fi

if ! command -v jq > /dev/null; then
    echo "please install jq"
    exit 1
fi

echo "checking promql in enterprise-suite/es-grafana/*.json"

# check_expr wraps one promql expression in a mock rule file for promtool.
# We test one expression at a time to more easily report the expression that
# failed, as promtool will report line number problems instead of expr bugs.
check_expr() {
  echo "$*"
  printf 'groups:\n- name: group\n  rules:\n  - record: metric\n    expr: %s\n' "$*" | promtool check rules /dev/stdin 2>&1
}

# extract our grafana plugin expressions as a virtual file, one row per expr
prom_lines() {
  jq -r '.[].promQL[]' "$1" | sed -e 's/ContextTags[,]*//g'
}

count=0
for json in ${script_dir}/../es-grafana/*.json; do
  prom_lines "$json" | while read -r promql; do
    out=$( check_expr "$promql" ) || {
      echo "error in $json"
      echo "$out"
      exit 1
    }
  done

  count=$(( count + $(prom_lines "$json"| wc -l) ))
done

echo "validated $count promql expressions"

echo "checking enterprise-suite/es-monitor-api/static-rules.yml"
cat ${script_dir}/../es-monitor-api/static-rules.yml |
  ( printf 'groups:\n- name: group\n  rules:\n' ; sed -e 's/^/  /') |
  promtool check rules /dev/stdin
