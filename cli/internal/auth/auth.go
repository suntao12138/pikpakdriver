package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	DefaultConfigDir  = ".config/pikpakdriver"
	DefaultConfigFile = "config.json"
)

type Credential struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func CredentialsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, DefaultConfigDir, DefaultConfigFile)
}

func SessionPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, DefaultConfigDir, "session.json")
}

func LoadCredentials() (*Credential, error) {
	v := viper.New()
	v.SetConfigFile(CredentialsPath())
	v.SetConfigType("json")
	if err := v.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no credential found. Run 'pikpakdriver login' to authenticate")
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cred Credential
	if err := v.Unmarshal(&cred); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	if cred.Email == "" || cred.Password == "" {
		return nil, fmt.Errorf("incomplete credentials in config")
	}
	return &cred, nil
}
