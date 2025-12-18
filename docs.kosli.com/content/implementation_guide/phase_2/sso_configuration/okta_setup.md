---
title: "Okta Setup"
bookCollapseSection: false
weight: 300
summary: "Step-by-step guide for configuring Single Sign-On (SSO) with Okta for your Kosli organization."
---

# Okta Setup for SSO

For Single Sign-On (SSO) integration between Okta and Kosli, you can choose and follow the steps outlined in one of the two methods provided below:

- [Create a new App integration](#create-a-new-app-integration)
- [Update or Rotate the Client Secret](#update-or-rotate-the-client-secret)

## Prerequisites

To begin the setup process, ensure that you:

- Are logged into the Okta Admin Console at https://admin.okta.com/
- Possess the necessary permissions to create a new application within Okta.

## Create a new App integration

### 1. Create the App Integration
Follow the official Okta documentation to create a new OIDC app integration, with the following settings:

- **Application type:** Web Application
- **Sign-in redirect URIs:** https://api.userfront.com/v0/auth/okta/login

### 2. Create a Client Secret

Follow the official Okta documentation to create a Client Secret for your newly created app integration.

## 3. Share details with Kosli Securely
Please share details below securely in order for Kosli to complete SSO setup.<br>

```
Okta client ID:                 abcdefghijklmnopqrst
Okta domain:                    mycompany.okta.com
Client Secret:                  xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
Client Secret Expiration Date:  1999-12-31 (format: yyyy-mm-dd)
```
See [Sharing Secrets Securely with Kosli]({{< relref "sharing_secrets_securely" >}}).

## Update or Rotate the Client Secret

To prevent downtime, we advise rotating your secrets safely and well in advance of their expiration date. This allows us to manage the update process smoothly.

### 1. Create a New Client Secret

Follow the official Okta documentation to create a new Client Secret for your existing app integration.

### 2. Share new Client Secret with Kosli Securely
Please share the new Client Secret securely with Kosli.

```
Client Secret:                  xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
Client Secret Expiration Date:  1999-12-31 (format: yyyy-mm-dd)
```

See [Sharing Secrets Securely with Kosli]({{< relref "sharing_secrets_securely" >}}).

## Troubleshooting

Once Kosli have confirmed the SSO setup, once you log in to Kosli, you should be redirected to the Okta login page.

### Common Issues

#### Problem: Unable to log in via SSO

Check the following common issues:

- **Wrong Redirect URI**
  - Ensure that the Redirect URI in your Okta app integration matches `https://api.userfront.com/v0/auth/okta/login`.
- **Invalid Client ID or Client Secret**
  - Verify that the values provided to Kosli are correct and correspond to those in your Okta app integration.
- **Expired Client Secret**
  - Ensure that the Client Secret provided to Kosli is still valid and has not expired.
  - If it has expired, follow the [Update or Rotate the Client Secret]({{< ref "#update-or-rotate-the-client-secret" >}}) steps to create a new client.

## References

### Okta Documentation

- [Create OpenID Connect app integrations](https://help.okta.com/en-us/content/topics/apps/apps_app_integration_wizard_oidc.htm)
- [Manage secrets and keys for OIDC app client authentication](https://help.okta.com/oie/en-us/content/topics/apps/oauth-client-cred-mgmt.htm)