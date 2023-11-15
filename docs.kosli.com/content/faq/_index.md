---
title: FAQ
bookCollapseSection: false
weight: 700
---

# Frequently asked questions

If you can't find the answer you're looking for please:

* email us at [support@kosli.com](mailto:support@kosli.com)
* join our slack community [here](https://join.slack.com/t/koslicommunity/shared_invite/zt-1dlchm3s7-DEP6TKjP3Mr58OZVB3hCBw)

## Why am I getting "unknown flag" error?

If you see an error like below (or similar, with a different flag):
```
Error: unknown flag: --artifact-type
```
It most likely mean you misspelled a flag.

## "unknown command" errors
E.g.
```
kosli expect deploymenct abc.exe --artifact-type file
Error: unknown command: deploymenct
available subcommands are: deployment
```

Note that there is a typo in deploymen**c**t.
This error will pop up if you're trying to use a command that is not present in the version of the kosli CLI you are using.

## zsh: no such user or named directory

When running commands with argument starting with `~` you can encounter following problem:

```shell {.command}
kosli list snapshots prod ~3..NOW
```
```plaintext {.light-console}
zsh: no such user or named directory: 3..NOW
```

To help zshell interpret the argument correctly, wrap it in quotation marks (single or double): 
```shell {.command}
kosli list snapshots prod '~3..NOW'
```
or
```shell {.command}
kosli list snapshots prod "~3..NOW"
```

## Github can't see KOSLI_API_TOKEN secret

Secrets in Github actions are not automatically exported as environment variables. You need to add required secrets to your GITHUB environment explicitly. E.g. to make kosli_api_token secret available for all cli commands as an environment variable use following:

```yaml
env:
  KOSLI_API_TOKEN: ${{ secrets.kosli_api_token }}
```

## Where can I find API documentation?

Kosli API documentation is available for logged in Kosli users here: https://app.kosli.com/api/v2/doc/  
You can find the link at [app.kosli.com](https://app.kosli.com) after clicking at your avatar (top-right corner of the page)

<!-- 
### Do you support uploading a spdx or sbom as evidence?

We are working on providing that functionality in a near future. -->

## Do I have to provide all the flags all the time? 

A number of flags won't change their values often (or at all) between commands, like `--org` or `--api-token`.  Some will differ between e.g. workflows, like `--flow`. You can define them as environment variable to avoid unnecessary redundancy. Check [Environment variables](/kosli_overview/kosli_tools/#environment-variables) section to learn more.

## What is dry run and how to use it?

You can use dry run to disable writing to app.kosli.com - e.g. if you're just trying things out, or troubleshooting (dry run will print the payload the CLI would send in a non dry run mode). 

Here are two possible ways of enabling a dry run:
1. use the `--dry-run` flag (no value needed) to enable it per command
1. set the `KOSLI_DRY_RUN` environment variable to `true` to enable it globally (e.g. in your terminal or CI)
1. set the `KOSLI_API_TOKEN` environment variable to `DRY_RUN` to enable it globally (e.g. in your terminal or CI)

## What is the `--config-file` flag?

A config file is an alternative for using Kosli flags or Environment variables. Usually you'd use a config file for the values that rarely change - like api token or org, but you can represent all Kosli flags with config file. The key for each value is the same as the flag name, capitalized, so `--api-token` would become `API-TOKEN`, and `--org` would become `ORG`, etc. 

You can use JSON, YAML or TOML format for your config file. 

If you want to keep certain Kosli configuration in a file use `--config-file` flag when running Kosli commands to let the CLI know where to look for the file. The path given to `--config-file` flag should be a path relative to the location you're running kosli from. The file needs a valid format and extension, e.g.:

**kosli-conf.json:**
```
{
  "ORG": "my-org",
  "API-TOKEN": "123456abcdef"
}
```

**kosli-conf.yaml:**
```
ORG: "my-org"
API-TOKEN: "123456abcdef"
```

**kosli-conf.toml:**
```
ORG = "my-org"
API-TOKEN = "123456abcdef"
```

When calling Kosli command you can skip file extension. For example, to list environments with `org` and `api-token` in the configuration file you would run:

```
$ kosli environment ls --config-file kosli-conf
```

`--config-file` defaults to `kosli`, so if you name your file `kosli.<yaml|toml|json>` and the file is in the same location as where you run Kosli commands from, you can skip the `--config-file` altogether.


## Reporting the same artifact and evidence multiple times
If an artifact or evidence is reported multiple times there are a few corner cases. 
The issues are described here:

### Template
When an artifact is reported, the template for the flow is stored together with the artifact. 
If the template has changed between the times the same artifact is reported, it is the last 
template that is considered the template for that artifact.

### Evidence
If a given named evidence is reported multiple times it is the compliance status of the last 
reported version of the evidence that is considered the compliance state of that evidence.

If an artifact is reported multiple times with different git-commit, we can have the same named 
commit-evidence being attached to the artifact through multiple git-commits. It is the last
reported version of the named commit-evidence that is considered the compliance state of that evidence.

### Evidence outside the template
If an artifact have evidence, either commit evidence or artifact evidence, that is not 
part of the template, the state of the extra evidence will affect the overall compliance of the artifact.

## How to set compliant status of generic evidence

The `--compliant` flag is a [boolean flag](#boolean-flags). 
To report generic evidence as non-compliant use `--compliant=false`, as in this example:
```
$ kosli report evidence artifact generic server:1.0 \
  --artifact-type docker \
  --name test \
  --description "generic test evidence" \
  --compliant=false \
  --flow server
```

Keep on mind a number of flags, usually represented with environment variables, are omitted in this example.  
`--compliance` flag is set to `true` by default, so if you want to report generic evidence as compliant, simply skip providing the flag altogether.

## Boolean flags

Flags with values can usually be specified with an `=` or with a **space** as a separator.
For example, `--artifact-type=file` or `--artifact-type file`.
However, an explicitly specified boolean flag value **must** use an `=`.
For example, if you try this:
```
kosli report evidence artifact generic Dockerfile --artifact-type file  --compliant true ...
```
You will get an error stating:
```
Error: only one argument ... is allowed.
The 2 supplied arguments are: [Dockerfile, true]
```
Here, `--artifact-type file` is parsed as if it was `--artifact-type=file`, leaving:
```
kosli report evidence artifact generic Dockerfile --compliant true ...
```
Then `--compliant` is parsed as if *implicitly* defaulting to `--compliant=true`, leaving:
```
kosli report evidence artifact generic Dockerfile true ...
```
The parser then sees `Dockerfile` and `true` as the two
arguments to `kosli report evidence artifact generic`.
