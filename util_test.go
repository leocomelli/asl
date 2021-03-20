package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmptySnake(t *testing.T) {
	r := Snake("")
	require.Equal(t, "", r)
}

func TestRoleNameSnake(t *testing.T) {
	r := Snake("DevSSOLogin")
	require.Equal(t, "Dev-SSO-Login", r)
}
