package config

import (
	"github.com/spf13/pflag"

	"github.com/anchore/go-logger"
	"github.com/anchore/go-logger/adapter/discard"
)

type Config struct {
	Logger     logger.Logger
	AppName    string `json:"-" yaml:"-" mapstructure:"-"`
	ConfigFile string `json:"config,omitempty" yaml:"config,omitempty" mapstructure:"-"`
}

func NewConfig(appName string) Config {
	return Config{
		Logger:  discard.New(),
		AppName: appName,
	}
}

func (r *Config) AddFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&r.ConfigFile, "config", "c", r.ConfigFile, "configuration file")
}
