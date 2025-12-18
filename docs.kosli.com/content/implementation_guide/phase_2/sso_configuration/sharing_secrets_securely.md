---
title: "Sharing Secrets Securely"
bookCollapseSection: false
weight: 400
summary: "How to securely share secrets with Kosli during Single Sign-On (SSO) configuration."
---

# Sharing Secrets Securely

For securely sharing your secrets with Kosli, we recommend using one of the following services:

- **Onetime Secret:** https://eu.onetimesecret.com
- **Yopass:** https://yopass.se

If your organization uses a different secret management tool that allows you to generate an access link, you can use that as well.

After encrypting the secret and generating the link, please email the link to support@kosli.com or your Kosli contact, so we can finalize the SSO registration process.

{{% hint warning %}}
**Important:**
- Please ensure that the expiration for this must be set to a **minimum of 7 days** to allow Kosli to process it correctly.
- Please allow **multiple access attempts**, as Kosli may need to access the secret more than once during the setup process.
- Kosli will only access the secret for the purpose of completing the SSO setup and will not store or share it beyond this use case.
{{% /hint %}}
