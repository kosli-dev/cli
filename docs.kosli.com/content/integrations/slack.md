---
title: Slack integration
bookCollapseSection: false
weight: 320
---
[Kosli Slack App](#kosli-slack-app) allows you to configure and receive notifications about changes in your environments
and query Kosli about your environments and artifacts without leaving Slack window.

## Installation

Visit https://slack.kosli.com to add Kosli Slack App to your Slack workspace.

After you install the app you need to invite it to channel you want to be able to use `/kosli` from.  
You can do it by typing below command in the channel:
```
/invite @kosli
```
## Usage

Now that Kosli Slack App is installed and invited to selected channel you can start using all `/kosli` commands in that channel.

The next step is connecting your Slack user with your Kosli user, use the command below to do that:
```
/kosli login
```

After that you may want to set up default Kosli organization, so you don't have to provide it every time you want to run `/kosli` commands from slack.  
E.g. if the organization name is **my-org**: 
```
/kosli config org my-org
```

When all of the above is done you can run `/kosli help` to see which commands are available.

In case of commands referring to snapshots you can specify snapshot(s) you're interested in multiple ways:
- environmentName~N *N'th behind the latest snapshot*
- environmentName#N *snapshot number N*
- environmentName@{YYYY-MM-DDTHH:MM:SS} *snapshot at specific moment in time in UTC*
- environmentName *the latest snapshot*

### Example

Here is an example of *search* command and the response:  

`/kosli search edb1a262`
{{<figure src="/images/slack-kosli-search.png" alt="Kosli search slack message" width="700">}}



