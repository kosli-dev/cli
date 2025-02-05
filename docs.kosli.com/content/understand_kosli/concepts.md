---
title: 'Concepts'
weight: 130
summary: "This section helps you understand the concepts Kosli is built on. The figure below gives an overview of the main Kosli concepts and how they are related to each other."
---

# Concepts

This section helps you understand the concepts Kosli is built on. The figure below gives an overview of the main Kosli concepts and how they are related to each other.

{{<figure src="/images/kosli_concepts.png" alt="Kosli Concepts" width="900">}}

## Organization

A Kosli organization is an account that owns Kosli resources, such as Flows and Environments. Only members within an organization can access its contents.

When signing up for Kosli, a personal organization is automatically created for you, bearing your username. This personal organization is exclusively accessible to you. Additionally, you can create `Shared` organizations and invite multiple team members to collaborate on different Flows and Environments.

## Flow

A Kosli Flow represents a business or software process for which you want to track changes and monitor compliance.

### Trail

A Kosli Trail represents a single execution instance of a process represented by a Kosli Flow.
Each Trail must have a unique identifier of your choice, based on your process and domain. Example identifiers include git commits or pull request numbers.
  
#### Artifact

Kosli Artifacts represent the software artifacts generated from every execution, portrayed as a Trail, of your software process depicted as a Flow. These artifacts play a crucial role in enabling **Binary Provenance**, providing a comprehensive chain of custody that records the origin, history, distribution, and execution details of each artifact.

Each Artifact is uniquely identified by its SHA256 fingerprint. Using this fingerprint, Kosli can link the creation of the Artifact with its runtime-related events, such as when the artifact starts or concludes execution within a specific Environment.

#### Attestation

An Attestation is a declaration about whether a particular Artifact or Trail adheres to a certain requirement or not. It is normally reported after performing a specific risk control or quality check (e.g. running tests). The attestation encompasses the procedure's results.

Kosli supports reporting specific types of attestations (e.g., a snyk scan, sonarcloud scan, junit tests) and a generic one for other use cases.

##### Evidence Vault

Attestations in Kosli have the capability to contain additional evidence files attached to them. This supporting evidence is securely stored within Kosli's evidence vault and is retrievable on demand.

## Audit package

During an audit process, Kosli enables you to download an audit package for a Trail, Artifact, or an individual Attestation. This package comprises a tar file containing metadata related to the selected resource, alongside any evidence files that have been attached. The audit package serves as a comprehensive collection of information aiding in audit-related investigations or reviews.

## Flow Template

A Flow Template defines the expected attestations for Flow Trails and Artifacts to be considered compliant. While each Flow has its own Template, each Trail in a Flow can override the Flow Template with its own.

## Environment

Environments in Kosli monitor changes in your software runtime systems.

Each physical or virtual runtime environment you want to track in Kosli should have its own Kosli Environment created. Kosli allows you to portray your environments precisely. For instance, with a Kubernetes cluster, you can treat it as one Kosli Environment or designate one or more namespaces in the cluster as separate Kosli Environments.

Kosli supports various types of runtime environments:
* Kubernetes cluster (K8S)
* Amazon ECS
* Amazon S3
* Amazon Lambda
* Physical/virtual server
* Docker host

### Environment Snapshot

An Environment Snapshot represents the reported status (running Artifacts) of your runtime environment at a specific point in time. Snapshots are immutable, append-only objects. Once a snapshot is created, it cannot be modified.

In each snapshot, Kosli links the running artifacts to the Flows and Trails that produced them. Snapshot compliance relies on the compliance status of each running artifact, while Environment compliance depends on its latest snapshot compliance.

Running artifacts that come from 3rd party sources, can be `allow-listed` in an Environment to make them compliant. 
