package main

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	logger "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var aslPath string

// ConfigOptions defines the ASL options
type ConfigOptions struct {
	AccountID  string `json:"accountId"`
	RoleName   string `json:"roleName"`
	StartURL   string `json:"startUrl"`
	Region     string `json:"region"`
	BackupFile bool   `json:"-"`
}

func configureCmd(ctx context.Context) *cobra.Command {
	o := &ConfigOptions{}

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Store the parameters used to log in to AWS SSO",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Debug().Str("aslPath", aslPath).Interface("options", o).Msg("configuring...")

			if err := Configure(o); err != nil {
				return err
			}

			logger.Info().Msg("it worked! please run: asl")

			return nil
		},
	}

	cmd.Flags().StringVarP(&o.AccountID, "account-id", "a", "", "the AWS account that is assigned to the user")
	cmd.Flags().StringVarP(&o.RoleName, "role-name", "r", "", "the role name that is assigned to the user")
	cmd.Flags().StringVarP(&o.StartURL, "start-url", "u", "", "the URL that points to the organization's AWS Single Sign-On (AWS SSO) user portal")
	cmd.Flags().StringVarP(&o.Region, "region", "l", "", "the region to use")

	_ = cmd.MarkFlagRequired("account-id")
	_ = cmd.MarkFlagRequired("role-name")
	_ = cmd.MarkFlagRequired("start-url")
	_ = cmd.MarkFlagRequired("region")

	return cmd
}

// Configure writes the ASL parameters to use when needed
func Configure(o *ConfigOptions) error {
	config := NewFile(aslPath)
	if err := config.WriteJSON(o); err != nil {
		return err
	}

	logger.Debug().Str("path", config.FullName).Msg("the asl config file has been successfully stored")

	return nil

}

// LoadConfig reads the ASL parameters
func LoadConfig() (*ConfigOptions, error) {
	config := NewFile(aslPath)

	if !config.Exists() {
		return nil, errors.New("asl config file not found. please run: asl configure")
	}

	logger.Info().Str("path", config.FullName).Msg("loading the asl config file")

	data := &ConfigOptions{}
	err := config.ReadJSON(data)
	if err != nil {
		return nil, err
	}

	logger.Debug().Interface("data", data).Msg("the asl config file has been successfully read")

	return data, nil
}

func init() {
	home, err := homedir.Dir()
	if err != nil {
		logger.Fatal().Err(err)
	}

	aslPath = filepath.Join(home, ".asl")
}
