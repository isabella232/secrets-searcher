#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TH_DIR="${ROOT_DIR}/truffleHog"
VENV_DIR="${TH_DIR}/venv"

if [[ ! -d "${TH_DIR}" ]]; then
  git clone git@github.com:pantheon-systems/truffleHog.git "${TH_DIR}"
  git checkout --git-dir="${TH_DIR}/.git" search-secrets-tmp
  python -m venv "${VENV_DIR}"
  "${TH_DIR}/venv/bin/pip" install -r "${TH_DIR}/requirements.txt"
fi

cd "${TH_DIR}"
"${TH_DIR}/venv/bin/python" -m truffleHog.truffleHog "$@"
