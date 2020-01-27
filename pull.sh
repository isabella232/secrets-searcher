#!/usr/bin/env bash

set -euo pipefail
trap "exit" INT

function main() {
  mkdir -p "data"

  local repos_raw_file="data/repos-raw.json"
  local repos_names_file="data/repos-names.json"

  echo "== Pulling repo data from GitHub"
  curl -s "https://api.github.com/orgs/pantheon-systems/repos?access_token=${GITHUB_ACCESS_TOKEN}" >"$repos_raw_file"
  jq -r '.[] | .name' "$repos_raw_file" | sort >"$repos_names_file"
  echo "OK"

  while read -r repo_name || [[ -n $repo_name ]]; do
    echo "== Processing ${repo_name}"

    local secrets_raw_file="data/secrets-${repo_name}-raw.json"
    local secrets_human_file="data/secrets-${repo_name}.json"

    # Change `false` to `true` to temporarily skip searched files during development
    if false && [[ -f "$secrets_raw_file" ]]; then
      echo "Repo already processed, skipping search"
    else
      echo "Searching ${repo_name} ..."
      truffleHog --regex --entropy=False --json "git@github.com:pantheon-systems/${repo_name}.git" >"${secrets_raw_file}.tmp" || true
      mv "${secrets_raw_file}.tmp" "$secrets_raw_file"
    fi

    if [ ! -s "$secrets_raw_file" ]; then
      echo "No secrets found"
      rm "$secrets_raw_file"
      continue
    fi

    count=$(jq -c . "$secrets_raw_file" | wc -l)
    count=${count##* }
    echo "${count} secrets found"

    jq -r '. | del(.diff) | del(.printDiff)' "$secrets_raw_file" >"$secrets_human_file"

  done <"$repos_names_file"
}

main
