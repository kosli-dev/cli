---
title: "Step 3: Configure your working environment"
bookCollapseSection: false
weight: 250
---

# Step 3: Configure your working environment

## Getting your Kosli API token

<!-- Put this in a separate page? -->
<!-- Add screen shot here? -->

To be able to run Kosli commands (from your local machine, but the same goes for any CI/CD system you use) you need to use Kosli API Token to be able to authenticate. It's a common practice to configure the token as an environment variable (or e.g. a secret in GitHub Actions or Bitbucket, etc)

To retrieve your API Token:

* Go to https://app.kosli.com
* Log in or sign up using your github account
* Open your Profile page (click on your avatar in the top right corner of the page) and copy the API Key

## Using environment variables

<!-- Put this in a separate page? -->

The `--api-token` and `--owner` flags are used in every `kosli` CLI command.  
Rather than retyping these every time you run `kosli`, you can set them as environment variables.

The owner is the name of the organization you intend to use - it is either your private organization, which has exactly the same name as your GitHub username, or a shared organization (if you created or have been invited to one).

By setting the environment variables:
```shell {.command}
export KOSLI_API_TOKEN=abcdefg
export KOSLI_OWNER=cyber-dojo
```

you can use

```shell {.command}
kosli pipeline ls 
```

instead of

```shell {.command}
kosli pipeline ls --api-token abcdefg --owner cyber-dojo 
```

You can represent **ANY** flag as an environment variable. To do that you need to capitalize the words in the flag, replacing dashes with underscores, and add the `KOSLI_` prefix. For example, `--api-token` becomes `KOSLI_API_TOKEN`.