<!-----

Using this Markdown file:

1. Paste this output into your source file.
2. See the notes and action items below regarding this conversion run.
3. Check the rendered output (headings, lists, code blocks, tables) for proper
   formatting and use a linkchecker before you publish this page.

----->

## Operator Capabilities

### Summary and Problem Definition:

The following is a proposal to improve the definitions of the five operator capability levels. More detailed definitions will allow us to more accurately and efficiently determine the capability level of an operator.

We have another document that was created before for the initial implmentation of automated audits [here](https://docs.google.com/document/d/14nK3krb_cctL8cfk1Fwu1dkAEfo9ZC_iaGCYN4eKMmY/edit#). The goal here is different. The goal is defining the levels not the automation tooling in the sense of implementation. That's another track that is also being treated by our team. The information defined here will feed the directions and implementations on the opcap tools on running capability audits against the operators and their operands.

Two main key principles are supposed to be present when evaluating those capabilities: generally applicable and test driven capability definitions. 

### Generally Applicable Capability Definitions:

**Applicable Precise Definitions:** a description that can be generalized, abstracted and applied widely to any operator regardless of domain or level being checked. Some flexibility may be used here. For example clustered operands vs non clustered operands. Those two categories will have different ways for testing resilience. So trying to test specific behaviors should be avoided. That said we need to find the most relevant aspects of automation that apply generally if not to all, to most operators letting very few exceptions to be treated as exceptions.

### Test Driven Capability Definitions:

The second thing to keep in mind is **test driven definitions.** Beyond being applicable to all operators they need to include how to test. Not in its implementation specifics but with what methodology. Is it a failover test? Is it a metric collected through some url? Is it checking readiness probes? That needs to be precisely stated. Each feature must be testable and generic enough to work on all operators.

Here is a possible path we can take to achieve testing with precise definitions in mind: some behaviors may be triggered by a common API with specially designed endpoints to run those tests. That API could be part of operator-sdk scaffolds, could be a plugin that includes new code on the operator projects or even be just a blueprint on how to implement a certain endpoint. With that not only the metrics endpoint would be available but we would leverage liveness and readiness probes for all workloads and also some other endpoints may be used for triggering expected behaviors. Yet others for collecting results. All of them should be predictable in their possible responses.

With that documented it should become easier for all partners to understand how to expose a certain feature for testing in order to gain a level badge and on the other hand we may see an easier path or an opportunity to add new scaffolded code or feature to operator framework that can accomplish the same goal in the future enhancing even more the user experience.

### Exclude CSV's Capability Field

Currently we have a field called **capability** in CSV files that allows operator developers to declare the capability level of their operators. With this work we don't want that to be true anymore. What we propose is to develop an automated framework capable of testing operator and operand behaviors that align well with cloud native applications. In other words work hand in hand with clouds and hybrid-cloud environments in a fully automated way.

That field would be replaced by a badge in a piece of metadata that is not accessible by developers and/or partners anymore. It could be integrated with the Operator Hub UIs pretty much like they are today but in the form of a list not a single label - with a few exceptions that will also enter the discussion later on each respective section. Those are the exclusion of level 1 and maybe level 2 as a badge. Those 2 levels are so fundamental to what an operator should be that we might consider making those levels prerequisite for listing in the OpenShift certified, Red Hat and Marketplace operators. We discuss those separately below.


## Level 1 - Basic Install

![](https://i.imgur.com/sQj5S8X.png)

### CR Only Interaction to Run Operands

Taking just a piece of the [cloud native definition],(https://github.com/cncf/toc/blob/main/DEFINITION.md) here is what we get:

_"Cloud native technologies empower organizations to build and run scalable applications in modern, dynamic environments such as public, private, and hybrid clouds. Containers, service meshes, microservices, immutable infrastructure, and declarative APIs exemplify this approach._

_These techniques enable loosely coupled systems that are resilient, manageable, and observable. Combined with robust automation, they allow engineers to make high-impact changes frequently and predictably with minimal toil …"_

[https://github.com/cncf/toc/blob/main/DEFINITION.md](https://github.com/cncf/toc/blob/main/DEFINITION.md)

From Operator Framework Level 1 (Basic Install capability definition):

_"...Avoid the practice of requiring the user to create/manage configuration files outside of Kubernetes ..."_

Those two excerpts combined give us what an install should look like - an experience with very little user interaction. This means that the user should interact only with the CR and not have to create any resources beforehand. Which leads me to the first requirement for Basic Install:

**Requirement: Users should be able to install in full any operand controlled by an operator just by creating and configuring a CR. No external resources should be created by users.**

Note on possible future implementation: all owned/secondary resources in-cluster and outside cluster should be verifiable and, even better, authenticated, through a standard API. Their health or completion status should be exposed. A complete status must include all resources in a healthy or complete state. Check health conditions below.


### ALM examples must work

The minimal CR is there for the user to test and try the applications. That is also the first step to test operator quality regarding the capabilities included with it. But we can't rely only on those. 

The advanced features should also figure in other examples and be included in the package in such a way that we can run multiple configurations for the same operand and test them. This may be far in the future but is a nice goal to pursue in order to push quality further.

**Requirement: ALM examples must work out of the box without human intervention. This builds on our first requirement of CR-only interaction.**


### Health Conditions

Question from Operator's Framework capability definitions:

_"...Operator waits for managed resources to reach a healthy state? …"_

"healthy state" is quite vague and could mean different things for different operators. We need a clear and standard way for the operator to declare that the application/operand is in a healthy state regardless of the meaning of that healthy state for the operator.

Another take on it may be finding common healthy state requirements that may be shared among all operands that should be used as well. We need a healthy state definition that is applicable in this case.

I believe for some individual pieces Kubernetes standardized the health state. So workloads should have liveness and readiness probes for example. And those may be a requirement if we think that makes sense. Other operand components may also have predefined health checks. So like RHEL certifies hardware and system integrations the opcap tool could in theory use the same philosophy to evaluate the corner cases and curate a list of validated capabilities for special operators.

### Operand Readiness and Status

From Operator Framwork's capability definition:

_"...Operator conveys readiness of application or managed resources to the user leveraging the status block of the Custom Resource?..."_

Here we are going beyond asking if there is a status field. We want to be prescriptive at least on a piece of the status field to make sure necessary information is present and can be read by automated tooling. 

From the healthy state definition and standardization we can also standardize a common field in status that would represent a minimal green light for the operand. We just need to define what it is and how it looks like and how it should be implemented.

**Requirement: liveness and readiness probes MUST be implemented for all runnable workloads and external tooling should be able to check those probes. All other resources that don't result in a running process should have a status or state condition declared on the CR's status field from the start.** 

We may think in the future deeper ways to check health conditions on all of those. 

Notes on implementation: we may have to define a health condition for most known Kubernetes objects that come as secondary resources. And on external resources a database should be built for every single operator that is certified with the type, probes and checks and what kind of credentials are needed to perform those. Example: SRIOV operator controls SRIOV capable NIC cards. Those cards can be checked on the system via other means than Kubernetes API. Cloud provider resources can also be listed and all of them provide specific APIs that can query their resources. If we're willing to build a fully featured quality validation environment most likely we'll need bare metal test environments with special hardware, cloud provider accounts and credentials for a complete end-to-end test pipeline and validation.


### OLM should be the installer

If the operator installs itself, it shouldn't even be considered part of a level or a feature of some sort. If OLM can't install an operator this operator is not there. It's not present. Therefore it can't be evaluated or tested. Or worse, used. This particular test should not be part of a level 1 validation. It's the bare minimum. Even installing the operand may be considered part of the bare minimum. Check the section "Shouldn't be a badge" below.

**Requirement: all operators should have a bundle and be packaged to install with OLM. We're not considering stand alone operator testing via yaml manifests. Those aren't built to be in the operator hubs. Even before being published they can have custom catalogs and be tested as well by using their bundles.**


### Shouldn't be a Badge

As suggested above, level 1 shouldn't be a badge. Manual tasks to be performed beforehand defeats the operator's purpose. It still needs to be checked for quality but shouldn't grant any special badge to an operator. Since our overall goal is to push operators towards levels 3 to 5, improving the quality of the whole ecosystem, both levels 1 and 2 shouldn't have their own badges. In this way we wouldn't "reward" a partner or developer for doing something basic and essential.


## Level 2 - Seamless Upgrades
![](https://i.imgur.com/tNd7XRQ.png)

### Operand Upgrade Strategies:

From Operator Framework's definitions:

1. _"...Operand can be upgraded in the process of upgrading the Operator, or…"_

This seems to be 1:1. Operator version and full application/operand package have version bumps at the same time.

2. _"...Operand can be upgraded as part of changing the CR…"_

The CR must have a version field that represents a full deployment of the operand no matter how many workloads it manages. (we may need to standardize this field) That is more complex than it seems. The operand may contain multiple workloads like deployments, daemonsets, statefulsets, secrets, configmaps etc. that need to receive an overall version tag that packages all of them into a single distribution pack. That is pretty much what a bundle with CSV does. The thing is that here we're talking about the operand alone here and not including the operator.

3. _"...Operator understands how to upgrade older versions of the Operand, managed previously by an older version of the Operator…"_

Upgrading older versions would happen or by matching the operator version with the operand "package" version or by changing the operand's version in the CR.

4. _"...Operator conveys inability to manage an unsupported version of the Operand in the status section of the CR…"_

This seems to push for another mandatory subfield on status where invalid operand versions should have a standard "not supported by this operator version message".

**Requirement: one of two upgrade methods must be implemented. 1:1 operator/operand or CR with an open version field for the operand. The operand in this case is the whole set of secondary resources and this operand version should represent that and not individual component versions.**

**Requirement:  With that, possibly a field on the CSV could hold the list of operand versions currently supported by that specific operator. This list should also be available through the status field of any CR and proper error messages both on logs and events should be triggered to inform the user a certain version is not supported.**

Note on implementation: regarding those upgrade strategies automation would either run through operator versions in a 1:1 fashion or install an operator version and then create one CR for each supported version and check for the install. At least one test with an unsupported version should also be performed to check messages, events and behavior.


### Operator Upgrade Strategies:

_"...Operator can be upgraded seamlessly and can either still manage older versions of the Operand or update them…"_

Upgrading an operator is the same as installing it with the difference that now it has an older instance already installed and running. Via OLM/subscriptions can be done manually or automatically and must not cause any disruption to the operand.

So what is the definition of seamlessly that we're looking for here? No operand disruption, meaning the health state stays the same to start with.

**Requirement: no changes to operand health state or condition after or during operator upgrade.**


### Shouldn't be a Badge:

On the same basis for the level 1 basic install not being a badge of some sort maybe level 2 shouldn't also be a badge. Those are pretty fundamental automation elements for a controller to go through. Upgrading software is fundamental to its nature. It will have new versions if it's an active project. Part of the seamless feature on the level two may be seen as part of level 3 features. Those look like resiliency and high availability. Some of the considerations below can be moved from this level to a level 3 feature.

Considerations on level 2 that can be part of level 3 features:

During operator/operand upgrades:



* How about user connections not being dropped?
* Network flows not getting disrupted in the data plane?
* Load balancers selectively sending traffic to chosen endpoints for the operator to accomplish upgrades?
* Rolling upgrades and canary deployment patterns: should they be considered? How can we capture that behavior?
* How about standardizing the roll back feature for failed upgrades?


## Level 3 - Full Lifecycle
![](https://i.imgur.com/Y4wYI3B.png)



Proposal: change "Full Lifecycle" for something related to business continuity, disaster recovery and operational resiliency. Full Lifecycle doesn't immediately tell what it is.

_ This includes liveness and readiness probes, multiple replicas, rolling deployment strategies, pod disruption budgets, CPU and memory requests and limits._

_Operator provides the ability to create backups of the Operand_

_Operator is able to restore a backup of an Operand_

The two above need to have a standard recipe to do backup/restore tests.

_Operator orchestrates complex re-configuration flows on the Operand_

I'm not sure this should belong to a level criteria. But there is no border limit between complex and simple here. It seems that any operand reconfiguration would be valid for this case as long as it's related to what we understand as a level 3 capability.

_Operator implements fail-over and fail-back of clustered Operands_

_Operator supports add/removing members to a clustered Operand_

Those two options above belong to a specific category of operand: the clustered ones. If the operand is clustered we need to ask if they have those features. If yes, both should provide a standard way for testing. That would include failover and fail backs with health checks and also including and removing members without disrupting operands operation.

_Operator enables application-aware scaling of the Operand_

"Application-aware" must be well defined. Is it telemetry based auto scaling, will there be fixed thresholds? Will the operator learn baselines?  

Feature list:

Backups

Restores

Reconfiguration coordination

Clustered operand awareness (quorum, failover, fail back, add, remove clustered members)

Liveness and Readiness probes on the operand (with well known fail causes declared and listed)

Rolling deployment strategy

PodDisruptionBudgets created by the operator for the operand

Operand's CPU requests and limits and possibly memory too


## Level 4 - Deep Insights
![](https://i.imgur.com/dyBMrvl.png)



_Health metrics endpoint_

_Operator watching operand for creating and exposing alerts_

_SOPs (Standard Operating Procedures) or runbooks for each alert_

_Critical Alerts created for service down and warning for the others_

_Custom kubernetes events_

_Application Performance metrics_

_RED method applied_


### [Abnormality Detection](https://sdk.operatorframework.io/docs/overview/operator-capabilities/#abnormality-detection)

_Operator determines deviations from a standard performance profile_

This item actually is part of Level 5 in operator framework's documentation. But in my humble opinion, this is actually part of Level 4. Detecting abnormalities is a deep insight. Fixing it automatically is auto-pilot.

I would advocate for a learning operator that can actually state the baseline behavior for all selected metrics. So defining what a baseline is and what statistical method to be used is the first step. With that we can determine profiles. Once those are in place and implemented they can be checked against with current values and determine deviations. Even what is considered a deviation needs discussion.

From the testing perspective the key is: how will those performance profiles/baselines and deviation be exposed by the operator?


## Level 5 - Autopilot
![](https://i.imgur.com/r7gBZdc.png)

_The Operator should understand the application-level performance indicators and determine when it’s healthy and performing well. _

How? We need to provide precise step by step on how to achieve this. That will impact all the other levels that also make use of that.


### [Auto-scaling](https://sdk.operatorframework.io/docs/overview/operator-capabilities/#auto-scaling)

_Operator scales the Operand up under increased load based on Operand metric_

_Operator scales the Operand down below a certain load based on Operand metric_

Testing the feature requires an "overload recipe" that can be run independently. This would need to be standardized as well.


### [Auto-Healing](https://sdk.operatorframework.io/docs/overview/operator-capabilities/#auto-healing)

_Operator can automatically heal unhealthy Operands based on Operand metrics/alerts/logs_

_Operator can prevent the Operand from transitioning into an unhealthy state based on Operand metrics_

How can we make this more precise? It seems that some mandatory metrics must be implemented in order to have this feature. Which ones? This will inform heavily the work on level 4.


### [Auto-tuning](https://sdk.operatorframework.io/docs/overview/operator-capabilities/#auto-tuning)

_Operator is able to automatically tune the Operand to a certain workload pattern_

What is this "certain workload pattern"? How to read it? And how to know the operator has tuned it? It seems to be like a map element with:

	Metric set with thresholds: (metric, threshold, value) - Metrics or whatever workload pattern means needs to be exposed in a precise and deterministic way.

	Related configurations: (var:value, var: value etc.) -  If it's part of the level 5 feature the configuration must be exposed for testing.

_Operator dynamically shifts workloads onto best suited nodes_

What are "best suited nodes"? Wouldn't the Kubernetes scheduler do this job? Are we talking about telemetry aware scheduling? Isn't that already a KEP?
