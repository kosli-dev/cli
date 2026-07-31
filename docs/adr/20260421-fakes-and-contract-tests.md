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

This created three problems:

1. **Test suite reliability**: Any test that calls a live external service can fail due to network issues, rate limits, credential expiry, or upstream changes — causes entirely unrelated to our code.
2. **Test suite speed and scope**: Tests that depend on real external state can only verify a narrow set of scenarios. It is impractical to test error paths or edge cases against live services.
3. **Who can run the tests**: Developers without credentials for a given external service could not run the tests for commands that depend on it. This limited who could contribute to those areas of the codebase.

The pattern was first introduced for AWS Lambda (#763), then extended to GitHub (#807) and AWS S3 (#758).

## Decision

For each external service integration, we:

1. **Define an interface** at the operation level that the command layer depends on, expressing what the service does in domain terms rather than SDK terms.

2. **Write a shared contract test function** that asserts the key behavioural properties of the interface — what fields are present, what error behaviour to expect, and how edge cases are handled.

3. **Run the contract tests against both implementations**: the shared contract test function is invoked twice — once against the fake (in the regular test suite, no credentials required) and once against the real service (env-gated, called from `make test_contract`). Running the same suite against both is what guarantees they behave identically.

4. **Build an in-memory fake** that satisfies the same interface and passes the same contract tests. The fake is the only implementation used in the main integration test suite.

5. **Inject the fake** into the command layer via a package-level factory variable. Tests swap the factory in `SetupTest` and restore it in `TearDownTest`.

## The interface abstraction level

An interface can sit at the **SDK client level** (individual SDK calls, e.g. `ListObjectsV2`) or the **operation level** (domain terms, e.g. "get PR evidence for this commit"). We use both, and pick per integration:

- **Prefer SDK-level** when the SDK client is an ordinary Go type whose methods a fake can implement directly, as with the AWS SDK v2 service clients. The real client then satisfies the interface implicitly, so there is no adapter to write and no adapter to keep correct. It also keeps behaviour like pagination in our own code, where a contract test can verify it against the real API rather than hiding it behind a wrapper we cannot exercise.

- **Fall back to operation-level** when SDK-level faking would mean reimplementing SDK machinery. GitHub's GraphQL client uses Go reflection internally, so a fake at that level is impractical. The same applies to SDK helpers with substantial internals of their own — the S3 transfer manager's `DownloadObject` drives ranged `GetObject`/`HeadObject` calls, so a fake stands in for the download *operation* and writes bytes straight to the destination.

Either way the interface stays **narrow**: only the operations this codebase actually calls. That is what bounds the contract test surface — we are testing our expectations of the dependency, not the dependency itself.

A single seam may mix the two levels. AWS S3 does: `S3ListAPI` is SDK-level (`ListObjectsV2`, so the SDK paginator still runs against our interface and the continuation-token contract is verified against real S3), while `S3DownloadAPI` is operation-level (the transfer manager's `DownloadObject`).

## Contract tests vs the main integration suite

The shared contract test function is run against **both** implementations:

- **Fake**: runs as part of the regular test suite (`make test_integration`), requires no external credentials. This is what keeps the fake honest on every PR.
- **Real service**: runs via `make test_contract` in CI on a schedule, gated on service credentials. This is the authoritative run that verifies the contract against the actual API.

Running the same test function against both implementations is the mechanism that guarantees they behave the same way. If the real API changes and the real-service run starts failing, it signals that the fake must be updated to match.

The main integration test suite runs against the local Kosli server only and uses fakes for all external service calls. This keeps the main suite fast and deterministic.

## Consequences

- The main integration suite no longer requires live AWS or GitHub credentials to run. Any developer can run the full suite and contribute to commands that depend on external services.
- Behavioural contracts between fakes and real APIs are made explicit and machine-checked.
- Adding a new external service integration requires writing a contract test before (or alongside) the fake, which documents the real API behaviour.
- The fake must be kept honest: if the real API changes in a way that breaks the contract tests, the fake must be updated to match.
