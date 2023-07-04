#!/bin/bash

# Define the Jira base URL
jiraBaseURL="https://kosli.atlassian.net/browse/"

# Get the current commit hash
commitHash=$(git rev-parse HEAD)

# Get the Jira ticket key from the current branch name
jiraKey=$(git symbolic-ref --short HEAD | grep -oE '[A-Z]+-[0-9]+')

if [ -z "$jiraKey" ]; then
    echo "No Jira ticket key found."
    exit 1
fi

# Construct the Jira ticket URL
jiraURL="${jiraBaseURL}${jiraKey}"
echo "Jira ticket URL: $jiraURL"

./kosli report evidence commit generic \
        --build-url=http://www.example.com \
        --commit=ff475e6958f1b8d529118a0b8410428ecc2060a5 \
        --name=jira-ticket-generic \
        --compliant=TRUE \
        --evidence-url=$jiraURL