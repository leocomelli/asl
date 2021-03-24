package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	logger "github.com/rs/zerolog/log"
	"gopkg.in/ini.v1"
)

const (
	awsSSOPath            = "sso"
	keyRegion             = "region"
	keySSOUrl             = "sso_start_url"
	keySSORegion          = "sso_region"
	keySSOAccountID       = "sso_account_id"
	keySSORoleName        = "sso_role_name"
	keyCrdAccessKeyID     = "aws_access_key_id"
	keyCrdSecretAccessKey = "aws_secret_access_key"
	keyCrdSessionToken    = "aws_session_token"
)

var awsPath string

// SSOCredential defines the structure returned by AWS Cli
type SSOCredential struct {
	URL          string `json:"startUrl"`
	Region       string `json:"region"`
	AccessToken  string `json:"accessToken"`
	ExpiresAsStr string `json:"expiresAt"`
}

// SSO implements the flow to retrieve the AWS SSO credentials
type SSO struct {
	Cmd        SSOCommand `json:"-"`
	AccountID  string     `json:"accountId"`
	RoleName   string     `json:"roleName"`
	StartURL   string     `json:"startUrl"`
	Region     string     `json:"region"`
	BackupFile bool       `json:"-"`
}

// Accounts defines the structure returned by AWS Cli
type Accounts struct {
	Items []*Account `json:"accountList"`
}

// Account defines the structure returned by AWS Cli
type Account struct {
	ID    string   `json:"accountId"`
	Name  string   `json:"accountName"`
	Email string   `json:"emailAddress"`
	Roles []string `json:"-"`
}

// AccountRoles defines the structure returned by AWS Cli
type AccountRoles struct {
	Items []*AccountRole `json:"roleList"`
}

// AccountRole defines the structure returned by AWS Cli
type AccountRole struct {
	RoleName  string `json:"roleName"`
	AccountID string `json:"accountId"`
}

// Credentials defines the structure returned by AWS Cli
type Credentials struct {
	Item *Credential `json:"roleCredentials"`
}

// Credential defines the structure returned by AWS Cli
type Credential struct {
	ProfileName     string `json:"-"`
	AccountName     string `json:"-"`
	Region          string `json:"-"`
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	Expiration      int64  `json:"expiration"`
}

// CredentialResultInfo defines the information about SSO credentials
type CredentialResultInfo struct {
	Filename  string
	ExpiresAt time.Time
}

// NewSSO returns a new SSO
func NewSSO(cmd SSOCommand, c *ConfigOptions) *SSO {
	return &SSO{
		cmd,
		c.AccountID,
		c.RoleName,
		c.StartURL,
		c.Region,
		c.BackupFile,
	}
}

// List returns a slice of account roles
func (r *AccountRoles) List() []string {
	var list []string
	for _, r := range r.Items {
		list = append(list, r.RoleName)
	}
	return list
}

// ExpiresAt parses the expiration date
// There is a workaround because the aws cli stores the expiration date
// in different formats.
func (c *SSOCredential) ExpiresAt() time.Time {
	// aws-cli/2.1.29 Python/3.8.8 Darwin/20.3.0 exe/x86_64 prompt/off
	layout := "2006-01-02T15:04:05Z"
	t, err := time.Parse(layout, c.ExpiresAsStr)
	if err != nil {
		//aws-cli/2.0.40 Python/3.8.5 Darwin/19.6.0 source/x86_64
		layout = "2006-01-02T15:04:05UTC"
		t, err = time.Parse(layout, c.ExpiresAsStr)
		if err != nil {
			logger.Warn().Str("value", c.ExpiresAsStr).Msg("error parsing date")
			return time.Now().UTC().AddDate(0, 0, -1)
		}
	}

	return t
}

// Expired returns if credentials are expired
func (c *SSOCredential) Expired() bool {
	return time.Now().UTC().After(c.ExpiresAt())
}

// ExpiresAt returns the expiratin date
func (d *Credential) ExpiresAt() time.Time {
	return time.Unix(d.Expiration/1000, 0)
}

// PersistConfig writes the sso config file
func (a *SSO) PersistConfig() error {
	config := NewFile(awsPath, "config")

	logger.Debug().Str("path", config.FullName).Msg("preparing to store the aws sso config file")

	if a.BackupFile {
		filename, err := config.Backup()
		if err != nil {
			return err
		}

		logger.Info().Str("path", filename).Msg("backup completed successfully")
	}

	cfg, _ := ini.LooseLoad(config.FullName)

	s := cfg.Section(fmt.Sprintf("profile %s", a.RoleName))
	s.Key("output").SetValue("json")
	s.Key(keyRegion).SetValue(a.Region)
	s.Key(keySSOUrl).SetValue(a.StartURL)
	s.Key(keySSORegion).SetValue(a.Region)
	s.Key(keySSOAccountID).SetValue(a.AccountID)
	s.Key(keySSORoleName).SetValue(a.RoleName)

	if err := cfg.SaveTo(config.FullName); err != nil {
		return err
	}

	logger.Info().Str("path", config.FullName).Msg("the aws sso config file has been successfully stored")

	return nil
}

