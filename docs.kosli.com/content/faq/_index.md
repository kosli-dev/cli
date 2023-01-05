---
title: FAQ
bookCollapseSection: False
weight: 60
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
Same error will pop up if you're trying to use a command that is not present in the version of kosli cli you are using.

## Usage

### Do I have to provide all the flags all the time? 

A number of flags won't change their values often (or at all) between commands, like `--owner` or `api-token`.  Some will differ between e.g. workflows, like `--pipeline`. You can define them as environment variable to avoid unnecessary redundancy. Check [Environment variables](/introducing_kosli/cli/#environment-variables) section to learn more.

### What is dry run and how to use it?

You can use dry run to disable writing to app.kosli.com - e.g. if you're just trying things out, or troubleshooting (dry run will print the payload cli would send in a non dry run mode). 

There is a few ways you can enable dry run mode
1. use `--dry-run` flag (no value needed) to enable it per command
1. set `KOSLI_API_TOKEN` environment variable to `DRY_RUN` to enable it globally (e.g. in your terminal or CI)

### What is the `--config-file` flag?

A config file is an alternative for using Kosli flags or Environment variables. Usually you'd use config file for the values that rarely change - like api token or owner, but you can represent all Kosli flags with config file. The key for each value is the same as the flag name, capitalized, so `--api-token` would become `API-TOKEN`, and `--owner` would become `OWNER`, etc. 

You can use JSON, YAML or TOML format for your config file. 

If you want to keep certain Kosli configuration in a file use `--config-file` flag when running Kosli commands to let the cli tool know where to look for the file. The path given to `--config-file` flag should be a path relative to the location you're running kosli from. The file needs a valid format and extension, e.g.:

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