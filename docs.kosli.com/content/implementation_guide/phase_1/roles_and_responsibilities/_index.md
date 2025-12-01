---
title: "Roles and Responsibilities"
bookCollapseSection: true
weight: 100
---

# Roles and Responsibilities

Kosli supports multiple stakeholders across engineering, security, and compliance. Successful adoption depends on clear ownership and collaboration across roles.
This guide provides:

- A RACI matrix to define responsibilities per phase
- Role-by-role expectations during rollout
- Links to relevant documentation for each group

## 🔄 Phases of Implementation

1. **Discovery and Planning:** Understand what to track, who is involved, and which flows to start with.
2. **Initial Setup and Pilot:** Configure Kosli for a single service or team. Validate the model and gather feedback.
3. **Rollout and Scale:** Extend flows and policies across teams and services. Standardize and automate.
4. **Governance and Optimization:** Measure success, refine policies, and prepare for audits with real data.

## 👥 Stakeholders
1. [**Platform Engineers and DevOps**]({{< relref "platform_engineers" >}}): Leads technical implementation and pipeline integration
2. [**Application Developer**]({{< relref "app_developers" >}}): Builds code and produces evidence automatically
3. [**Security and Compliance**]({{< relref "security_compliance" >}}): Defines control objectives and verifies evidence
4. [**Sponsors**]({{< relref "sponsors" >}}): Champions adoption, aligns on outcomes, and tracks impact

## 📊 RACI Matrix

The RACI model helps teams and stakeholders know who to talk to, who drives a decision, and who just needs visibility. It’s especially helpful when rolling out tools like Kosli across multiple teams with different priorities and domain focus.

| Task                                 | Platform Engineer  | Application Developer | Security & Compliance   | Sponsor |
|--------------------------------------|--------------------|-----------------------|-------------------------|---------|
| Identify key flows and services       | R                  | C                     | C                       | A       |
| Define success criteria and metrics   | C                  | C                     | C                       | A       |
| Select pilot team/service            | R                  | C                     | C                       | A       |
| Set up Kosli CLI and pipelines       | A                  | I                     | C                       | C       |
| Define attestation types              | R                  | C                     | A                       | C       |
| Configure environment snapshots       | A                  | I                     | C                       | C       |
| Set up environment policies          | R                  | I                     | A                       | C       |
| Validate compliance status           | C                  | C                     | A                       | C       |
| Export and review audit packages     | C                  | I                     | A                       | C       |
| Roll out to additional teams         | R                  | C                     | C                       | A       |
| Track measures of success            | R                  | C                     | C                       | A       |

- **A - Accountable**

    The owner of the outcome. This is the person who ensures the task is completed successfully, even if others do the work. There should only be one "A" per task.

- **R - Responsible**

    The doer. This person (or team) performs the work. They are hands-on with the implementation and execution of the task.

- **C - Consulted**

    Someone who provides input, guidance, or subject matter expertise. This is a two-way communication role. Their feedback is important for shaping the work.

- **I - Informed**

    Kept in the loop. This person doesn't need to be consulted during the task but should be notified of progress or outcomes. It’s a one-way communication role.

## Subpages