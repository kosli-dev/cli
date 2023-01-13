---
title: Slack integration
bookCollapseSection: false
weight: 320
---
# Slack integration

You can use slack webhook to receive notifications about changes in the status of your environments.  

When one (or more) of your environments become non-compliant you'll get a notification in configured channel:

{{<figure src="/images/slack-noncompliant-env.png" alt="Slack non-compliant notification" width="700">}}


You'll also get a notification if the status changes from non-compliant to compliant:

{{<figure src="/images/slack-compliant-env.png" alt="Slack compliant notification" width="700">}}


## Slack integration setup

In order to receive the notifications you need to create a slack app: https://api.slack.com/authentication/basics#creating

When your app is created add **Incoming Webhooks** feature. Once you activate it you can **Add new Webhook** to chosen slack workspace, where you can select the channel to which the notifications will be sent. 

When the webhook is ready you can copy **Webhook URL**, go to [app.kosli.com](https://app.kosli.com), and paste it to **Slack Webhook** field under **Notifications** tab in **Settings** page for your organization.

{{<figure src="/images/slack.png" alt="Slack webhook setting" width="900">}}

Remember to click **Save** button. You can also check if it works by sending test notification, using the button under Webhook URL field.
