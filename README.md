# Capabilities-tool

**Capabilities-tool** is a commandline interface for checking if
[OpenShift](https://www.openshift.com/) operator meets minimum
requirements for [Operator Capabilities Level](https://sdk.operatorframework.io/docs/overview/operator-capabilities/).

This project is in active and rapid development! The current implementation check for level 1 capabilities only.

## Requirements

The capabilities-tool binary currently requires that you have the following variables set in the environment

| Name                    |            Value            | 
|-------------------------|:---------------------------:| 
| MINIO_ENDPOINT          |     Endpoint for Minio      |
| MINIO_SECRET_ACCESS_KEY | Secret Access key for Minio |
| MINIO_ACCESS_KEY        |    Access Key for Minio     |

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
  --container-engine    podman/docker
  --output-path         path where results will be stored
  --bundle-image        image path of the bundle to be tested
  --bucket-name         minio bucket name where result will be stored
  --bundle-name         name of the bundle to be tested
  -h,--help             help for capabilities-tool
  -v,--version          version for capabilities-tool
Use "capabilities-tool [command] --help" for more information about a command.
```

To check an Operator bundle, utilize the `index capabilities` sub-command:

```text
Example:
capabilities-tool index capabilities --bucket-name=audit-tool-s3-bucket --bundle-image=registry.connect.redhat.com/minio/minio-operator1@sha256:357a5089b211e9653efff6cacc433633993cf3317dcca29eb54f924374b47b88 --bundle-name=minio-operator.v4.0.2
```