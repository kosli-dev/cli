---
title: "Platform Engineers"
bookCollapseSection: false
weight: 200
summary: "How Platform Engineers and DevOps teams can use Kosli to build secure, compliant software delivery pipelines."
---

## Platform and DevOps Engineers

You build the internal tooling, workflows, and golden paths that help developers ship software reliably and securely. You care about scaling delivery without scaling your team.

If you’re supporting CI/CD pipelines, infrastructure, or compliance enablement across multiple services or teams, this page is for you.

## How Kosli helps you

Kosli gives you a single, unified way to track everything that moves through your delivery pipelines: code, artifacts, tests, approvals, deployments and prove it’s been done safely and correctly.

With Kosli, you can:
- Automate compliance and eliminate manual change approval processes.
- Capture tamper-proof evidence across your SDLC (without slowing down delivery).
- Monitor all runtime environments and deployments across teams.
- Offer developers secure paved paths that embed governance from the start.


## Your role in using Kosli

As a platform engineer, you're typically responsible for:
- Setting up Kosli in CI/CD and infrastructure environments.
- Creating and maintaining **Flows**, which model how changes move through pipelines.
- Defining and triggering **Trails** to capture each run of those pipelines.
- Configuring **Attestations** for tests, scans, and internal checks (e.g., Jira, Snyk).
- Capturing **Environment Snapshots** and enforcing **Policies** to govern deployments.
- Building reusable Kosli integrations (e.g., GitHub Actions, GitLab CI templates) so your dev teams don’t have to think about it.

You’ll often be the first person to integrate Kosli into your platform and roll it out to the rest of the org.

## What you’ll work with

You’ll primarily interact with:
- **Kosli CLI:** integrated into your CI/CD pipelines and scripts.
- **Flows** and **Trails:** to represent and track software delivery runs.
- **Artifacts** and **Attestations:** to connect builds and compliance evidence.
- **Environment Snapshots** & **Policies:** to enforce governance in prod and staging.
- **Kosli UI:** to review deployment status, compliance views, and audits.

If you're running Kubernetes, Terraform, or other infrastructure tools, Kosli also integrates easily to monitor state and changes.

## What success looks like

When Kosli is successfully adopted by platform engineering, you’ll see:

- Your pipelines continuously produce verifiable, compliant deployments.
- You eliminate the need for spreadsheet-driven approvals and CAB meetings.
- Developers onboard Kosli passively via the platform, they rarely have to learn it directly.
- Security and compliance teams get everything they need with minimal friction.
- Audits are a non-event: you already have the evidence.

## Common questions you might have

**“Do I need to change our pipelines to use Kosli?”**<br>
No major changes. Kosli integrates via CLI commands you can drop into any pipeline.

**“Can I templatize this across many teams?”**<br>
Yes. Use flow templates and reusable CI snippets to roll out a consistent setup.

**“Does Kosli work with our existing tools?”**<br>
Almost certainly. Kosli is tool-agnostic and supports GitHub Actions, GitLab, Jenkins, Kubernetes, Terraform, and more.

**“How do I know it’s working?”**<br>
Kosli automatically gives you compliance status per environment and per change. You can inspect Trails, download audit packages, and integrate with Slack or through Webhooks for alerts.

## Where to start

- [**Getting Started Guide**]({{< ref "/getting_started" >}}): For a complete technical setup walkthrough.
- [**CLI Reference**]({{< ref "/client_reference" >}}): Full list of commands.
- [**Concepts Overview**]({{< ref "/understand_kosli/concepts" >}}): Understand how Flows, Trails, and Attestations fit together.