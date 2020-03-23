#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TH_DIR="${ROOT_DIR}/truffleHog"

if [[ ! -d "${TH_DIR}" ]]; then
  git clone git@github.com:pantheon-systems/truffleHog.git "${TH_DIR}"
  cd "${TH_DIR}"
  git checkout search-secrets-tmp
  pip install -r requirements.txt
fi

cd "${TH_DIR}"

python -m truffleHog.truffleHog "$@"
