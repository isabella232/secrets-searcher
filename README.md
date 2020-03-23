# searchsecrets

Search for sensitive information stored in Pantheon git repositories.

## Usage

Clone:

```shell script
cd ~/go/src/github.com/pantheon-systems
git clone git@github.com:pantheon-systems/search-secrets.git
cd search-secrets
go build .
```

Run from within the source directory (for now):

```shell script
search-secrets search \
  --github-token=<GITHUB_TOKEN> \
  --organization=pantheon-systems \
  --repos=titan-mt
```

You will find the output at `./output`.

## TODO

- Make the install and usage less complex.
