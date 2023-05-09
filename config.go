package fangs

import (
	"github.com/spf13/pflag"

	"github.com/anchore/go-logger"
	"github.com/anchore/go-logger/adapter/discard"
)

type Config struct {
	Logger  logger.Logger `json:"-" yaml:"-" mapstructure:"-"`
	AppName string        `json:"-" yaml:"-" mapstructure:"-"`
	TagName string        `json:"-" yaml:"-" mapstructure:"-"`
	File    string        `json:"config,omitempty" yaml:"config,omitempty" mapstructure:"-"`
	Finders []Finder      `json:"-" yaml:"-" mapstructure:"-"`
}

func NewConfig(appName string) Config {
	return Config{
		Logger:  discard.New(),
		AppName: appName,
		TagName: "yaml",
		// search for configs in specific order
		Finders: []Finder{
			// 1. look for a directly configured file
			FindDirect,
			// 2. look for ./.<appname>.<ext>
			FindInCwd,
			// 3. look for ./.<appname>/config.<ext>
			FindInAppNameSubdir,
			// 4. look for ~/.<appname>.<ext>
			FindInHomeDir,
			// 5. look for <appname>/config.<ext> in xdg locations
			FindInXDG,
		},
	}
}

func (r *Config) AddFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&r.File, "config", "c", r.File, "configuration file")
}
