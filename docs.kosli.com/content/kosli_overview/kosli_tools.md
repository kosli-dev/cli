---
title: 'How to use Kosli?'
weight: 140
---
# How to use Kosli?

## CLI

In order to [record environments](/how_to/record), [artifacts and events](/how_to/connect) to Kosli you need to use [Kosli CLI](https://github.com/kosli-dev/cli). 
The same tool can be used to [search](/how_to/search) Kosli and find out all you need to know about your runtime environments and artifacts.

Our CLI in an open source tool written in go and it's available for a number of different platforms.

To learn more about to install Kosli CLI click [here](/getting_started/installation)

### Usage

<!-- TODO:

explain kosli version and kosli status commands -->

[Reference](/client_reference/) section contains all the information you may need to run Kosli CLI. 

Most of the commands require a number of flags. Some of them are **required**, some are **optional** - you don't need to use them if you don't want to and some are **conditional** - you need to use it if a certain conditions occurs, e.g.:
* if you use `--sha256` flag it means you provide artifact's fingerprint on your own and we don't need to calculate it, so the flag `--artifact-type` is not needed
* if you want to read docker digest from registry without pulling the image you need to provide registry information: `--registry-password`, `--registry-username` and `--registry-provider`

Each conditional flag is explained in its description.

Some of the flags are **defaulted**, and the default value will be always printed in the description. You can skip the flag if the default value is what you chose to use.

Depending on the CI tool you are using some of the flags (including required ones) may also be defaulted, depending on the environment variables provided by the tool. If the flag is defaulted in your CI you don't have to provide it in the command. [Here](/ci-defaults) you can find more details about flags defaulted depending on CI.

### Environment variables

Each flag can be provided directly or represented with environment variable. In order to represent a flag with environment variable you need to create a variable with a `KOSLI_` prefix, followed by the flag name capitalized and internal dashes replaced by underscores, e.g.:

* `--api-token` can be represented by `KOSLI_API_TOKEN` 
* `--owner` can be represented by `KOSLI_OWNER`

etc.

### Dry run

You can use dry run to disable reporting to app.kosli.com - e.g. if you're just trying things out, or troubleshooting (dry run will print the payload cli would send in a non dry run mode). 

Here are two possible ways of enabling a dry run:
1. use `--dry-run` flag (no value needed) to enable it per command
1. set `KOSLI_API_TOKEN` environment variable to `DRY_RUN` to enable it globally (e.g. in your terminal or CI)

## Web UI

[app.kosli.com](https://app.kosli.com) is an easy way to monitor the status of your environments and pipelines. All you need to log in is a GitHub account.

![app.kosli.com](/images/app.png)

On the left of the page you can see the menu where you can:

1. Switch to another organization using dropdown menu
2. Switch to the Environments or the Pipelines view
3. Enter organization settings page
4. Access this documentation page

In the top right corner of the page you will see your GitHub avatar where you can access your profile settings (containing your Kosli api key). You'll also find there a link to the page where you can create a shared organization and links to [cyber-dojo demo project](https://app.kosli.com/cyber-dojo/environments/) and to log out. 