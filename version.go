package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

const tmpl = `
Version: %s
BuildDate: %s
GitCommit: %s
`

func versionCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of ASL",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf(tmpl, Version, BuildDate, GitHash)
			return nil
		},
	}

	return cmd
}
