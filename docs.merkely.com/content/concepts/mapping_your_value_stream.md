---
title: 'Mapping your value stream'
weight: 3
---

# Mapping your value stream

When implementing DevOps Change Management, the first step is to uncover your process as a team.  The way we most often do this is to run a value stream mapping exercise with the team.  This can be as simple as spending 30 minutes around a whiteboard, or as involved as you like.

Many aspects of the process will be implemented in your devops automation and tooling such as:

* How you use version control
* How you build
* How you test & qualify
* How you deploy

You may also want to define the role people play in the process (what we call human-in-the loop needs)

* Code review/pull request expectations
* Deployment sign-off
* Validation

![Team looking at a whiteboard with excitement!](/images/whiteboard.jpg)

> **What are human-in-the-loop controls?**
>
> While we would like to believe that everything in your process can be automated, the reality is that many controls can only be performed by real people.  A common example is in the financial sector, where the need for code review is mandated to reduce insider threat risks. We call these human-in-the-loop controls.
> 
> How does this fit into Devops Change Management?  Even though these essential risk controls must be performed by humans, we can still improve the processes by taking a devops approach:
>
> * **Automate the control**: check that the control has been performed programmatically in the devops pipeline
> * **Automate the audit trail**: document the control has been performed programatically in the devops pipeline
> * **Automate the audit trail**: document the control has been performed programatically in the devops pipeline

Once you have mapped out your value steam, the next stage is to model this in your devops tooling.