---
title: 'Prerequisites'
weight: 1
---

# Prerequisites

To follow the "Getting Started" guide you'll need to set up a few things:

1. Kosli account
2. GitHub repository where you'll store your code (you can fork our demo repository) 
3. Your own k8s cluster where you'll deploy your application
4. hub.docker.com account

## GitHub
If you want to use out workflow examples, there's a few things you need to configure in your GitHub repository - you can fork the [github-k8s-demo repository](https://github.com/merkely-development/github-k8s-demo), create your own from scratch, or use an already existing project. 

Workflows in [github-k8s-demo repository](https://github.com/merkely-development/github-k8s-demo) are complete version of workflows we're developing in this guide.

In our example we use Google Cloud to host k8s cluster and we rely on `google-github-actions/get-gke-credentials` action to authenticate to GKE cluster via a `kubeconfig` file. If you're hosting your k8s cluster somewhere you need to use a different action.

### Secrets

Create following Actions Secrets in your repository on GitHub:
* **MERKELY_API_TOKEN** - you can find the Kosli API Token under your profile at https://app.merkely.com/ (click on your avatar in the right top corner of the window and select 'Profile')
* **GCP_K8S_CREDENTIALS** - service account credentials (.json file), with k8s access permissions
* **DOCKERHUB_TOKEN** - your DockerHub Access Token (you can create one at hub.docker.com, under *Account Settings* > *Security*)



Once these are in place you're ready to go!


