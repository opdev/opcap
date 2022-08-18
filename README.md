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

## Prerequisites

- Provide opcap tool with kubeconfig for the cluster

## Environment Variables

The opcap binary currently requires that you have the following variables set in the environment

| Name                 |                    Value                      | 
|----------------------|:---------------------------------------------:|
| KUBECONFIG		       |	kubeconfig for the cluster to run opcap tool |
| S3_ENDPOINT          |  any non-SSL S3 compatible endpoint           |
| S3_ACCESS_KEY        |  the user/username to login to the S3 backend |
| S3_SECRET_ACCESS_KEY |  the password to login to the S3 backend      |

## Usage

The `opcap` utility allows users to confirm that Operator projects in a provided catalogSource
comply with Operator capabilities levels.

A brief summary of the available sub-commands is as follows:

```text
A utility that allows users to check if the operator satisfies the capabilities levels of an operator.
```

Usage:
  opcap [command]

Available Commands:
  check        Run checks for an operator bundle capability levels
  upload       It parses the stdout.json created by check command and creates a report.json. It uploads the report.json to Minio/S3 bucket provided

To check an Operator capabilites, utilize the `check` sub-command:

```text
Example:
opcap check --catalogsource=certified-operators --catalogsourcenamespace=openshift-marketplace
```


```
Flags:
  --catalogsource                 specifies the catalogsource to test against
  --catalogsourcenamespace        specifies the namespace where the catalogsource exists
```

To upload the results of the Operator capabilities from 'check' command, use the 'upload' command and provide it with the following flags or set them as environment variable as described as specified above:

```
Flags:
  --bucket          s3 bucket where result will be stored(will create a bucket if not provided)
  --path            s3 path where result will be stored
  --endpoint        s3 endpoint where bucket will be created
  --accesskeyid     s3 access key id for authentication
  --secretaccesskey s3 secret access key for authentication
  --usessl          when used s3 backend is expected to be accessible via https; false by default
  --trace           enable tracing; false by default
```

```text
Example:
opcap upload --bucket=xxxxxx --path=xxxxxx --endpoint=xxxxxx --accesskeyid=xxxxxx --secretaccesskey=xxxxxx
```