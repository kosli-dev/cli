---
title: "Security and Compliance"
draft: true
bookCollapseSection: false
weight: 400
summary: "How Security and Compliance teams can use Kosli to define control objectives and verify evidence."
---

# Security and Compliance

You are responsible for ensuring that software delivery meets regulatory, security, or internal governance requirements. You translate frameworks like SOC 2, ISO 27001, or custom internal controls into practical expectations for teams.

You may work in AppSec, GRC, risk management, or a compliance function. You care about provable controls, trustworthy evidence, and making audits repeatable and painless.

## How Kosli helps you

Kosli creates a continuous, tamper-proof record of how software changes move through your organization. It captures real evidence for controls like peer review, test coverage, security scanning, and approval steps, all without relying on spreadsheets or screenshots.

With Kosli, you can:
- Automatically collect and store control evidence for every change
- Get instant visibility into which changes are compliant and which are not
- Replace change request tickets with actual audit-ready data
- Export audit packages in seconds for any service, environment, or release


## Your role in using Kosli

You help define what counts as compliant. Kosli helps you enforce that through policy and automation. Your responsibilities may include:

- Working with platform teams to translate controls into **Attestations** and **Policies**
- Reviewing **Environment** or **Trail** compliance reports
- Verifying that changes meet requirements for deployment to sensitive environments
- Preparing for or responding to internal and external audits using Kosli data

You may not configure pipelines directly, but you rely on Kosli’s outputs to validate that controls are working.

## What you’ll work with

You interact with Kosli through:

- **The Kosli UI**, where you can see compliance status per environment, service, or release
- **Audit Packages**, which you can export to support internal reviews or formal audits
- **Attestation** and **Policy** definitions, often managed in collaboration with platform or security engineering teams
- **Environment Snapshots**, which show what is running and why it is or is not compliant

You may also use the **CLI** or **API** if you need detailed reports or integrations.

## What success looks like

- You can prove to auditors or regulators that your SDLC is secure and compliant
- Controls are codified and enforced consistently across all delivery pipelines
- You no longer chase teams for screenshots or spreadsheets during audits
- You have full traceability from change request to deployed artifact with supporting evidence

## Common questions you might have

**"How do I know a change is compliant?"**<br>
Kosli validates Trails and Environments based on policies and recorded attestations. You can view compliant and non-compliant changes in the UI or export audit reports.

**"Can we map Kosli data to our compliance framework?"**<br>
Yes. Attestations can represent any type of control evidence, such as test results, PR approvals, vulnerability scans, or change reviews.

**"How secure is the evidence?"**<br>
Kosli stores all records immutably and securely. Attestations can include signed metadata and attachments, stored in a tamper-evident Evidence Vault.

**"How do I use Kosli in an audit?"**<br>
You can export a complete Audit Package for any Trail, Artifact, or Environment. This includes all recorded evidence and metadata for traceable, reviewable compliance.

## Where to start

- [**Concepts**]({{< ref "/understand_kosli/concepts" >}}): Understand how Flows, Trails, and Attestations fit together.
