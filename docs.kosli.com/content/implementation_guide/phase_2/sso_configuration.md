---
title: "SSO Configuration"
bookCollapseSection: false
weight: 200
summary: "Step-by-step guide for configuring Single Sign-On (SSO) with Microsoft Entra ID for your Kosli organization."
---

# Microsoft Entra ID Setup for SSO

For Single Sign-On (SSO) integration between Microsoft Entra ID and Kosli, you can choose and follow the steps outlined in one of the two methods provided below:

- [Create a new App Registration](#create-a-new-app-registration)
- [Update or Rotate the Client Secret](#update-or-rotate-the-client-secret)

## Prerequisites

To begin the setup process, ensure that you:

- Are logged into the Azure Portal at https://portal.azure.com/
- Possess the necessary permissions to create a new App registration within Microsoft Entra ID.

## Create a new App Registration

To configure Single Sign-On (SSO) with Kosli for the first time, proceed with the following setup steps:

### 1. Create the App Registration

1. In the left menu, go to **Microsoft Entra ID → App registrations**.
2. Select **New registration**.
3. Enter a meaningful **Name** for the app (for example: `kosli-sso`).
4. Under Supported account types, choose:
    - **Accounts in this organizational directory only ([name] - Single tenant).**
5. Under **Redirect URI**, select:
    - **Platform**: Web
    - **URI**: https://api.userfront.com/v0/auth/azure/login
6. Click **Register**.

### 2. Create a Client Secret

1. Inside the new app, open **Certificates & secrets**.
2. Under **Client secrets**, click **New client secret**.
3. Add a description for your client secret
4. Choose an expiration period (typically **12 months**).
5. Click **Add**.
6. Record the **Value** immediately.

{{% hint warning %}}
**Note:**
This secret value is never displayed again after you leave this page.
{{% /hint %}}

{{% hint info %}}
**Important:**
Make sure to assign the necessary user and group assignments to the application so the intended users can access Kosli via SSO.
{{% /hint %}}

### 3. Share details with Kosli Securely
Please share details below securely in order for Kosli to complete SSO setup.<br>

```
Application (client) ID:        aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
Directory (tenant) ID:          11111111-2222-3333-4444-555555555555
Client Secret:                  xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
Client Secret Expiration Date:  1999-12-31 (format: yyyy-mm-dd)
```
See [Securely share secrets with Kosli](#securely-share-secrets-with-kosli).

## Update or Rotate the Client Secret

To prevent downtime, we advise rotating your secrets safely and well in advance of their expiration date. This allows us to manage the update process smoothly.

### 1. Create a New Client Secret

1. Go to **Microsoft Entra ID → App registrations**
2. Select tab **All applications**
3. Find the **Application (client) ID** that matches your Kosli app.
4. Open **Certificates & secrets**.
5. Under **Client secrets**, select **New client secret**.
6. Add a description (e.g., `Rotation <year>`), choose an expiration period, and click **Add**.
7. Record the **Value** immediately.

{{% hint warning %}}
**Note:**
This secret value is never displayed again after you leave this page.
{{% /hint %}}

### 2. Share new Client Secret with Kosli Securely
Please share the new Client Secret securely with Kosli.

```
Client Secret:                  xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
Client Secret Expiration Date:  1999-12-31 (format: yyyy-mm-dd)
```

See [Securely share secrets with Kosli]({{< ref "#securely-share-secrets-with-kosli" >}}).

## Securely share secrets with Kosli

For securely sharing your secrets with Kosli, we recommend using one of the following services:

* **Onetime Secret:** https://eu.onetimesecret.com
* **Yopass:** https://yopass.se

After encrypting the secret and generating the link, please email the link to support@kosli.com or your Kosli contact, so we can finalize the SSO registration process.

{{% hint warning %}}
**Important:**
The expiration for this must be set to a minimum of 7 days to allow Kosli to process it correctly.
{{% /hint %}}


## Troubleshooting

Once Kosli have confirmed the SSO setup, once you log in to Kosli, you should be redirected to the Microsoft Entra ID login page.

### Common Issues

#### Problem: Unable to log in via SSO

if you encounter the following error message when attempting to log in via SSO, depending on the Kosli region you are using:

{{< tabs "region" >}}
{{< tab"EU" >}}
https://app.kosli.com/?error_message=Failed+obtaining+azure+access+token
{{< /tab >}}
{{< tab "US" >}}

https://app.us.kosli.com/?error_message=Failed+obtaining+azure+access+token
{{< /tab >}}
{{< /tabs >}}

Check the following common issues:

- **Wrong Redirect URI**
  - Ensure that the Redirect URI in your Microsoft Entra ID app registration matches `https://api.userfront.com/v0/auth/azure/login`.
- **Invalid Application ID, Directory ID, or Client Secret**
  - Verify that the values provided to Kosli are correct and correspond to those in your Microsoft Entra ID app registration.
- **Expired Client Secret**
  - Ensure that the Client Secret provided to Kosli is still valid and has not expired
  - If it has expired, follow the [Update or Rotate the Client Secret]({{< ref "#update-or-rotate-the-client-secret" >}}) steps to create a new client
- **User and Group Assignments**
  - Ensure that the necessary user and group assignments have been made to the application in Microsoft Entra ID so that users can access Kosli via SSO.

## References

### Microsoft Documentation
- [Register an application in Microsoft Entra ID](https://learn.microsoft.com/en-us/azure/active-directory/develop/quickstart-register-app)
- [Add and manage application credentials in Microsoft Entra ID](https://learn.microsoft.com/en-us/entra/identity-platform/how-to-add-credentials?tabs=client-secret)
- [Manage users and groups assignment to an application](https://learn.microsoft.com/en-us/entra/identity/enterprise-apps/assign-user-or-group-access-portal?pivots=portal)