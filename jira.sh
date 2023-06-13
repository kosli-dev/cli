#!/bin/bash
echo "Jira ticket URL"
# Define the Jira base URL
jiraBaseURL="https://your-jira-instance/browse/"

# Get the current commit hash
commitHash=$(git rev-parse HEAD)

# Get the Jira ticket key from the current branch or commit message
jiraKey=$(git name-rev --name-only HEAD | grep -oE 'ABC-[0-9]+')

if [ -z "$jiraKey" ]; then
    echo "No Jira ticket key found."
    exit 1
fi

# Construct the Jira ticket URL
jiraURL="${jiraBaseURL}${jiraKey}"
echo "Jira ticket URL: $jiraURL"