---
title: "Application Developers"
bookCollapseSection: false
weight: 300
summary: "How Application Developers can use Kosli to build secure, compliant software delivery pipelines."
---

# Application Developers

You build and maintain services, applications, or APIs that deliver customer or business value. You write code, push changes, and expect your work to move safely from commit to production. You care about quality, security, and release velocity, but you do not want to be slowed down by compliance overhead.


# How Kosli helps you

Kosli captures the evidence that your changes have passed the right checks, like tests, code reviews, security scans, and approvals, so you can deploy with confidence and stay focused on building.

With Kosli, you can:
- Ship code without worrying about compliance gates or approval tickets
- Get clarity on why something cannot deploy, and what needs to happen
- Use existing CI workflows without learning new tools
- Trace what changed, where it went, and whether it passed all required controls

## Your role in using Kosli

As an application developer, you are a contributor to the system of record that Kosli observes. You may:
- Write code that passes through a Flow defined by your platform team
- Produce build artifacts and test results that Kosli records as evidence
- Trigger attestations through CI jobs (e.g., when tests run or scans complete)
- Occasionally check compliance status in the UI or via pull request checks

You are usually not responsible for setting up Kosli. It runs quietly underneath your normal delivery workflows.


## What you’ll Work with

You typically interact with Kosli through:
- **Your CI/CD pipeline**, which calls Kosli CLI under the hood
- **Pull requests or merge gates**, where Kosli may block or allow merges based on compliance
- **The Kosli UI**, to check deployment or compliance status if needed
- **Your platform team's guidance**, for understanding what evidence is expected

You do not need to memorize Kosli commands or manage configurations. Most of it is abstracted away by your Platform team

## What success looks like

- You write and commit code as usual, and your changes flow smoothly through CI and into production
- You do not need to fill out compliance tickets or wait for manual approvals
- If something is blocked, Kosli tells you what evidence is missing and how to resolve it
- You gain confidence that your work is secure and production-ready without extra effort

## Common questions you might have

**"Why did Kosli mark my build or deployment non-compliant?"**<br>
Most likely a required check did not run, failed, or was not reported to Kosli. Your pipeline or platform team can help you identify what is missing.

**"Do I need to learn another CLI or tool?"**<br>
No. Kosli is used behind the scenes by your platform team. You may see its results in PR checks or dashboards, but you do not need to run it manually.

**"How do I know if my change was successfully deployed?"**<br>
You can use the Kosli UI to trace a git commit, artifact, or deployment. Kosli shows where it is running and what evidence was attached.

**"Can I use Kosli in debugging or incident response?"**<br>
Yes. Kosli helps you trace what changed and when across environments. You can see exactly what was deployed and what passed or failed.

## Where to start
- [**Getting Started**]({{< ref "/getting_started" >}}): Follow this if you're curious about how Kosli works behind the scenes
- [**Querying Kosli**]({{< ref "/tutorials/querying_kosli/" >}}): Learn how to search for artifacts or changes
- [**Concepts**]({{< ref "/understand_kosli/concepts" >}}): Understand what Kosli tracks and why