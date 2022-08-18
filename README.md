# OpenShift Operator Capabilities - opcap tool

opcap is a cli tool that interacts with all operators available on OpenShift's operator hub in a live cluster to assess how smart and cloud native they are. 

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