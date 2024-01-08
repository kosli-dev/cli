---
title: 'What is Kosli?'
weight: 120
---
# What is Kosli?

Kosli is a change recording and compliance monitoring platform which allows you to record, track and query changes about any software or business process so you can prove compliance and maintain security without slowing down.

Kosli connects the recorded changes to establish immutable "chains of custody" which enables you to:

1. **Track Changes**: Trace how your business or software processes change over time.
2. **Identify Sources**: Understand where changes originated from, which can help in identifying issues.
3. **Continuous Compliance**: Ensure that you continuously adhere to your compliance requirements.
4. **Enable Audits**: Access audit packages on demand allowing audits and investigations into the software supply chain.
5. **Enhance Trust**: Build trust among users, customers, and stakeholders by providing transparent and verified information about the change history.

# When to use Kosli?

The following are some example use cases where you can use Kosli to ensure compliance:

- Monitor and prove that your software delivery is compliant with your requirements and policies. Example policies are:
  - All artifacts running in production have passed security scanning.
  - All code that goes to production must have reviewed in a pull request.

- Monitor and prove that some business process is compliant with your requirements and policies. Example use cases are:
  - Employee onboarding/off-boarding
  - Recording production server access

# Where does Kosli fit in the growing tools landscape?

Kosli is tool-agnostic. It is designed to work with any tools (CI systems, code analysis tools, runtime environments, etc.). Kosli can function as the compliance and change data hub aggregating data from all your other tools in one place; your compliance and change single pane of glass. 

# How does Kosli work?

Kosli works like a black box recorder. It is an append-only store of immutable change records. You report changes you care about to Kosli via CLI or API, Kosli stores them and actively monitors whether you are compliant with your policies or not.

Change sources can come from build systems (e.g. CI systems) and from runtime environments (e.g. a Kubernetes cluster).

{{<figure src="/images/kosli-overview-docs.jpg" alt="Kosli overview" width="1000">}}


