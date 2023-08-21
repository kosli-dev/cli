---
title: Slack integration
bookCollapseSection: false
weight: 330
---
# Kosli Slack App
[Kosli Slack App](#kosli-slack-app) allows you to configure and receive notifications about changes in your environments
and query Kosli about your environments and artifacts without leaving Slack window.

## Installation

Visit https://slack.kosli.com to add Kosli Slack App to your Slack workspace.
## Usage

Now that Kosli Slack App is installed you can start using all `/kosli` commands in any channel.

At any time you can run `/kosli help` to see which commands are available.

The next step is connecting your Slack user with your Kosli user, use the command below to do that:
```
/kosli login
```

After that you may want to set up default Kosli organization, so you don't have to provide it every time you want to run `/kosli` commands from slack.  
E.g. if the organization name is **my-org**: 
```
/kosli config org my-org
```

In case of commands referring to snapshots you can specify snapshot(s) you're interested in multiple ways:
- environmentName~N *N'th behind the latest snapshot*
- environmentName#N *snapshot number N*
- environmentName@{YYYY-MM-DDTHH:MM:SS} *snapshot at specific moment in time in UTC*
- environmentName *the latest snapshot*

### Example

Here is an example of *search* command and the response:  

`/kosli search edb1a262`
{{<figure src="/images/slack-kosli-search.png" alt="Kosli search slack message" width="700">}}



