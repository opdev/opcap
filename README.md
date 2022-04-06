# Capabilities-tool

**Capabilities-tool** is a commandline interface for checking if
[OpenShift](https://www.openshift.com/) operator meets minimum
requirements for [Operator Capabilities Level](https://sdk.operatorframework.io/docs/overview/operator-capabilities/).

This project is in active and rapid development! The current implementation check for level 1 capabilities only.

## Prerequisites

- Create a secret with json credentials file to authenticate against registry.redhat.io and registry.connect.redhat.com
- Create a secret with kubeconfig for the cluster

## Deployment

To deploy this project use the provided capabilities-tool-job.yaml manifest. Also create a config map `audit-env-var` with below mention environment variables.

## Environment Variables

The capabilities-tool binary currently requires that you have the following variables set in the environment

| Name                    |                    Value                     | 
|-------------------------|:--------------------------------------------:|
| MINIO_ENDPOINT          |      any non-SSL S3 compatible endpoint      |
| MINIO_ACCESS_KEY        | the user/username to login to the S3 backend |
| MINIO_SECRET_ACCESS_KEY |   the password to login to the S3 backend    |

## Usage

The `capabilities-tool` utility allows one to confirm that Operator projects in an index image 
comply with Operator capabilities levels.

A brief summary of the available sub-commands is as follows:

```text
A utility that allows you to pre-test your operator bundles before submitting for Red Hat Certification.

Usage:
  capabilties-tool [command]

Available Commands:
  index capabilities          Run checks for an operator bundle capability levels

Flags:
  --container-engine    specifies the container tool to use
  --output-path         path where results will be stored
  --bundle-image        image path of the bundle to be tested
  --pull-secret-name    name of Kubernetes Secret to use for pulling registry images
  --bucket-name         minio bucket name where result will be stored
  --bundle-name         name of the bundle to be tested
  --service-account     name of Kubernetes Service Account to use
  -h,--help             help for capabilities-tool
  -v,--version          version for capabilities-tool
Use "capabilities-tool [command] --help" for more information about a command.
```

To check an Operator bundle, utilize the `index capabilities` sub-command:

```text
Example:
capabilities-tool index capabilities --bucket-name=audit-tool-s3-bucket --bundle-image=registry.connect.redhat.com/minio/minio-operator1@sha256:357a5089b211e9653efff6cacc433633993cf3317dcca29eb54f924374b47b88 --bundle-name=minio-operator.v4.0.2
```

Note: this project is not intended to run by itself. To run the capabilities-tool CLI, use [Audit-tool-orchestrator](https://github.com/mrhillsman/audit-tool-orchestrator)