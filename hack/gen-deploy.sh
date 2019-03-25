#!/bin/bash

# Adapted for this project from cert-manager
# License: https://github.com/jetstack/cert-manager/blob/master/LICENSE
# Script: https://github.com/jetstack/cert-manager/blob/master/hack/update-deploy-gen.sh
set -euo pipefail
IFS=$'\n\t'

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")
REPO_ROOT="${SCRIPT_ROOT}/.."

gen() {
	OUTPUT=$1
    VALUES_PATH=""
	TMP_OUTPUT=$(mktemp)
	mkdir -p "$(dirname ${OUTPUT})"
    VALUES=${2:-}
    if [[ ! -z "$VALUES" ]]; then
        VALUES_PATH="--values=${SCRIPT_ROOT}/deploy/values/${VALUES}.yaml"
    fi
    helm template \
        "${REPO_ROOT}/charts/k8s-cloudwatch-adapter/" \
        --namespace "custom-metrics" \
        --name "k8s-cloudwatch-adapter" \
        --set "fullnameOverride=k8s-cloudwatch-adapter" \
		--set "createNamespaceResource=true" > "${TMP_OUTPUT}" \
        ${VALUES_PATH}
    
	mv "${TMP_OUTPUT}" "${OUTPUT}"
}

# Additional deployment files can be generated here with the following format:
#   gen /path/to/deployment-file.yaml [values.yaml file name in deploy/values]
#   gen ${REPO_ROOT}/deploy/adapter.yaml the-values 
gen "${REPO_ROOT}/deploy/adapter.yaml"
