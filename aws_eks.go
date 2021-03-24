package main

import (
	"encoding/json"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	logger "github.com/rs/zerolog/log"
)

var kubeConfig string

// EKSClusters defines the structure returned by AWS Cli
type EKSClusters struct {
	Items []string `json:"clusters"`
}

// EKS implements the flow to retrieve the EKS clusters configuration to use them with kubectl
type EKS struct {
	Cmd            EKSCommand
	KubeConfigPath string
	BackupFile     bool
}

// NewEKS returns a new EKS
func NewEKS(cmd EKSCommand, c *ConfigOptions) *EKS {
	return &EKS{
		Cmd:            cmd,
		KubeConfigPath: kubeConfig,
		BackupFile:     c.BackupFile,
	}
}

// UpdateKubeConfig constructs a configuration with prepopulated server and certificate
// authority data values for each credential retrieved.
func (e *EKS) UpdateKubeConfig(creds []*Credential) error {
	if e.BackupFile {
		kubeConfigFile := NewFile(e.KubeConfigPath)
		filename, err := kubeConfigFile.Backup()
		if err != nil {
			return err
		}

		logger.Info().Str("path", filename).Msg("backup completed successfully")
	}

	for _, cred := range creds {
		out, err := e.Cmd.ListClusters(cred.Region, cred.ProfileName)
		if err != nil {
			return err
		}

		logger.Debug().Str("profile", cred.ProfileName).Str("region", cred.Region).Msg("listing eks clusters...")

		clusters := &EKSClusters{}
		err = json.Unmarshal([]byte(out), clusters)
		if err != nil {
			return err
		}

		logger.Debug().Msgf("%d clusters were found", len(clusters.Items))

		for _, c := range clusters.Items {
			out, err := e.Cmd.UpdateKubeConfig(cred.Region, cred.ProfileName, c)
			if err != nil {
				return err
			}

			logger.Info().Str("cluster", c).Str("profile", cred.ProfileName).Msg("kubeconfig successfully updated")

			logger.Trace().Str("cluster", c).Msg(out)
		}
	}

	return nil
}

func init() {
	home, err := homedir.Dir()
	if err != nil {
		logger.Fatal().Err(err)
	}

	kubeConfig = filepath.Join(home, ".kube", "config")
}
