#!/bin/bash

# Adapted for this project from cert-manager
# License: https://github.com/jetstack/cert-manager/blob/master/LICENSE
# Script: https://github.com/jetstack/cert-manager/blob/master/hack/verify-deploy-gen.sh

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..

DIFFROOT="${SCRIPT_ROOT}/deploy"
TMP_DIFFROOT="${SCRIPT_ROOT}/_tmp/deploy/manifests"
_tmp="${SCRIPT_ROOT}/_tmp"

cleanup() {
  rm -rf "${_tmp}"
}
trap "cleanup" EXIT SIGINT

cleanup

mkdir -p "${TMP_DIFFROOT}"
cp -a "${DIFFROOT}"/* "${TMP_DIFFROOT}"

"${SCRIPT_ROOT}/hack/gen-deploy.sh"
echo "diffing ${DIFFROOT} against freshly deploy-gen"
ret=0
diff -Naupr "${DIFFROOT}" "${TMP_DIFFROOT}" || ret=$?
cp -a "${TMP_DIFFROOT}"/* "${DIFFROOT}"
if [[ $ret -eq 0 ]]
then
  echo "${DIFFROOT} up to date."
else
  echo "${DIFFROOT} is out of date. Please run hack/gen-deploy.sh"
  exit 1
fi