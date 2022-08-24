# OpenShift Operator Capabilities - opcap tool

opcap is a cli tool that interacts with all operators available on OpenShift's operator hub in a live cluster to assess how smart and cloud native they are. 

## The Problem opcap Solves

The [operator framework](https://operatorframework.io/) has defined 5 different [maturity levels](https://sdk.operatorframework.io/docs/overview/operator-capabilities/) that range from a basic install to auto pilot. The [Operator Lifecycle Manager](https://olm.operatorframework.io/), which is part of the operator framework, provides a way for operator developers to declare on which of those 5 levels a particular operator is classified.

As the market understands how essential a robust and resilient automation stack is to an application lifecycle, operators became more and more popular. Now we have hundreds of them being developed out there to accomplish tasks once done by human hands in a movement to win the hybrid cloud game.

How can we assure that those operators being published to the operator hub actually have those features listed in the maturity levels? All sorts of issues have been seen with the maturity level model that we have now. From operators that require a lot of manual work from the user in order to be installed to operators that claim auto-pilot level without minimum requirements even for a basic install. At the same time maturity level understanding varies among developers since what we have are guidelines but not precise and testable atomic definitions for those maturity features.

On another note operators are constantly evolving. It's impossible to guarantee by hand that every new version will keep up with the same features they started with.

Hence, opcap's goal is to provide essential tooling for the operator hub to properly classify the operator's maturity without any human interaction, with precise methods and bringing, when applicable, new maturity features to improve the ecosystem.

## History Summary and Roadmap

It started as a tool, used by the Operator Enablement and Tooling team at Red Hat, to identify problems installing operators and their respective applications with them. Quickly evolved into a bigger discussion on the whole operator maturity model and an attempt to automate complex operator behavior testing.

The tool is currently under development but already tests operator installation as well as all the APIs, AKA CRDs, provided with them on initial deployment.

The roadmap includes climbing the maturity level testing up to auto-pilot without limiting the capabilities to this model in order to allow flexible badges that represent more granular features or features that may be important and are not included in the current model like security for example. Check out our general [operator maturity design document](/docs/maturity.md) for more detailed concepts.

## Who Is It For?

opcap is supposed to be used alongside OpenShift clusters and query catalog sources to run operators from published packages. That primarily puts opcap on the tooling shelf for quality engineers and operator enablement folks to check for operator capabilities and quickly identify problems that need to be addressed by reaching out to operator developers.

Another potential big user are operator developers that want to make sure, in development time, that their next version will keep all the badges they already have and/or include a new one related to new features being included.

## Architectural Design and Operator Maturity Concepts

- For a more in depth view on how this tool works check our [Architectural Design](/docs/design.md) doc.

- For more in depth and precise operator maturity feature definitions check our [maturity](/docs/maturity.md) doc:

## Prerequisites

- A working OpenShift cluster or Code Ready Containers (CRC) - For a full catalog check it's important to have a multi-node cluster. Some operators/operands don't install in Single Node OpenShift. They were design for multi-node experience.
- KUBECONFIG environment variable set on opcap's terminal with cluster admin access or present on default home path ($HOME/.kube/config)
- If testing against custom catalog sources those must be running before opcap runs

opcap can store json reports on S3 type object storage devices by using the upload command. In order to use this command you may also opt to provide environment variables like the ones below or use flags to configure it.

| Name                 |                    Value                      | 
|----------------------|:---------------------------------------------:|
| S3_ENDPOINT          |  any non-SSL S3 compatible endpoint           |
| S3_ACCESS_KEY        |  the user/username to login to the S3 backend |
| S3_SECRET_ACCESS_KEY |  the password to login to the S3 backend      |

## Usage

### Checking all the operators in a catalog source:

```
opcap check --catalogsource=certified-operators --catalogsourcenamespace=openshift-marketplace
```

That will print to the screen a report that can be saved by redirection and also creates files where opcap is placed with the available reports. For now it defaults to check if operators only are installing correctly and renders a file that looks like below:

```
{"level":"info","message":"Succeeded","package":"hazelcast-platform-operator","channel":"alpha","installmode":"OwnNamespace"}
{"level":"info","message":"Succeeded","package":"kubeturbo-certified","channel":"stable","installmode":"OwnNamespace"}
{"level":"info","message":"Succeeded","package":"elasticsearch-eck-operator-certified","channel":"stable","installmode":"OwnNamespace"}
{"level":"info","message":"timeout","package":"mongodb-enterprise","channel":"stable","installmode":"OwnNamespace"}
```

In order to test operands, i.e., the CRs or applications with the operators we need to modify the audit plan like below:

```
./bin/opcap check --auditplan=OperatorInstall,OperandInstall,OperandCleanUp,OperatorCleanUp
```

And that's what you should see on the screen. Both operator and operand tested for basic install.

```
I0818 15:24:28.288017  360146 request.go:601] Waited for 1.000589353s due to client-side throttling, not priority and fairness, request: GET:https://api.exedb-sno-pool-7nw9p.opdev.io:6443/apis/template.openshift.io/v1?timeout=32s
check called

Operator Install Report:
-----------------------------------------
Report Date: 2022-08-18 15:25:42.951700038 -0400 EDT m=+76.015550330
OpenShift Version: 4.10.26
Package Name: mongodb-enterprise
Channel: stable
Catalog Source: certified-operators
Install Mode: OwnNamespace
Result: timeout
Message: installing: waiting for deployment mongodb-enterprise-operator to become ready: deployment "mongodb-enterprise-operator" not available: Deployment does not have minimum availability.
Reason: InstallWaiting
-----------------------------------------

Operand Install Report:
-----------------------------------------
Report Date: 2022-08-18 15:25:45.855014223 -0400 EDT m=+78.918864516
OpenShift Version: 4.10.26
Package Name: mongodb-enterprise
Operand Kind: MongoDB
Operand Name: my-replica-set
Operand Creation: Succeeded
-----------------------------------------
```

And another file for operands will be created that will look like below:

```
{"package":"hazelcast-platform-operator", "Operand Kind": "Hazelcast", "Operand Name": "hazelcast","message":"created"}
{"package":"kubeturbo-certified", "Operand Kind": "Kubeturbo", "Operand Name": "kubeturbo-release","message":"created"}
{"package":"elasticsearch-eck-operator-certified", "Operand Kind": "Elasticsearch", "Operand Name": "elasticsearch-sample","message":"created"}
{"package":"mongodb-enterprise", "Operand Kind": "MongoDB", "Operand Name": "my-replica-set","message":"created"}
```

### Upload operator reports to S3 buckets:

```
opcap upload --bucket=xxxxxx --path=xxxxxx --endpoint=xxxxxx --accesskeyid=xxxxxx --secretaccesskey=xxxxxx
```

That command will read the reports on file and upload to the bucket.

### Listing available packages from a catalog source:

```
opcap check --list-packages --catalogsource=certified-operators --catalogsourcenamespace=openshift-marketplace
```

# How to Build and Test opcap

### Requirements

- Go version 1.17+
- Make 4+
- Linux or MacOS

### Simple build for testing

To build opcap nothing more than `go build` is necessary if it's only to run commands and test the tool.

### Building with version

The `make build` command will include the version contained in the OPCAP_VERSION make var with other detailed information about the building environment and user. That information will be generated as compile time variables and will inform the `opcap version` command contained in the binary. That should be align with the tag used to create releases.

With that if you run:

`OPCAP_VERSION=0.0.1 make build`

You should see a binary in the bin folder called opcap. When running `./bin/opcap version` you should see something like below:

```
Version:         0.0.1
Go Version:      go1.18.3
Build Time:      Thu Aug 18 03:15:09 PM EDT 2022
Git User:        Author: acmenezes <adcmenezes@gmail.com>
Git Commit:      5a3a6367518ccc6382ea06aabc4c611be62f3a72
```

### Using VSCode with launch.json

If you want to run opcap with VSCode debugging features here is an example on how the launch.json file can looks like:

```
{
    "version": "0.2.0",
    "configurations": [
      {
        "name": "Launch Package",
        "type": "go",
        "request": "launch",
        "mode": "auto",
        "program": "${fileDirname}",
        "args": ["check"],
        "env": {"KUBECONFIG":"/path/to/my/cluster/kubeconfig"}
      }
    ]
  }

```

If you want to know more about debugging on VSCode check this page https://code.visualstudio.com/docs/editor/debugging

    TIP: if you want to check a single operator or operand you can put a break point at the very beginning of OperatorInstall or OperandInstall functions that will allow you to stop right before creating a new one and also break at OperatorCleanUp and OperandCleanUp that will allow developers to check what's in the cluster before it gets cleaned up.

### Contribution Guidelines

Check our contribution guidelines [here](docs/contribution.md).

### Weekly Meeting and Slack Channel

If you are a member of the CoreOS workspace in Slack you can find us at #forum-opcap.

Our weekly meeting happens on Fridays 10:30-11:30am EST via google meet:

Operator Audit Meeting
Video call link: https://meet.google.com/vix-fpch-nkw
