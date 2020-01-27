# search-secrets

Find committed secrets in Pantheon's public and private GitHub repos.

## Prerequisites

Install [truffleHog](https://github.com/dxa4481/truffleHog) and make it available in your path.

## Usage

```shell script
export GITHUB_ACCESS_TOKEN="[your key here]"
./pull.sh
```

The output should look like this:

```
== Pulling repo data from GitHub
OK
== Processing apm
Searching apm ...
No secrets found
== Processing certificate-authentication-for-jenkins
Searching certificate-authentication-for-jenkins ...
No secrets found
== Processing createjobadvanced-plugin
Searching createjobadvanced-plugin ...
No secrets found
== Processing dashboard-view-plugin
Searching dashboard-view-plugin ...
No secrets found
== Processing dnscheck
...
```


## TODO

- Ignore false positives so we can return a useful 0 or 1 code
- Convert this to a Kubernetes CronJob that fires periodically and notifies #security-operations (INFOSEC-1165)
