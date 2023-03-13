---
title: FAQ
bookCollapseSection: false
weight: 700
---

# Frequently asked questions

If you can't find the answer you're looking for please:

* email us [here](mailto:info@kosli.com)
* join our slack community [here](https://join.slack.com/t/koslicommunity/shared_invite/zt-1dlchm3s7-DEP6TKjP3Mr58OZVB3hCBw)

## Errors

### Why am I getting "unknown flag" error?

If you see an error like below (or similar, with a different flag):
```
Error: unknown flag: --artifact-type
```
It most likely mean you misspelled a flag or a command.

E.g.
```
kosli expect deploymenct abc.exe --artifact-type file
Error: unknown flag: --artifact-type
```

The flag is spelled correctly, but there is a typo in deploymen**c**t.
The same error will pop up if you're trying to use a command that is not present in the version of the kosli CLI you are using.

### zsh: no such user or named directory

When running commands with argument starting with `~` you can encounter following problem:

```shell {.command}
kosli list snapshots prod ~3..NOW
```
```plaintext {.light-console}
zsh: no such user or named directory: 3..NOW
```

To help zshell interpret the argument correctly, wrap in in quotation marks (single or double): 
```shell {.command}
kosli list snapshots prod '~3..NOW'
```
or
```shell {.command}
kosli list snapshots prod "~3..NOW"
```

## Usage

### Where can I find API documentation?

At this point our API is not publicly available. The reason for this is that we are introducing a lot of changes and we can't guarantee the API endpoint you use will stay the same.  
That will change in the future.
<!-- 
### Do you support uploading a spdx or sbom as evidence?

We are working on providing that functionality in a near future. -->

### Do I have to provide all the flags all the time? 

A number of flags won't change their values often (or at all) between commands, like `--owner` or `--api-token`.  Some will differ between e.g. workflows, like `--flow`. You can define them as environment variable to avoid unnecessary redundancy. Check [Environment variables](/kosli_overview/kosli_tools/#environment-variables) section to learn more.

### What is dry run and how to use it?

You can use dry run to disable writing to app.kosli.com - e.g. if you're just trying things out, or troubleshooting (dry run will print the payload the CLI would send in a non dry run mode). 

Here are two possible ways of enabling a dry run:
1. use the `--dry-run` flag (no value needed) to enable it per command
1. set the `KOSLI_API_TOKEN` environment variable to `DRY_RUN` to enable it globally (e.g. in your terminal or CI)

### What is the `--config-file` flag?

A config file is an alternative for using Kosli flags or Environment variables. Usually you'd use a config file for the values that rarely change - like api token or owner, but you can represent all Kosli flags with config file. The key for each value is the same as the flag name, capitalized, so `--api-token` would become `API-TOKEN`, and `--owner` would become `OWNER`, etc. 

You can use JSON, YAML or TOML format for your config file. 

If you want to keep certain Kosli configuration in a file use `--config-file` flag when running Kosli commands to let the CLI know where to look for the file. The path given to `--config-file` flag should be a path relative to the location you're running kosli from. The file needs a valid format and extension, e.g.:

**kosli-conf.json:**
```
{
  "OWNER": "my-org",
  "API-TOKEN": "123456abcdef"
}
```

**kosli-conf.yaml:**
```
OWNER: "my-org"
API-TOKEN: "123456abcdef"
```

**kosli-conf.toml:**
```
OWNER = "my-org"
API-TOKEN = "123456abcdef"
```

When calling Kosli command you can skip file extension. For example, to list environments with `owner` and `api-token` in the configuration file you would run:

```
$ kosli environment ls --config-file kosli-conf
```

`--config-file` defaults to `kosli`, so if you name your file `kosli.<yaml|toml|json>` and the file is in the same location as where you run Kosli commands from, you can skip the `--config-file` altogether.


### Reporting same artifact and evidence multiple times
If an artifact or evidence is reported multiple times there are a few corner cases. 
The issues are described here.

## Template
When an artifact is reported, the template for the flow is stored together with the artifact. 
If the template has changed between the times the same artifact is reported, it is the last 
template that is considered the template for that artifact.

## Evidence
If a given named evidence is reported multiple times it is the compliance status of the last 
reported version of the evidence that is considered the compliance state of that evidence.

If an artifact is reported multiple times with different git-commit, we can have the same named 
commit-evidence being attached to the artifact through multiple git-commits. It is the last
reported version of the named commit-evidence that is considered the compliance state of that evidence.

## Evidence outside the template
If an artifact has an evidence, either commit evidence or artifact evidence, that is not 
part of the template the artifact is non-compliant.
