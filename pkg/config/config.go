package config

import (
	"errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

const (
	FlagAccessKey      = "accessKey"
	FlagSecretKey      = "secretKey"
	FlagNamesrvAddrKey = "namesrvAddrs"
)

type RocketMQConfig struct {
	ConfigFile   string
	AccessKey    string `yaml:"AccessKey"`
	SecretKey    string `yaml:"SecretKey"`
	NamesrvAddrs string `yaml:"NamesrvAddrs"`
}

func (r *RocketMQConfig) Load() error {
	if r.ConfigFile == "" {
		// default config file
		configDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		r.ConfigFile = filepath.Join(configDir, "/.config", "rmq", "rmq.yaml")
	}
	data, err := os.ReadFile(r.ConfigFile)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return errors.New("empty config file")
	}
	err = yaml.Unmarshal(data, r)

	if err != nil {
		return err
	}
	return nil
}

func (r *RocketMQConfig) InstallRocketMQFlags(cmd *cobra.Command) {
	pf := cmd.PersistentFlags()
	pf.StringVar(&r.NamesrvAddrs, FlagNamesrvAddrKey, "127.0.0.1:9876", "Comma separated list of namesrv host:ports")
	pf.StringVar(&r.AccessKey, FlagAccessKey, "", "Access key")
	pf.StringVar(&r.SecretKey, FlagSecretKey, "", "Access key")

	_ = pf.MarkHidden(FlagNamesrvAddrKey)
	_ = pf.MarkHidden(FlagAccessKey)
	_ = pf.MarkHidden(FlagSecretKey)
}

func (r *RocketMQConfig) GetNamesrvAddrs() []string {
	return strings.Split(r.NamesrvAddrs, ",")
}
