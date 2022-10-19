---
title: 'CLI'
weight: 40
---

## CLI

In order to [record environments](/how_to/record), [artifacts and events](/how_to/connect) to Kosli you need to use [Kosli CLI](https://github.com/kosli-dev/cli). 
The same tool can be used to [query](/how_to/query) Kosli and find out all you need to know about your runtime environments and artifacts.

Our CLI in an open source tool written in go and it's available for a number of different platforms.

To learn more about to install Kosli CLI click [here](/getting_started/installation)

## Usage

[Reference](/client_reference/) section contains all the information you may need to run Kosli CLI. 

Most of the commands reguire a number of flags. Some of them are **required**, some are **optional** - you don't need to use them if you don't want to and some are **conditional** - you need to use it if a certain conditions occurs, e.g.:
* if you use `--sha256` flag it means you provide artifact's fingerprint on your own and we don't need to calculate it, so the flag `--artifact-type` is not needed
* if you want to read docker digest from registry without pulling the image you need to provide registry information: `--registry-password`, `--registry-username` and `--registry-provider`

Each conditional flag is explained in its description.

Some of the flags are **defaulted**, and the default value will be always printed in the description. You can skip the flag if the default value is what you chose to use.

Depending on the CI tool you are using some of the flags (including required ones) may also be defaulted, depending on the environment variables provided by the tool. If the flag is defaulted in your CI you don't have to provide it in the command. [Here](/ci-defaults) you can find more details about flags defaulted depending on CI.

## Environment variables

Each flag can be provided directly or represented with environment variable. In order to represent a flag with environment variable you need to create a variable with a `KOSLI_` prefix, followed by the flag name capitalized and internal dashes replaced by underscores, e.g.:

* `--api-token` can be represented by `KOSLI_API_TOKEN` 
* `--owner` can be represented by `KOSLI_OWNER`

etc.