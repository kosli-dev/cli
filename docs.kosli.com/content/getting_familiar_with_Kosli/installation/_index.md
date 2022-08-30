---
title: Installing Kosli CLI
bookCollapseSection: false
weight: 1
---


## Installing Kosli CLI

If you have [Homebrew](https://brew.sh/) (available on MacOS, Linux or Windows Subsystem for Linux), 
you can install the Kosli CLI by running: 

```shell
$ brew install kosli-dev/tap/kosli
```

Alternatively, the Kosli CLI can be downloaded from: https://github.com/kosli-dev/cli/releases
Put it in a location you'll be running it from (as `./kosli`) or add it to your PATH so you can use it anywhere (as `kosli`)


## Using environment variables

All the kosli commands contain some common
flags `--api-token` and `--owner`. By setting
these as environment variables we don't need to specify them. 

You do this by capitalizing the flag in snake case and adding the `KOSLI_` prefix. 
For example, to set `--api-token xx` from an environment variable, you can `export KOSLI_API_TOKEN=xx`, etc:

```shell
export KOSLI_API_TOKEN=<put your kosli API token here>
export KOSLI_OWNER=<put your github username here>
```

## Getting your Kosli API token

To get the kosli API token go to https://app.kosli.com, log in using your github account, and go to your Profile (you'll find it by clicking on your avatar in the top right corner of the page).



