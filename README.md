# secrets-searcher

Search for sensitive information stored in one or more git repositories.

This repository was inspired by [truffleHog](https://github.com/trufflesecurity/truffleHog) and [shhgit](https://github.com/eth0izzle/shhgit).

## Build

```shell script
cd ~/go/src/github.com/pantheon-systems
git clone git@github.com:pantheon-systems/secrets-searcher.git
cd secrets-searcher
make
```

## Pantheon usage

Obtain a GitHub token with `repo` access to the repositories you want to scan. To generate a new token, log into
GitHub, then go to "Settings > Developer settings > Personal access tokens > Generate new token".

Then run:

```shell script
cd _pantheon
export SECRETS_SOURCE_API_TOKEN="[GITHUB_TOKEN]"
../secrets-searcher --config="config.yaml,config.rules.yaml"
```

The tool will create an `./output/report` directory that includes an HTML report.
