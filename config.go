package fangs

import (
	"fmt"
	"os"

	"github.com/anchore/go-logger"
	"github.com/anchore/go-logger/adapter/discard"
)

type Config struct {
	Logger  logger.Logger `yaml:"-" json:"-" mapstructure:"-"`
	AppName string        `yaml:"-" json:"-" mapstructure:"-"`
	TagName string        `yaml:"-" json:"-" mapstructure:"-"`
	File    string        `yaml:"-" json:"-" mapstructure:"-"`
	Finders []Finder      `yaml:"-" json:"-" mapstructure:"-"`
}

var _ FlagAdder = (*Config)(nil)

func NewConfig(appName string) Config {
	return Config{
		Logger:  discard.New(),
		AppName: appName,
		TagName: "mapstructure",
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

func (c Config) WithConfigEnvVar() Config {
	// check for environment variable <app_name>_CONFIG, if set this will be the default
	// but will be overridden by a command-line flag
	c.File = os.Getenv(envVar(c.AppName, "CONFIG"))
	return c
}

func (c *Config) AddFlags(flags FlagSet) {
	flags.StringVarP(&c.File, "config", "c", fmt.Sprintf("%s configuration file", c.AppName))
}
