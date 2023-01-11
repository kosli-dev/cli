---
title: Get familiar with Kosli
bookCollapseSection: false
weight: 220
---

# Get familiar with Kosli

The following guide is not a real life use case for Kosli - usually you'd run Kosli in your CI and remote environments. But this is the easiest and quickest way to try Kosli out and understand it's features. So you can try it out using just your local machine and `docker`. In our *How to* and *Kosli integrations* sections you'll find all the information needed to run it in actual projects.

In this tutorial, you'll learn how Kosli allows you to follow a source code change to runtime environments.
You'll set up a `docker` environment, use Kosli to record build and deployment events, and track what artifacts are running in your runtime environments. 

This tutorial uses the `docker` Kosli environment type, but the same steps can be applied to other supported environment types.

{{< hint info >}}
As you go through the guide you can also check your progress from 
[your browser](https://app.kosli.com).

In the upper left corner there is a house icon. Next to it you can select
which organization you want to view. Your personal organization
has the same name as your github login name, and is the organization (in the context of Kosli CLI called "owner") you will
be using in this guide.
{{< /hint >}}