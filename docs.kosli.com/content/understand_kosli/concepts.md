---
title: 'Concepts'
weight: 130
summary: "This section helps you understand the concepts Kosli is built on. The figure below gives an overview of the main Kosli concepts and how they are related to each other."
---

# Concepts

This section helps you understand the concepts Kosli is built on. The figure below gives an overview of the main Kosli concepts and how they are related to each other.

{{<figure src="/images/kosli-concepts.jpg" alt="Kosli Concepts" width="900">}}

## Organization

A Kosli organization is an account that owns Kosli resources, such as Flows and Environments. Only members within an organization can access its resources.

When signing up for Kosli, a personal organization is automatically created for you, bearing your username. This personal organization is exclusively accessible to you. Additionally, you can create `Shared` organizations and invite multiple team members to collaborate on different Flows and Environments.

## ![icon](/images/diagram-elements/kosli-icon-round-flows.png) Flow

A Kosli Flow represents a business or software process for which you want to track changes and monitor compliance.

As an example, a flow can be created to track the controls involved with building an application in your CI system.

### ![icon](/images/diagram-elements/kosli-icon-round-trails.png) Trail

A Kosli Trail represents a single execution instance of a Kosli Flow.
Each Trail must have a unique identifier of your choice, based on your process and domain. Example identifiers include git commits or pull request numbers.
  
### ![icon](/images/diagram-elements/kosli-icon-round-artifact-green.png) Artifact

Kosli Artifacts represents the software artifacts generated from your CI pipeline.
Its creation is captured in Kosli through the trail.

When captured by Kosli, the Artifact is uniquely identified by its SHA256 fingerprint. Using this fingerprint, Kosli can link the creation of the Artifact with its runtime-related events, such as when the artifact starts or concludes execution within a specific Environment.

These artifacts play a crucial role in enabling **Binary Provenance**, providing a comprehensive chain of custody that records the origin, history and distribution of each artifact.

### ![icon](/images/diagram-elements/kosli-icon-round-attestations-2.png) Attestation

An Attestation is a record of compliance checks or controls that have been performed a particular Artifact or Trail. It is normally reported after performing a specific risk control or quality check (e.g. running tests). The attestation encompasses the procedure's results.

Kosli provides specific built-in types of attestations (e.g., a snyk scan, sonar scan, junit tests) and allows to define your own custom types with jq quering for compliance status, or generic ones that simply creates an attestation without evaluation.

Attestations can be connected either to the trail, or to a specific artifact.

### ![icon](/images/diagram-elements/kosli-icon-round-vault.png) Evidence Vault

Attestations in Kosli have the capability to contain additional evidence files attached to them. This supporting evidence is securely stored within Kosli's evidence vault and is retrievable on demand.

## ![icon](/images/diagram-elements/kosli-icon-round-package.png) Audit package

During an audit process, Kosli enables you to download an audit package for a Trail, Artifact, or an individual Attestation. This package comprises a tar file containing metadata related to the selected resource, alongside any evidence files that have been attached. The audit package serves as a comprehensive collection of information aiding in audit-related investigations or reviews.

## Flow Template

A Flow Template defines the expected Artifacts and attestations for Trails of that given Flow to be considered compliant.

While each Flow has its own Template, each Trail in a Flow can override the Flow Template with its own.

## ![icon](/images/diagram-elements/kosli-icon-round-environment.png) Environment

Environments in Kosli monitor changes in your software runtime systems.

Each physical or virtual runtime environment you want to track in Kosli should have its own Kosli Environment created. Kosli allows you to portray your environments precisely. For instance, with a Kubernetes cluster, you can treat it as one Kosli Environment or designate one or more namespaces in the cluster as separate Kosli Environments.

Kosli supports various types of runtime environments:

* Kubernetes cluster (K8S)
* Amazon ECS
* Amazon S3
* Amazon Lambda
* Physical/virtual server
* Docker host
* Azure Web Apps and Function Apps

### ![icon](/images/diagram-elements/kosli-icon-round-snapshots.png) Environment Snapshot

An Environment Snapshot represents the reported status (running Artifacts) of your runtime environment at a specific point in time.

In each snapshot, Kosli links the running artifacts to the Flows and Trails that produced them. Snapshot compliance relies on the compliance status of each running artifact, while Environment compliance depends on its latest snapshot compliance.

Running artifacts that come from 3rd party sources, can be `allow-listed` in an Environment to make them compliant.

### ![icon](/images/diagram-elements/kosli-icon-round-policy.png) Environment Policy

Environment Policy enables you to define and enforce compliance requirements for artifact deployments across different environments.
