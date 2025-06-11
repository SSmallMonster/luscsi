#!/bin/bash
set -e

# shellcheck disable=SC2046
PROGRAM=$(cd $(dirname "${BASH_SOURCE[0]}")/../../; pwd)
HELM_LUSCSI_DIR=${PROGRAM}/luscsi/deploy/luscsi

# render values.yaml according chart version or release tag
function render_image_tag() {
  local imageTag=$1
  sed -i "s/tag: \"\"/tag: ${imageTag}/g" "${HELM_LUSCSI_DIR}/values.yaml"
}

function get_chart_version() {
   helm show chart "${HELM_LUSCSI_DIR}"|grep version|awk '{print $2}'
}

render_image_tag "$(get_chart_version)"