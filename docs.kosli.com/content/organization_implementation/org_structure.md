---
title: "Org structures"
weight: 2
summary: ""
---

# Org structrues


### Production

This is your live environment for customer-facing applications. It's the destination for all final data and workflows, and data here is considered business critical.

### Pre-Production (Dry Run)

This environment is used to test compliance workflows and processes before they impact live systems. For example, it's a safe space for new teams to test attestation workflows during onboarding without cluttering the production environment. Data here is temporary.

### Development (CI/CD)

The Development environment is where the compliance teams test new features and attestation types.
It's a key part of your CI/CD pipeline, used for automated smoke tests and quality checks to ensure new code works as expected. Data in this environment is temporary.

### Sandbox (Personal)

This is a personal workspace for individual developers to experiment and learn. It's a flexible, low-risk environment for trying out new tools or configurations, separate from shared development resources. Data in this environment is temporary.
