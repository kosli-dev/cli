---
title: 'DevOps and Change Management'
weight: 2
---

# DevOps and Change Management

If you work in regulated industries such as finance, medical, and retail, or even just need to follow certain industry standards such as ISO - the way you make software has compliance needs.

At a high level, all software processes have three components:

1. Process: You must have a defined (documented) way of working
2. Implementation: You must follow this process
3. Proof: You must be able to prove that you have followed this process

So how are these needs typically met in traditional IT change management vs DevOps Change management?

| Approach    | Process            | Implementation                              | Proof                                    |
|-------------|--------------------|---------------------------------------------|------------------------------------------|
| Traditional | Wiki Documentation | Manual build, qualification, and deployment | Meeting minutes and Change Documentation |
| DevOps Lite | Wiki Documentation | DevOps Pipelines                            | Meeting minutes and Change Documentation |
| DevOps      | Live Documentation | DevOps Pipelines                            | DevOps Journal                           |

It can be tempting to adopt DevOps and carry on with traditional change management approaches. In this setup, you implement continuous integration and continuous delivery, and when it comes time to release you create the necessary proof that processes have been followed.

![Diagram of change management as a gate](/images/change_management_as_a_gate.png)

This approach provides the illusion of speed, however it comes with some serious consequences.  While the team can feel that they are efficient and able to work in fast loops, they are not connected to the final delivery to the customer until long after the work is complete. This batching of changes increases lead times, slows feedback, increases the risk that a change will fail, and makes it difficult to debug failures that occur.  We call this the DevOps-Lite Trap, or DOLT.. (insert picture of homer simpson)


> **How do you know if you are in the devops-lite trap?**
> 
> You could be in the trap if your team has CI which automatically builds and tests their software but:
> * Cannot deploy on demand
> * Depend upon another team to deliver your work
> * Are not responsible for operating the application

A better approach is to automatically bake in as much change management and risk controls as possible into every change, using DevOps Change Management.

![Diagram of continuous change management](/images/continuous_change_management.png)

This approach:

* Reduces manual work
* Improves process conformance
* Shortens lead time for changes
* Lowers deployment risks


