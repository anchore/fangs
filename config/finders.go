package config

import (
	"fmt"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type Finder func(cfg Config) []string

// FindDirect attempts to find a directly configured cfg.File
func FindDirect(cfg Config) []string {
	if cfg.File == "" {
		return nil
	}
	file, err := homedir.Expand(cfg.File)
	if err != nil {
		cfg.Logger.Debugf("unable to expand path: %s", cfg.File)
		file = cfg.File
	}
	if fileExists(file) {
		return []string{file}
	}
	return nil
}

// FindInCwdConfigYaml loads ./config.yaml
// NOTE: THIS IS WAS A BUG IN SYFT/et. al. and is DEPRECATED and SHOULD NOT BE USED
// it is just here for temporary compatibility -- See syft #1634
func FindInCwdConfigYaml(cfg Config) []string {
	// check if config.yaml exists in the current directory
	// DEPRECATED: this will be removed in v1.0.0
	f := "./config.yaml"
	if fileExists(f) {
		cfg.Logger.Warnf("DEPRECATED: %s as a configuration file is deprecated and will be removed as an option in v1.0.0, please rename to .syft.yaml", f)
		return []string{f}
	}

	return nil
}

func FindInCwd(cfg Config) []string {
	return findConfigFiles(".", "."+cfg.AppName)
}

// FindInHomeDir loads from
func FindInHomeDir(cfg Config) []string {
	home, err := homedir.Dir()
	if err != nil {
		cfg.Logger.Debugf("unable to determine home dir: %w", err)
		return nil
	}
	return findConfigFiles(home, "."+cfg.AppName)
}

// FindInXDG looks for <appname>/config.yaml in xdg locations, starting with xdg home config dir then moving upwards
func FindInXDG(cfg Config) (out []string) {
	dirs := []string{path.Join(xdg.ConfigHome, cfg.AppName)}
	for _, dir := range xdg.ConfigDirs {
		dirs = append(dirs, path.Join(dir, cfg.AppName))
	}
	for _, dir := range dirs {
		out = append(out, findConfigFiles(dir, "config")...)
	}
	return
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func findConfigFiles(dir string, base string) (out []string) {
	for _, ext := range viper.SupportedExts {
		name := path.Join(dir, fmt.Sprintf("%s.%s", base, ext))
		if fileExists(name) {
			out = append(out, name)
		}
	}
	return
}