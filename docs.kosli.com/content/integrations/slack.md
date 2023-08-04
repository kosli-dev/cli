---
title: Slack integration
bookCollapseSection: false
weight: 320
---
# Slack integration

There are two ways of using Kosli with Slack:
- you can use [Slack webhooks](#slack-webhooks) to receive notifications about changes in your environments
- you can install [Kosli Slack App](#kosli-slack-app) that on top of notifications allows you to query Kosli without leaving Slack window.

## Kosli Slack App

Visit https://slack.kosli.com to add Kosli Slack App to your Slack workspace.

### Usage

After you install the app you need to invite it to channel you want to be able to use `/kosli` from.  

The next step is running `/kosli login` so you can connect your Slack user with your Kosli user.  

After that you want to set up default Kosli organization, so you don't have to provide it every time you want to run `/kosli` commands from slack.  
E.g. if the organization name is **my-org** you'd run `/kosli config org my-org`

When all of the above is done you can run `/kosli help` to see which commands are available.

In case of commands referring to snapshots you can specify snapshot(s) you're interested in multiple ways:
- environmentName~N *N'th behind the latest snapshot*
- environmentName#N *snapshot number N*
- environmentName@{YYYY-MM-DDTHH:MM:SS} snapshot at specific moment in time in UTC
- environmentName the latest snapshot

### Example

Here is an example of *search* command and the response:  

`/kosli search edb1a262`
{{<figure src="/images/slack-kosli-search.png" alt="Kosli search slack message" width="700">}}

## Slack webhooks

### Slack webhooks setup setup

In order to receive the notifications you need to create a slack app: https://api.slack.com/authentication/basics#creating

When your app is created add **Incoming Webhooks** feature. Once you activate it you can **Add new Webhook** to chosen slack workspace, where you can select the channel to which the notifications will be sent. 

When the webhook is ready you can copy **Webhook URL**, go to [app.kosli.com](https://app.kosli.com), click on **Notifications** tab in **Settings** page for your organization. Then click **Create new notification** button.  
In pop-up menu, paste the webhook URL in **Slack webhook** field and fill the rest of the form.
The **Triggers** field is a list of events that will trigger the notification. And the **Environments** field is a list of environments for which the triggers will be checked.

{{<figure src="/images/slack.png" alt="Slack webhook setting" width="900">}}

Remember to click **Create** button. You can also check if it works by sending test notification, using the button under the **Slack Webhook** field.

### Example 

When one of your environments becomes non-compliant you'll get a notification in a configured channel:

{{<figure src="/images/slack-noncompliant-env.png" alt="Slack non-compliant notification" width="700">}}

You can also get a notification when the status changes from non-compliant to compliant:

{{<figure src="/images/slack-compliant-env.png" alt="Slack compliant notification" width="700">}}


