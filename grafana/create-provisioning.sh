#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

jq -n --arg DS_PROMETHEUS prometheus_ccloudexporter -f <(sed 's/"${DS_PROMETHEUS}"/$DS_PROMETHEUS/g' <ccloud-exporter.json) >provisioning/dashboards/ccloud-exporter.json
