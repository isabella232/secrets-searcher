#!/usr/bin/env bash

# pull.sh - A script to detect secrets in every repo of a GitHub organization

set -euo pipefail
trap "exit" INT

# Constants
readonly REPOS_RAW_FILE="data/repos-raw.json"

#######################################
# Run a search on a repo and output `data/secrets-{repo_name}.json`
# Globals:
#   SECRETS_FORCE_SEARCH
#   SECRETS_ORG
# Arguments:
#   Repo name (ex. "titan-mt")
# Returns:
#   None
#######################################
pull_secrets() {
  local repo_name="$1"
  local secrets_raw_file="data/secrets-${repo_name}-raw.json"
  local secrets_human_file="data/secrets-${repo_name}.json"

  echo "== Processing ${repo_name}"

  if [[ -f "$secrets_raw_file" ]] && [[ "$SECRETS_FORCE_SEARCH" == true ]]; then
    rm "$secrets_raw_file"
  fi

  # Search for secrets and create raw file
  if [[ -f "$secrets_raw_file" ]]; then
    echo "Repo already processed and saved to ${secrets_raw_file} ..."
  else
    echo "Searching for secrets, outputting to ${secrets_raw_file} ..."
    truffleHog --regex --entropy=False --json "git@github.com:${SECRETS_ORG}/${repo_name}.git" >"${secrets_raw_file}.tmp" || true
    mv "${secrets_raw_file}.tmp" "$secrets_raw_file"
  fi

  # Generate a human readable file and count the secrets
  if [ -s "$secrets_raw_file" ]; then
    jq -r '. | del(.diff) | del(.printDiff)' "$secrets_raw_file" >"$secrets_human_file"

    count=$(jq -c . "$secrets_raw_file" | wc -l)
    count=${count##* }
  else
    # No secrets found, don't generate a human readable file
    count=0
  fi

  echo "${count} secrets found"
}

#######################################
# Pulls a list of repos from GitHub.
# Globals:
#   SECRETS_GITHUB_ACCESS_TOKEN
#   SECRETS_ORG
# Arguments:
#   None
# Returns:
#   None
#######################################
pull_repos() {
  echo "== Saving repo list to ${REPOS_RAW_FILE}"
  curl -s "https://api.github.com/orgs/${SECRETS_ORG}/repos?access_token=${SECRETS_GITHUB_ACCESS_TOKEN}" >"$REPOS_RAW_FILE"
  echo "OK"
}

#######################################
# Pulls a list of repos from GitHub, searches for secrets in each repo, outputs to `data` directory.
# Globals:
#   None
# Arguments:
#   None
# Returns:
#   None
#######################################
pull() {
  mkdir -p "data"
  find data -name "*.tmp" -exec rm {} \;

  pull_repos

  while read -r repo_name; do
    pull_secrets "$repo_name"
  done < <(jq -r '.[] | .name' "$REPOS_RAW_FILE" | sort)
}

#######################################
# Execute script.
# Globals:
#   None
# Arguments:
#   None
# Returns:
#   None
#######################################
main() {
  pull
}

main
