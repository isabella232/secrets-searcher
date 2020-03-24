# search-secrets

Search for sensitive information stored in Pantheon git repositories.

## Usage

Run from within the source directory (for now):

```shell script
search-secrets search \
  --github-token="$GITHUB_TOKEN" \
  --organization=pantheon-systems \
  --repos=titan-mt
```

You will find the output at `./output`.

## TODO

- Make the install and usage less complex.
