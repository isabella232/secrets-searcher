# search-secrets

Search for sensitive information stored in one or more git repositories.

## Install

Run this from the root directory:

```shell script
go build .
```

## Usage

Run this way:

```shell script
search-secrets --github-token="$GITHUB_TOKEN" --config-file="config.pantheon.yaml"
```

You will find the output at `./output`.

## TODO

- Make Github an optional dependency
