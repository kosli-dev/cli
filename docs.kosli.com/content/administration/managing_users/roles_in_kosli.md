---
title: Roles in Kosli
bookCollapseSection: false
weight: 100
summary: "Kosli provides three user roles to help administrators manage access and permissions within their organization: Admin, Member, and Reader."
---

# Roles in Kosli

Kosli provides three user roles to help administrators manage access and permissions within their organization. Understanding these roles is essential for assigning the appropriate level of access to your team members.

## Overview

| Role | Description | Best for |
|------|-------------|----------|
| **Admin** | Full control over the organization | Organization owners, security leads, platform engineering leads |
| **Member** | Can create and modify resources | Developers, platform engineers, CI/CD systems |
| **Reader** | Read-only access to view data | Auditors, compliance officers, stakeholders, reporting systems |

## Permissions Matrix

| Capability | Admin | Member | Reader |
|------------|:-----:|:------:|:------:|
| **User Management** | | | |
| Invite and remove users | ✅ | ❌ | ❌ |
| Change user roles | ✅ | ❌ | ❌ |
| **Organization Settings** | | | |
| Modify organization settings | ✅ | ❌ | ❌ |
| Configure integrations (Slack, LaunchDarkly) | ✅ | ✅ | ❌ |
| **Service Accounts** | | | |
| Create and manage service accounts | ✅ | ✅ | ❌ |
| Generate service account API keys | ✅ | ✅ | ❌ |
| **Resource Management** | | | |
| Create flows | ✅ | ✅ | ❌ |
| Update/delete flows | ✅ | ✅ | ❌ |
| Create/update environments | ✅ | ✅ | ❌ |
| Delete environments | ✅ | ❌ | ❌ |
| Create/update policies | ✅ | ✅ | ❌ |
| Delete policies | ❌ | ❌ | ❌ |
| Create attestation types | ✅ | ✅ | ❌ |
| Update/delete attestation types | ✅ | ✅ | ❌ |
| **Attestations & Snapshots** | | | |
| Report attestations | ✅ | ✅ | ❌ |
| Report environment snapshots | ✅ | ✅ | ❌ |
| Create and manage approvals | ✅ | ✅ | ❌ |
| **Actions** | | | |
| Create, update, and delete actions | ✅ | ✅ | ❌ |
| View actions | ✅ | ✅ | ✅ |
| **Data Access** | | | |
| View trails and artifacts | ✅ | ✅ | ✅ |
| View attestations | ✅ | ✅ | ✅ |
| View snapshots | ✅ | ✅ | ✅ |
| Query and search data | ✅ | ✅ | ✅ |
| Export and generate reports | ✅ | ✅ | ✅ |
| View flow/policy configurations | ✅ | ✅ | ✅ |

---

## Admin

Administrators have full control over the organization and its resources.

### Permissions

Admins can perform all actions in Kosli, including:

- **User Management**: Invite, remove, and change roles of organization members (Admin only)
- **Organization Settings**: Modify organization-wide settings and configurations (Admin only)
- **Service Accounts**: Create and manage service accounts and their API keys
- **Integrations**: Configure integrations with external systems (Slack, LaunchDarkly, etc.)
- **Resource Management**: Create, update, and delete flows, environments, policies, and attestation types
- **Attestations & Snapshots**: Report attestations, environment snapshots, and manage approvals
- **Actions**: Create, update, and delete actions for automated workflows and notifications
- **Data Access**: View all trails, artifacts, attestations, and snapshots

### When to assign

Assign the Admin role to:
- Organization owners or senior leaders responsible for overall Kosli implementation
- Security engineers who need to manage user access and compliance processes
- Platform engineering leads who need to configure integrations and manage organization settings

{{% hint warning %}}
Limit the number of Admins to maintain security and control over your organization. Most users should be Members or Readers.
{{% /hint %}}

---

## Member

Members can create and modify resources, manage service accounts, and configure integrations, but cannot manage users or organization-wide settings.

### Permissions

Members can:

- **Service Accounts**: Create and manage service accounts and their API keys
- **Integrations**: Configure integrations with external systems (Slack, LaunchDarkly, etc.)
- **Resource Management**: Create, update, and delete flows, environments, policies, and attestation types
- **Attestations & Snapshots**: Report attestations, environment snapshots, and manage approvals
- **Actions**: Create, update, and delete actions for automated workflows and notifications
- **Data Access**: View all trails, artifacts, attestations, and snapshots

