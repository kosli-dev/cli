---
title: Actions
bookCollapseSection: false
weight: 320
---

# Kosli Actions

Actions enable you to automate the execution of if-this-do-that workflows based on Kosli events. You can configure actions to either receive a Slack notification or a JSON payload on a custom webhook when certain Kosli events happen.   

You can configure actions to be triggered by one or more of the following events occurring in one or more environments:

- When a new artifact starts execution in an environment.
- When an artifact ceases execution in an environment.
- When instances of an artifact are scaled up or down.
- When an artifact is added to the allow-list in an environment.
- When an environment transitions from a **Compliant** state to a **Non-Compliant** state.
- When an environment changes from a **Non-Compliant** state to a **Compliant** state.


## Slack Notifications

To receive Kosli notifications in Slack, you have two options:

1) Using Kosli Slack App (recommended)

Subscribe to Kosli notifications using the [Kosli Slack App](/integrations/slack/). This method is recommended for a seamless integration.
Use the app to create notification settings by running the `/kosli subscribe` slash command.

2) Using Slack Incoming Webhooks

- Create a [Slack incoming webhook](https://api.slack.com/messaging/webhooks#create_a_webhook).
- Utilize this webhook to [create a notification settings in the Kosli UI](/integrations/actions/#manage-actions-in-the-ui).
  
Both approaches allow you to configure Kosli notifications in Slack, offering flexibility based on your preferences.

## Custom Webhook Notifications

Custom webhook notifications empower you to implement automation workflows for "if-this-then-that" scenarios. Whenever an event that matches your specified notification settings occurs, a JSON payload, as outlined below, is transmitted to your designated custom webhook:

```json
{
    "version": "1.0",
    "timestamp": "1692616493",
    "org": "cyber-dojo",
    "environment": "aws-prod",
    "event_type": "ARTIFACT_STARTED",
    "description": "1 instance started running (from 0 to 1)",
    "snapshot":  {
           "index": "1035",
           "status": "compliant",
           "html_url": "https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/1035",
           "api_url": "https://app.kosli.com/api/v2/snapshots/cyber-dojo/aws-prod/1035"
     },
    "artifact": {
        "name": "runner",
        "fingerprint": "719defb995c86ad7c406ad74258fe98b9ebd71dfa80cd786870c967cb6c1f08d",
        "provenance": {
            "flow": "runner",
            "status": "compliant",
            "commit": "1ac157003dd6fb9ec764daa47726b7bfed65c312",
            "commit_url": "https://github.com/cyber-dojo/runner/commit/1ac157003dd6fb9ec764daa47726b7bfed65c312",
            "html_url": "https://app.kosli.com/cyber-dojo/runner/719defb995c86ad7c406ad74258fe98b9ebd71dfa80cd786870c967cb6c1f08d",
            "api_url": "https://app.kosli.com/api/v2/artifacts/cyber-dojo/runner/fingerprint/719defb995c86ad7c406ad74258fe98b9ebd71dfa80cd786870c967cb6c1f08d",
            "build_url": "https://github.com/cyber-dojo/runner/actions/runs/5891969166",
            "deployments": [ 
                {
                   "number": "44",
                   "timestamp": "1692618644",
                   "build_url": "https://github.com/cyber-dojo/runner/actions/runs/5891969166",
                   "html_url": "https://app.kosli.com/cyber-dojo/flows/runner/deployments/44",
                   "api_url": "https://app.kosli.com/api/v2/deployments/cyber-dojo/runner/44"
               }
            ],
            "approvals": [
                {
                   "number": "42",
                   "timestamp": "1692617329",
                   "state": "approved",
                   "latest_reviewer": "username",
                   "latest_review_comment": "lgtm",
                    "html_url": "https://app.kosli.com/cyber-dojo/flows/runner/approvals/42",
                    "api_url": "https://app.kosli.com/api/v2/approvals/cyber-dojo/runner/42"
               }
            ]
        }
    }
}
```

# Manage Actions in the UI

You can manage Actions for your organization in the Kosli UI from the `Actions` section in the left navigation menu. The Actions sections enables you to:
- Create Notifications: Create a new notifications settings.
- Delete Notifications: Remove existing notification settings that are no longer needed.
- Update Notifications: Modify notification settings as needed.

