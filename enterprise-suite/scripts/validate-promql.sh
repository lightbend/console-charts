#!/usr/bin/env bash

set -e

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

if ! command -v promtool > /dev/null; then
    echo "please install promtool"
    exit 1
fi

echo "checking enterprise-suite/console-api/static-rules.yml"
cat ${script_dir}/../console-api/static-rules.yml |
  ( printf 'groups:\n- name: group\n  rules:\n' ; sed -e 's/^/  /') |
  promtool check rules /dev/stdin
