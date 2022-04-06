---
title: 'Report Environment'
weight: 10
---

# Report Environment

## Create an environment in Merkely

The first thing we need to configure is an **environment** in [Merkely](https://app.merkely.com).  
Merkely **Environment** is where you'll be reporting the state of your actual environments, like *staging* or *production*. 

When you log in to Merkely the **Environments** page is the first thing you see. If you clicked around before reading this guide you'll find link to **Environments** on the left side of the window in Merkely. 

Click "Add environment" button to create a new Merkely **environment**. On the next page you'll have to select the type - for the purpose of this guide we'll use 'Kubernetes cluster'.

Next you need to give your **environment** a name - it doesn't have to be the same name you use for the actual environment, but it certainly helps to identify it in the future. In the guide we'll use **test-env** as a name of the **environment**.
You also need to provide the description of the environment (that becomes useful once the amount of your environments grow)

Click "Save Environment" and you're ready to move on to the next step.

## Report an environment

Time to implement an actual reporting of what's running in your k8s cluster - which means we need to reach out to the cluster and check which docker images were used to run the containers that are currently up in given namespace. 

### CLI

You report the environment using [Merkely CLI tool](https://github.com/merkely-development/cli/releases).  
You need to download a correct package depending of the architecture of the machine you use to run the CLI. 

You can run the [command](https://docs.merkely.com/client_reference/merkely_environment_report_k8s/) manually on any machine that can access your k8s cluster, but it is much better to automate the reporting from the start, and we'll use GitHub Actions for that.

### GitHub workflow

TODO: reporting credentials to use in the workflow

```
TODO: GitHub workflow
- download CLI
- report env
merkely environment report k8s -n namespace test-env
```

Reporting an **environment** is an easy way to get the answer to a question like: "What is running in production?". 
Once you know the answer to that, the next thing you may want to figure out is: "Is it verified?" and reporting your artifacts in Merkely will cover that part.