package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	logger "github.com/rs/zerolog/log"
)

// SSOCommand represents the commands for interacting with AWS SSO
type SSOCommand interface {
	Login(string) (string, error)
	ListAccounts(string, string) (string, error)
	ListAccountRoles(string, string, string) (string, error)
	GetRoleCredentials(string, string, string, string) (string, error)
}

// EKSCommand represents the commands for interacting with EKS
type EKSCommand interface {
	ListClusters(string, string) (string, error)
	UpdateKubeConfig(string, string, string) (string, error)
}

// ----- SSO -----

// SSOCli implements commands to perform SSO actions through AWS Cli
type SSOCli struct{}

// Login retrieves  and  caches an AWS SSO access token to exchange for AWS credentials
func (c *SSOCli) Login(roleName string) (string, error) {
	return execCli("sso", "login", "--profile", roleName)
}

// ListAccounts lists  all  AWS  accounts  assigned to the user
func (c *SSOCli) ListAccounts(accessToken string, region string) (string, error) {
	return execCli("sso", "list-accounts", "--access-token", accessToken, "--region", region)
}

// ListAccountRoles lists  all roles that are assigned to the user for a given AWS account
func (c *SSOCli) ListAccountRoles(accessToken string, region string, accountID string) (string, error) {
	return execCli("sso", "list-account-roles", "--access-token", accessToken, "--region", region, "--account-id", accountID)
}

// GetRoleCredentials returns the STS short-term credentials for a given role name that is assigned to the user
func (c *SSOCli) GetRoleCredentials(accessToken string, region string, accountID string, roleName string) (string, error) {
	return execCli("sso", "get-role-credentials", "--access-token", accessToken, "--region", region, "--account-id", accountID, "--role-name", roleName)
}

// ----- EKS -----

// EKSCli implements commands to perform EKS actions through AWS Cli
type EKSCli struct{}

// ListClusters lists the Amazon EKS clusters in your AWS account in the specified region
func (k *EKSCli) ListClusters(region string, profile string) (string, error) {
	return execCli("eks", "list-clusters", "--region", region, "--profile", profile)
}

// UpdateKubeConfig configures kubectl so that you can connect to an Amazon EKS cluster
func (k *EKSCli) UpdateKubeConfig(region string, profile string, name string) (string, error) {
	return execCli("eks", "update-kubeconfig", "--name", name, "--region", region, "--profile", profile)
}

func execCli(args ...string) (string, error) {
	cmd := exec.Command("aws", args...)
	out, err := cmd.CombinedOutput()
	outStr := strings.ReplaceAll(string(out), "\n", "")

	logger.Trace().Interface("command", args).Msg(strings.ReplaceAll(outStr, "\n", ""))

	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() != 0 {
		return "", fmt.Errorf("[aws cli] %s", outStr)
	}

	return outStr, err
}
