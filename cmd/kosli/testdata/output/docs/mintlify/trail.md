---
title: "trail"
description: "Get the metadata of a specific trail."
---

## Synopsis

```shell
trail TRAIL-NAME [flags]
```

Get the metadata of a specific trail.

## Flags
| Flag | Type | Description |
| :--- | :--- | :--- |
| `-f`, `--flow` | string | The Kosli flow name. |
| `-h`, `--help` | bool | help for trail |
| `-o`, `--output` | string | [defaulted] The format of the output. Valid formats are: [table, json, markdown]. (default "table") |


## JSON output

With `--output json`, this command prints the response of the [Get trail](/api-reference/trails/get-trail) API operation as the server returned it. The shape follows the server's published API spec, not the CLI version.

