---
title: "20260421 - Fakes and contract tests for external service integrations"
description: "Use in-memory fakes and contract tests to decouple integration test suites from live external services"
status: "Accepted"
date: "2026-04-21"
---

# 20260421 - Fakes and contract tests for external service integrations

## Overview

Use in-memory fakes and a shared contract test suite to decouple the main integration test suite from live external services (AWS Lambda, GitHub API, etc.), while ensuring the fakes remain faithful to the real APIs.

## Context

Several CLI commands interact with external services — AWS Lambda, GitHub, GitLab, Bitbucket, Azure DevOps — to gather evidence for attestations and snapshots. The integration tests for these commands previously required live credentials and network access to real external services.

This created two problems:

1. **Test suite reliability**: Any test that calls a live external service can fail due to network issues, rate limits, credential expiry, or upstream changes — causes entirely unrelated to our code.
2. **Test suite speed and scope**: Tests that depend on real external state can only verify a narrow set of scenarios. It is impractical to test error paths or edge cases against live services.

The pattern was first introduced for AWS Lambda (#763) and then extended to GitHub (#807).

## Decision

For each external service integration, we:

1. **Define an interface** at the operation level that the command layer depends on, expressing what the service does in domain terms rather than SDK terms.

2. **Write a shared contract test function** that asserts the key behavioural properties of the interface — what fields are present, what error behaviour to expect, and how edge cases are handled.

3. **Run the contract tests against the real service** (env-gated, called from `make test_contract`). This is the authoritative run that documents what the real API actually does.

4. **Build an in-memory fake** that satisfies the same interface and passes the same contract tests. The fake is the only implementation used in the main integration test suite.

5. **Inject the fake** into the command layer via a package-level factory variable. Tests swap the factory in `SetupTest` and restore it in `TearDownTest`.

## The interface abstraction level

We chose to fake at the **operation level** rather than the **SDK client level**. This means the interface speaks in domain terms (e.g. "get PR evidence for this commit") rather than SDK terms (individual HTTP, GraphQL, or SDK calls).

This keeps fakes simple and free of SDK types, and the interface boundary is stable even if the underlying SDK or transport changes. In some cases (e.g. GraphQL clients that use Go reflection internally) SDK-level faking is impractical without reimplementing SDK machinery, which makes operation-level faking the only viable option.

## Contract tests vs the main integration suite

The contract tests run against real external services and are **separate from the main integration test suite**. They run in CI on a schedule via `make test_contract` and require credentials that are not available during regular PR builds.

The main integration test suite runs against the local Kosli server only and uses fakes for all external service calls. This keeps the main suite fast and deterministic.

## Consequences

- The main integration suite no longer requires live AWS or GitHub credentials to run.
- Behavioural contracts between fakes and real APIs are made explicit and machine-checked.
- Adding a new external service integration requires writing a contract test before (or alongside) the fake, which documents the real API behaviour.
- The fake must be kept honest: if the real API changes in a way that breaks the contract tests, the fake must be updated to match.
