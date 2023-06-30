---
title: Slack integration
bookCollapseSection: false
weight: 320
---
# Slack integration

You can use slack webhook to receive notifications about changes in your environments.  

For example, when one of your environments becomes non-compliant you'll get a notification in a configured channel:

{{<figure src="/images/slack-noncompliant-env.png" alt="Slack non-compliant notification" width="700">}}

You can also get a notification when the status changes from non-compliant to compliant:

{{<figure src="/images/slack-compliant-env.png" alt="Slack compliant notification" width="700">}}


## Slack integration setup

In order to receive the notifications you need to create a slack app: https://api.slack.com/authentication/basics#creating

When your app is created add **Incoming Webhooks** feature. Once you activate it you can **Add new Webhook** to chosen slack workspace, where you can select the channel to which the notifications will be sent. 

When the webhook is ready you can copy **Webhook URL**, go to [app.kosli.com](https://app.kosli.com), click on **Notifications** tab in **Settings** page for your organization. Then click **Create new notification** button.  
In pop-up menu, paste the webhook URL in **Slack webhook** field and fill the rest of the form.
The **Triggers** field is a list of events that will trigger the notification. And the **Environments** field is a list of environments for which the triggers will be checked.

{{<figure src="/images/slack.png" alt="Slack webhook setting" width="900">}}

Remember to click **Create** button. You can also check if it works by sending test notification, using the button under the **Slack Webhook** field.