Members cannot:
- Manage users or change user roles
- Modify organization-wide settings

### When to assign

Assign the Member role to:
- Platform engineers who need to implement Kosli across teams and manage service accounts
- Application developers who need to report attestations and manage flows
- Team leads who need to configure integrations and create service accounts for their teams
- CI/CD systems that need to report attestations and snapshots (via service accounts)

---

## Reader

Readers have read-only access to view data in Kosli without the ability to create or modify resources.

### Permissions

Readers can:

- **View Data**: Access trails, artifacts, attestations, and snapshots
- **Query Information**: Search and filter data across flows and environments
- **Generate Reports**: Export and analyze compliance data
- **View Configurations**: See flow definitions, policies, attestation types, and actions (but cannot modify them)

Readers cannot:
- Create, update, or delete any resources
- Report attestations or snapshots
- Manage approvals
- Create or manage actions
- Create or manage service accounts
- Configure integrations
- Invite users or change settings

### When to assign

Assign the Reader role to:
- Auditors who need visibility into compliance data
- Compliance officers reviewing attestation and deployment history
- Stakeholders and executives who want to monitor software delivery
- Reporting and monitoring systems that query Kosli data for dashboards

---

## Assigning Roles

To assign or change a user's role:

1. Log in to Kosli as an Admin
2. Navigate to your organization from the left navigation menu
3. Go to `Settings` > `Members`
4. Find the user you want to modify
5. Select their new role from the dropdown menu

{{% hint info %}}
Role changes take effect immediately. Users will see their updated permissions the next time they interact with Kosli.
{{% /hint %}}

---

## Best Practices

### Follow the Principle of Least Privilege

Assign users the minimum role required to perform their job functions. Start with Reader access and increase permissions as needed.

### Use Service Accounts for Automation

For CI/CD pipelines and automated systems, create service accounts with the Member role rather than using personal API keys. This provides better auditability and security.

### Regular Access Reviews

Periodically review user roles and remove access for team members who no longer need it. This is especially important when people change roles or leave the organization.

### Separate Concerns

- **Admins**: Focus on governance, security, and organization-wide configuration
- **Members**: Handle day-to-day operations and resource management
- **Readers**: Provide visibility without risk of accidental changes

---

## Mapping Roles to Your Organization

When implementing Kosli, you need to map organizational roles to Kosli user roles. This table provides recommended mappings based on typical responsibilities:

| Organizational Role | Recommended Kosli Role | Alternative | Rationale |
|---------------------|------------------------|-------------|-----------|
| **Platform Engineers** | Member | Admin (for leads) | Platform engineers need to set up flows, manage service accounts, configure integrations, and implement Kosli across teams. Member role provides these capabilities. Lead platform engineers managing the overall setup may need Admin access. |
| **Application Developers** | Member | Reader (for view-only) | Developers typically need to report attestations and manage flows for their applications. Member role enables this. Some developers may only need visibility into deployments and compliance status, making Reader sufficient. |
| **Security & Compliance** | Admin | N/A | Security and compliance teams need to manage policies, review audit data, control user access, and configure organization-wide settings. Admin role is required for these governance responsibilities. |
| **Sponsors** | Reader | N/A | Sponsors need visibility into adoption progress, compliance status, and overall system health but don't need to make technical changes. Reader role provides necessary oversight without operational access. |

### Understanding the Mapping

This mapping is a starting point. Your organization's structure and responsibilities may require adjustments:

- **Small teams**: Developers might need Admin access if they handle all aspects
- **Large enterprises**: Strict separation may require more Readers, fewer Admins
- **Regulated industries**: Security teams might need dedicated Admin accounts separate from operations

The key principle: Assign the minimum role required for someone to fulfill their responsibilities effectively.

### Learn More About Organizational Roles

For detailed guidance on each organizational role's responsibilities during Kosli implementation, see:

- [Implementation Guide: Roles and Responsibilities]({{< ref "/implementation_guide/phase_1/roles_and_responsibilities" >}})
- [Platform Engineers]({{< ref "/implementation_guide/phase_1/roles_and_responsibilities/platform_engineers" >}})
- [Application Developers]({{< ref "/implementation_guide/phase_1/roles_and_responsibilities/app_developers" >}})
- [Security & Compliance]({{< ref "/implementation_guide/phase_1/roles_and_responsibilities/security_compliance" >}})
- [Sponsors]({{< ref "/implementation_guide/phase_1/roles_and_responsibilities/sponsors" >}})