// Login checks if the sso cache file is valid,
// when cache credential has expired forces a login
func (a *SSO) Login(retry ...bool) (*SSOCredential, error) {
	c, err := a.ReadCacheFile()

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if c != nil && !c.Expired() {
		logger.Debug().Bool("expired", c.Expired()).Time("expiresAt", c.ExpiresAt()).Msg("the sso cache file token is not expired")
		return c, nil
	}

	if len(retry) > 0 && retry[0] {
		return nil, errors.New("can not renew the sso token")
	}

	_, err = a.Cmd.Login(a.RoleName)
	if err != nil {
		return nil, err
	}

	logger.Info().Msg("the aws sso cache file has been updated successfully")

	return a.Login(true)
}

// ListAccounts lists accounts assigned to the user
func (a *SSO) ListAccounts(c *SSOCredential) ([]*Account, error) {

	out, err := a.Cmd.ListAccounts(c.AccessToken, c.Region)
	if err != nil {
		return nil, err
	}

	accounts := &Accounts{}
	err = json.Unmarshal([]byte(out), accounts)
	if err != nil {
		return nil, err
	}

	logger.Debug().Interface("accounts", accounts).Msg("accounts obtained with credentials")

	for _, account := range accounts.Items {
		out, err := a.Cmd.ListAccountRoles(c.AccessToken, c.Region, account.ID)
		if err != nil {
			return nil, err
		}

		roles := &AccountRoles{}
		err = json.Unmarshal([]byte(out), roles)
		if err != nil {
			return nil, err
		}

		logger.Debug().Interface("roles", roles).Str("accountID", account.ID).Msg("roles by account")

		account.Roles = roles.List()
	}

	return accounts.Items, nil
}

// GetCredentials retrieves the credentials of the account assigned to the user
func (a *SSO) GetCredentials(c *SSOCredential, accounts []*Account) ([]*Credential, error) {
	var creds []*Credential
	for _, acc := range accounts {
		for i, r := range acc.Roles {
			out, err := a.Cmd.GetRoleCredentials(c.AccessToken, c.Region, acc.ID, r)
			if err != nil {
				return nil, err
			}

			cs := &Credentials{}
			if err := json.Unmarshal([]byte(out), cs); err != nil {
				return nil, err
			}

			logger.Debug().Interface("credentials", cs).Str("accountID", acc.ID).Str("role", r).Msg("credentials...")

			profile := strings.ReplaceAll(acc.Name, " ", "-")
			if i > 0 {
				profile = fmt.Sprintf("%s-%s", profile, Snake(r))
			}

			cs.Item.AccountName = acc.Name
			cs.Item.Region = c.Region
			cs.Item.ProfileName = strings.ToLower(profile)

			logger.Info().Str("account", acc.Name).Str("region", c.Region).Msgf("credentials profile %s", cs.Item.ProfileName)

			creds = append(creds, cs.Item)
		}
	}

	logger.Debug().Msgf("%d credentials have been generated", len(creds))

	if len(creds) == 0 {
		return nil, errors.New("no credentials were found")
	}

	return creds, nil
}

// PersistCredentials writes the credentials to the AWS file
func (a *SSO) PersistCredentials(creds []*Credential) (*CredentialResultInfo, error) {
	cred := NewFile(awsPath, "credentials")

	if a.BackupFile {
		filename, err := cred.Backup()
		if err != nil {
			return nil, err
		}

		logger.Info().Str("path", filename).Msg("backup completed successfully")
	}

	cfg, _ := ini.LooseLoad(cred.FullName)

	for _, c := range creds {
		s := cfg.Section(c.ProfileName)
		s.Key("output").SetValue("json")
		s.Key(keyRegion).SetValue(c.Region)
		s.Key(keyCrdAccessKeyID).SetValue(c.AccessKeyID)
		s.Key(keyCrdSecretAccessKey).SetValue(c.SecretAccessKey)
		s.Key(keyCrdSessionToken).SetValue(c.SessionToken)
	}

	if err := cfg.SaveTo(cred.FullName); err != nil {
		return nil, err
	}

	return &CredentialResultInfo{
		Filename:  cred.FullName,
		ExpiresAt: creds[len(creds)-1].ExpiresAt(),
	}, nil
}

// ReadCacheFile reads the sso cache file for a given sso
func (a *SSO) ReadCacheFile() (*SSOCredential, error) {
	hash := sha1.New()
	_, err := hash.Write([]byte(a.StartURL))
	if err != nil {
		return nil, err
	}

	cacheFilename := strings.ToLower(hex.EncodeToString(hash.Sum(nil))) + ".json"
	cache := NewFile(awsPath, awsSSOPath, "cache", cacheFilename)

	logger.Debug().Str("path", cache.FullName).Msg("searching for the aws sso cache file...")

	if !cache.Exists() {
		logger.Debug().Str("path", cache.FullName).Msg("aws sso cache file not found")
		return nil, os.ErrNotExist
	}

	data := &SSOCredential{}
	if err := cache.ReadJSON(data); err != nil {
		return nil, err
	}

	logger.Debug().Interface("data", data).Msg("cache file was read successfully")

	return data, err
}

func init() {
	home, err := homedir.Dir()
	if err != nil {
		logger.Fatal().Err(err)
	}

	awsPath = filepath.Join(home, ".aws")
}
