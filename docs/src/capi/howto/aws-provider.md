# Setting up the AWS Infrastructure provider

The AWS infrastructure provider requires some initial steps to setup.

## Install clusterawsadm

The AWS infrastructure provider requires the `clusterawsadm` tool to be
installed:

```sh
curl -L https://github.com/kubernetes-sigs/cluster-api-provider-aws/releases/download/v2.5.2/clusterawsadm-linux-amd64 -o clusterawsadm
chmod +x clusterawsadm
sudo mv clusterawsadm /usr/local/bin
```

`clusterawsadm` helps you bootstrapping the AWS environment that CAPI will use
and set the necessary permissions.

Start by setting up environment variables defining the AWS account to use, if
these are not already defined:

```sh
export AWS_REGION=<your-region-eg-us-east-1>
export AWS_ACCESS_KEY_ID=<your-access-key>
export AWS_SECRET_ACCESS_KEY=<your-secret-access-key>
```

If you are using multi-factor authentication, you will also need:

```sh
export AWS_SESSION_TOKEN=<session-token>
```

clusterawsadm` uses these details to create a CloudFormation stack in your AWS
account with the correct IAM resources:

```sh
clusterawsadm bootstrap iam create-cloudformation-stack
```

The credentials should also be encoded and stored as a Kubernetes secret:

```sh
export AWS_B64ENCODED_CREDENTIALS=$(clusterawsadm bootstrap credentials encode-as-profile)
```

You are now all set to deploy the AWS CAPI infrastructure provider.
Visit [getting-started] for next steps.

<!-- Links -->
[getting-started]: ../tutorial/getting-started.md
