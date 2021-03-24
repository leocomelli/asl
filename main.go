package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	logger "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// Options defines root command options
type Options struct {
	Backup        bool
	EKS           bool
	ForceSSOLogin bool
}

var (
	// Version contains the current version of the app.
	Version = ""
	// BuildDate contains the date and time of build process.
	BuildDate = ""
	// GitHash contains the hash of last commit in the repository.
	GitHash = ""

	opts = &Options{}
)

const (
	msgTmpl = `it worked! \o/

*****************************************************************************************************************
%s
%s
note that it will expire at %s
after this time, you may safely rerun this cli to refresh your credentials
*****************************************************************************************************************
`
	ssoMsgTmpl = `SSO
   your new access key pair has been stored in the aws configuration file %s
   to use these credentials, set the AWS_PROFILE or call the aws cli with the --profile option.
`
	eksMsgTmpl = `EKS
   your kubernetes config has been updated in the kubeconfig file %s
   to use these contexts, run kubectl config set-context <name> or call the kubectl with the --context option
`
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "asl",
		Short: "Get credentials for all accounts for which you have permission in AWS SSO",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := LoadConfig(opts)
			if err != nil {
				return err
			}

			sso := NewSSO(&SSOCli{}, cfg)
			err = sso.PersistConfig()
			if err != nil {
				return err
			}

			ssoCred, err := sso.Login()
			if err != nil {
				return err
			}

			accounts, err := sso.ListAccounts(ssoCred)
			if err != nil {
				return err
			}

			c, err := sso.GetCredentials(ssoCred, accounts)
			if err != nil {
				return err
			}

			res, err := sso.PersistCredentials(c)
			if err != nil {
				return err
			}
			ssoMsg := fmt.Sprintf(ssoMsgTmpl, res.Filename)

			var eksMsg string
			if opts.EKS {
				eks := NewEKS(&EKSCli{}, cfg)
				err := eks.UpdateKubeConfig(c)
				if err != nil {
					return err
				}

				eksMsg = fmt.Sprintf(eksMsgTmpl, eks.KubeConfigPath)
			}

			logger.Info().Msgf(msgTmpl, ssoMsg, eksMsg, res.ExpiresAt)

			return nil
		},
	}

	rootCmd.PersistentFlags().StringP("loglevel", "d", "info", "set log level [info|debug|trace]")
	rootCmd.PersistentFlags().BoolVarP(&opts.Backup, "backup", "b", false, "force a back up of the configuration files [.aws/config|.aws/credentials|.kube/config]")
	rootCmd.PersistentFlags().BoolVarP(&opts.EKS, "eks", "k", false, "configure kubectl so that you can connect to an Amazon EKS cluster")
	rootCmd.PersistentFlags().BoolVarP(&opts.ForceSSOLogin, "login", "l", false, "force login to review the SSO access token")

	logger.Logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	setLogLevel(os.Args)

	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	rootCmd.AddCommand([]*cobra.Command{
		configureCmd(ctx),
		versionCmd(ctx),
	}...)

	err := rootCmd.Execute()
	if err != nil {
		logger.Fatal().Err(err)
	}
}

func setLogLevel(args []string) {
	level := "info"
	for i, a := range args {
		if a == "--loglevel" {
			level = args[i+1]
			break
		}
	}

	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)
}
