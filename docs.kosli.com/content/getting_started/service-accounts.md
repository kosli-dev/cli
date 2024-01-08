---
title: "Part 3: Service Accounts"
bookCollapseSection: false
weight: 230
---
# Part 3: Create a Service Account

Prior to engaging with Kosli, authentication is necessary. There are two methods to achieve this:

1. Utilizing a service account API key (recommended).
2. Using a personal API key.

## Service Accounts

{{< hint warning >}}
Service accounts are exclusively available within shared organizations.
{{< /hint >}}

A service account represents a machine user designed for interactions with Kosli from external systems, such as CI or runtime environments.

To create a service account:

- Log in to Kosli.
- From the left navigation menu, choose the organization where you wish to create the service account.
- Navigate to `Settings` in the left navigation menu.
- Select `Service accounts` from the settings sub-menu.
- Click `Add new service account`, provide a name for the service account, and click Add.
- Once created, generate an API key for the service account by clicking `Add API Key`.
- Choose a Time-To-Live (TTL) for the key, add a descriptive label, and then click `Add`.
- Ensure to copy the generated key as it won't be retrievable later. This key serves as the authentication token.


## Personal API Keys

{{< hint warning >}}
Personal API keys possess equivalent permissions to your user account, encompassing access to multiple organizations. Therefore, exercise caution while utilizing personal API keys. These keys grant access and perform actions as per the associated user's permissions across various organizations.
{{< /hint >}}

To create a personal API key:
- Login to Kosli 
- From your user menu on the top right corner, click `Profile`
- In the API Keys section, click `Add API Key`, select a Time-To-Live (TTL) for the key, add a descriptive label, and then click `Add`
- Ensure to copy the generated key as it won't be retrievable later. This key serves as the authentication token.


### API Keys rotation

You can execute a zero-downtime API key rotation by following these steps:

- **Generate a New Key**: 
Create a new API key that will replace the existing key.

- **Replace the Old Key Where Used**: 
Implement the new key in all areas where the old key is currently utilized for authentication or access.

- **Delete the Old Key:**
Once the new key is in place and operational, remove or delete the old key from the system or applications where it was previously employed for security or authentication purposes.

By systematically following these steps, you can ensure a seamless API key rotation without causing any downtime or interruptions in service.


### Using API Keys

#### In CLI

you can assign an API key to any CLI command by one of the following options:
- using the `--api-token` flag
- exporting an environment variable called `KOSLI_API_TOKEN`
- setting it in a config file and passing the config file using `--config-file` (see [here](/getting_started/install#assigning-flags-via-config-files))

#### In API

When utilizing the Kosli API directly, you can authenticate your requests using basic authentication. Set the `username` to your API key and the `password` to any string value. 