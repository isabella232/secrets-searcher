# search-secrets

Search for sensitive information stored in one or more git repositories.

## Build

```shell script
cd ~/go/src/github.com/pantheon-systems
git clone git@github.com:pantheon-systems/search-secrets.git
cd search-secrets
make
```

## Pantheon usage

Obtain a GitHub token with `repo` access to the repositories you want to scan. To generate a new token, log into
GitHub, then go to "Settings > Developer settings > Personal access tokens > Generate new token".

Then run:

```shell script
cd _pantheon
export SECRETS_SOURCE_API_TOKEN="[GITHUB_TOKEN]"
../search-secrets --config="config.yaml"
```

The tool will create an `./output/report` directory that includes an HTML report.
