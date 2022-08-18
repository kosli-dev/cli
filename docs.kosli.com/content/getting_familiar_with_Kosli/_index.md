---
title: Getting Familiar with Kosli
bookCollapseSection: true
weight: 1
---

# Getting Familiar with Kosli

Kosli stores information about the SW you build in CI pipelines
and run in your runtime environment. The Kosli CLI is used for reporting
and querying the information.

Typically all the reporting will be done as part of your CI and runtime systems.
In the getting started you don't need any of this. Local code, git and a terminal are enough.

The Kosli CLI is tool agnostic and can run on any major platform (Linux, Mac, Windows).
Kosli does not require you to change your existing process.

The purpose of this guide is to familiarize you with the Kosli CLI and concepts.

When you are done with the guide you should be able to start adding Kosli to
your CI system and runtime environment.


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

To get the kosli API token go to https://app.kosli.com, log in using your github account, and go to your Profile (you'll find it if you click on your avatar in the top right corner of the page)

## Using a web browser

As you go through the guide you can also check your progress from 
[your browser](https://app.kosli.com).

In the upper left corner there is a house icon. Next to it you can select
which organization you want to view. Your personal organization
has the same name as your github login name, and is the organization you will
be using in this guide.

