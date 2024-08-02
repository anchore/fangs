package fangs

import (
	"fmt"
	"os"
	"strings"

	"github.com/anchore/go-logger"
	"github.com/anchore/go-logger/adapter/discard"
)

type Config struct {
	Logger   logger.Logger `yaml:"-" json:"-" mapstructure:"-"`
	AppName  string        `yaml:"-" json:"-" mapstructure:"-"`
	TagName  string        `yaml:"-" json:"-" mapstructure:"-"`
	Files    []string      `yaml:"-" json:"-" mapstructure:"-"`
	Finders  []Finder      `yaml:"-" json:"-" mapstructure:"-"`
	Profiles []string      `yaml:"-" json:"-" mapstructure:"-"`
}

var _ FlagAdder = (*Config)(nil)

// NewConfig creates a new Config object with defaults
func NewConfig(appName string) Config {
	return Config{
		Logger:  discard.New(),
		AppName: appName,
		TagName: "mapstructure",
		// search for configs in specific order
		Finders: []Finder{
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

// WithConfigEnvVar looks for the environment variable: <APP_NAME>_CONFIG as a way to specify a config file
// This will be overridden by a command-line flag
func (c Config) WithConfigEnvVar() Config {
	c.Files = strings.Split(os.Getenv(envVar(c.AppName, "CONFIG")), ",")
	return c
}

func (c *Config) AddFlags(flags FlagSet) {
	flags.StringArrayVarP(&c.Files, "config", "c", fmt.Sprintf("%s configuration file", c.AppName))
	flags.StringArrayVarP(&c.Profiles, "profiles", "", "configuration profiles to use")
}
