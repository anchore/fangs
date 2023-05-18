package fangs

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/anchore/go-logger"
	"github.com/anchore/go-logger/adapter/discard"
)

type Config struct {
	Logger  logger.Logger
	AppName string
	TagName string
	File    string
	Finders []Finder
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

func (c *Config) AddFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&c.File, "config", "c", c.File, fmt.Sprintf("%s configuration file", c.AppName))
}
