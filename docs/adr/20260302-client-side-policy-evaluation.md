---
title: "20260302 - Client-side policy evaluation with OPA/Rego"
description: "Embed OPA in the CLI for local Rego policy evaluation to enable fast feedback"
status: "Proposed"
date: "2026-03-02"
---

# 20260302 - Client-side policy evaluation with OPA/Rego

## Overview

Embed the Open Policy Agent (OPA) engine in the Kosli CLI to evaluate Rego policies against trail data locally, as an early prototype to learn how OPA/Rego fits with Kosli and to gather customer feedback.

## Context

The `kosli evaluate trail` command needs to assess whether a trail's compliance status satisfies user-defined policies. This is an early prototype of using OPA/Rego with Kosli — the primary goal right now is to learn. We need to get this into customers' hands to gather feedback on the policy authoring experience, the data shape, and how evaluation fits into their workflows.

We believe the final destination for policy evaluation is server-side. Client-side evaluation has a fundamental trust problem: the CLI fetches data from the server, evaluates locally, and posts the result back — data could be modified locally between fetch and evaluation, breaking the chain of trust. Server-side evaluation eliminates this by keeping data and evaluation together. However, the server-side API does not exist yet, and waiting for it would delay learning.

## Decision Drivers

- We need to learn — getting an early prototype to customers shapes the feature before we invest in server-side infrastructure
- Client-side evaluation has a trust gap (fetch → local modify → evaluate → post), which makes server-side the long-term goal
- The solution should be useful now and not block on server-side work
- Embedding in the existing CLI avoids distribution and adoption friction — customers already have the CLI installed and approved, and don't need permission to use an additional tool
- Rego is an established policy language with good tooling and documentation
- The OPA Go library can be embedded directly without requiring a separate daemon

## Options Considered

### Option 1: Server-side evaluation

Send trail data and policy to the Kosli server for evaluation.

**Pros:**
- Central enforcement — policies can't be bypassed
- Server can evolve the evaluation without CLI updates
- No OPA dependency in the binary

**Cons:**
- API doesn't exist yet — blocks the feature entirely
- Adds latency to every evaluation
- Requires network access

### Option 2: Client-side with embedded OPA/Rego

Embed the OPA library in the CLI binary and evaluate Rego policies locally.

**Pros:**
- Fast feedback — evaluation is instant, no network round-trip
- Works offline
- Unblocks the feature immediately
- Rego is a well-known policy language with existing ecosystem
- Users can iterate on policies locally before server-side enforcement

**Cons:**
- Adds OPA as a dependency, increasing binary size
- Policies are local files — no central enforcement
- Will need a migration path when server-side evaluation arrives
- Users must learn Rego

### Option 3: Standalone evaluation tool

Ship a separate CLI tool dedicated to policy evaluation.

**Pros:**
- Decoupled release cycle from the main CLI
- Could be lighter-weight without the rest of the CLI

**Cons:**
- Customers need to install, approve, and distribute an additional tool
- Many customers require permission to use new tools — adds adoption friction
- Duplicates distribution infrastructure already solved by the CLI

### Option 4: Custom DSL

Design a bespoke policy language tailored to Kosli trail data.

**Pros:**
- Could be simpler for common cases
- No external dependency

**Cons:**
- Significant investment to design, implement, and maintain
- Too much work for an interim solution
- Users learn a proprietary language with no transferable skills

## Decision

Embed the OPA library in the CLI for client-side Rego policy evaluation.

**Rationale:** This is a prototype to learn. Getting OPA/Rego evaluation into customers' hands now lets us validate the approach, discover what data policies actually need, and shape the server-side design with real feedback. Rego is an established policy language — users investing time in writing policies gain transferable skills. The OPA Go library integrates cleanly as an embedded dependency.

**Policy contract:** Policies must use `package policy` and declare an `allow` rule (boolean). An optional `violations` rule can provide human-readable denial reasons. Trail data is available as `input`.

**Trade-offs:** We accept the trust gap of client-side evaluation and the OPA dependency in exchange for learning now. When server-side evaluation arrives, we will need to consider how client-side and server-side evaluation coexist or migrate.

## Consequences

**Positive:**
- Enables early customer feedback to shape the policy evaluation feature
- Validates OPA/Rego as a policy language for Kosli before committing to server-side investment
- Rego is well-documented with an active community
- Policies can be version-controlled alongside code

**Negative:**
- Client-side evaluation has a trust gap — data can be modified between fetch and evaluation, so results cannot be fully trusted
- OPA library increases binary size
- No central enforcement — policies can be bypassed
- Migration path needed when server-side evaluation is built

**Neutral:**
- Policy files are local Rego files passed via `--policy` flag
- The `internal/evaluate` package encapsulates all OPA interaction

## Related Decisions

- [20260302-client-side-enrichment-pipeline](20260302-client-side-enrichment-pipeline.md) — client-side data enrichment needed to make trail data policy-friendly
