# ASL ::: Amazon Single Sign-On Login

ASL is a cli to get the STS short-term credentials for all accounts and role names that is assigned to the AWS SSO user.

## What does ASL do?

ASL retrieves and caches an AWS SSO access token to exchange for AWS credentials, when the cached access token expires, a new login is requested. Using a valid access token, the ASL lists all AWS accounts assigned to the user and then get the roles for each one. After that, the STS short-term credentials are stored in AWS credential file.

Note: ASL override all content of AWS credential file (`$HOME/.aws/credentials`). If you need to preserve the current contet, use the `--backup` flag to back up.

## Prerequisites

* [AWS Command Line Interface](https://aws.amazon.com/cli/)

## Usage

Run the `asl configure` command to store the AWS SSO Login parameters to be used when needed. Whenever the AWS SSO access token needs to be renewed, these parameters are used.

```sh
asl configure \
  --account-id 123456789012 \
  --start-url https://d-123456w78w.awsapps.com/start/ \
  --role-name MyRoleSSOLogin \
  --region us-east-1
```

Run the `asl` command to store the STS short-term credentials for each account and role assigned to the user. You may safely rerun the `asl` command to refresh your credentials.

```sh
asl
```

Make sure everything works well

```sh
aws sts get-caller-identity --profile your-profile
```

### EKS

Use the flag `--eks` to update the kubeconfig with all existing clusters in the accounts assigned to the user.

```sh
asl --eks
```

