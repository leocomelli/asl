package main

import "testing"

type SSOMock struct{}

func (c *SSOMock) Login(roleName string) (string, error) {
	return "", nil
}

func (c *SSOMock) ListAccounts(accessToken string, region string) (string, error) {
	return "", nil
}

func (c *SSOMock) ListAccountRoles(accessToken string, region string, accountID string) (string, error) {
	return "", nil
}

func (c *SSOMock) GetRoleCredentials(accessToken string, region string, accountID string, roleName string) (string, error) {
	return "", nil
}

func TestX(t *testing.T) {}
