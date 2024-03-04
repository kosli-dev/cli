---
title: 'Concepts'
weight: 130
---

# Concepts

This section helps you understand the concepts Kosli is built on. The figure below gives an overview of the main Kosli concepts and how they are related to each other.

{{<figure src="/images/kosli_concepts.png" alt="Kosli Concepts" width="900">}}

## Organization

A Kosli organization is an account that owns Kosli resources, such as Flows and Environments. Only members within an organization can access its contents.

When signing up for Kosli, a personal organization is automatically created for you, bearing your username. This personal organization is exclusively accessible to you. Additionally, you can create `Shared` organizations and invite multiple team members to collaborate on different Flows and Environments.

## Flow

A Kosli flow represents a business or software process for which you want to track changes and monitor compliance.

### Trail

A Kosli trail represents a single execution instance of a process represented by a Kosli flow.
Each trail must have a unique identifier of your choice, based on your process and domain. Example identifiers include git commits or pull request numbers.
  
#### Artifact

Kosli artifacts represent the software artifacts generated from every execution, portrayed as a Trail, of your software process depicted as a Flow. These artifacts play a crucial role in enabling **Binary Provenance**, providing a comprehensive chain of custody that records the origin, history, distribution, and execution details of each artifact.

Each artifact is distinctly identified by its SHA256 fingerprint. Utilizing this fingerprint, Kosli can effectively link the creation of the artifact with its runtime-related events, such as when the artifact starts or concludes execution within a specific environment.

#### Attestation

An attestation is a declaration about whether a particular Artifact or Trail adheres to a certain requirement or not. It is normally reported after performing a specific risk control or quality check (e.g. running tests). The attestation encompasses the procedure's results

Kosli supports reporting specific types of attestations (e.g., a snyk scan) and a generic one for other use cases.

##### Evidence Vault

Attestations in Kosli have the capability to contain additional evidence files attached to them. This supporting evidence is securely stored within Kosli's evidence vault and is retrievable on demand.

## Audit package

During an audit process, Kosli enables you to download an audit package for a trail, artifact, or an individual attestation. This package comprises a tar file containing metadata related to the selected resource, alongside any evidence files that have been attached. The audit package serves as a comprehensive collection of information aiding in audit-related investigations or reviews.

## Flow Template

A flow template defines the expected attestations for flow trails and artifacts to be considered compliant. While each flow has its own template, each trail in a flow can override the flow template with its own.

## Environment

Environments in Kosli monitor changes in your software runtime systems.

Each physical or virtual runtime environment you want to track in Kosli should have its own Kosli environment created. Kosli allows you to portray your environments precisely. For instance, with a Kubernetes cluster, you can treat it as one Kosli environment or designate one or more namespaces in the cluster as separate Kosli environments.

Kosli supports various types of runtime environments:
* Kubernetes cluster (K8S)
* Amazon ECS
* Amazon S3
* Amazon Lambda
* Physical/virtual server
* Docker host

### Environment Snapshot

An environment snapshot represents the reported status (running artifacts) of your runtime environment at a specific point in time. Snapshots are immutable, append-only objects. Once a snapshot is created, it cannot be modified.

In each snapshot, Kosli links the running artifacts to the Flows that produced them. Snapshot compliance relies on the compliance status of each running artifact, while environment compliance depends on its latest snapshot compliance.

Running artifacts that come from 3rd party sources, can be `allow-listed` in an environment to make them compliant. 


