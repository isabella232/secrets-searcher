#!/usr/bin/env bash

set -euo pipefail
trap "exit" INT

function main() {
  mkdir -p "data"

  local repos_raw_file="data/repos-raw.json"
  local repos_names_file="data/repos-names.json"

  # Pull a list of repos from GitHub
  echo "== Pulling repo data from GitHub"
  curl -s "https://api.github.com/orgs/pantheon-systems/repos?access_token=${GITHUB_ACCESS_TOKEN}" >"$repos_raw_file"
  jq -r '.[] | .name' "$repos_raw_file" | sort >"$repos_names_file"
  echo "OK"

  # For each repo, run a search and output `data/secrets-{repo_name}.json`
  while read -r repo_name || [[ -n $repo_name ]]; do
    echo "== Processing ${repo_name}"

    local secrets_raw_file="data/secrets-${repo_name}-raw.json"
    local secrets_human_file="data/secrets-${repo_name}.json"

    # Search for secrets
    # Note: Change `false` to `true` to temporarily skip searched repos during development
    if false && [[ -f "$secrets_raw_file" ]]; then
      echo "Repo already processed, skipping search"
    else
      echo "Searching ${repo_name} ..."
      truffleHog --regex --entropy=False --json "git@github.com:pantheon-systems/${repo_name}.git" >"${secrets_raw_file}.tmp" || true
      mv "${secrets_raw_file}.tmp" "$secrets_raw_file"
    fi

    # If there weren't any secrets found, delete the data file and continue
    if [ ! -s "$secrets_raw_file" ]; then
      echo "No secrets found"
      rm "$secrets_raw_file"
      continue
    fi

    # How many secrets were found?
    count=$(jq -c . "$secrets_raw_file" | wc -l)
    count=${count##* }
    echo "${count} secrets found"

    # Generate a human readable data file
    jq -r '. | del(.diff) | del(.printDiff)' "$secrets_raw_file" >"$secrets_human_file"

  done <"$repos_names_file"
}

main
