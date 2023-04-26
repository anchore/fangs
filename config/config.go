package config

import (
	"github.com/spf13/pflag"

	"github.com/anchore/fangs/config/log"
)

type Config struct {
	Log        log.Log
	AppName    string `json:"-" yaml:"-" mapstructure:"-"`
	ConfigFile string `json:"config,omitempty" yaml:"config,omitempty" mapstructure:"-"`
}

func NewConfig(appName string) Config {
	return Config{
		Log:     log.NewDiscard(),
		AppName: appName,
	}
}

func (r *Config) AddFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&r.ConfigFile, "config", "c", r.ConfigFile, "configuration file")
}
