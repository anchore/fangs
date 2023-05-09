package fangs

import (
	"fmt"
	"path"
	"reflect"

	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

func AllDescribers() []FinderDescriber {
	return []FinderDescriber{
		FindDirectDescriber,
		FindConfigYamlInCwdDescriber,
		FindInCwdDescriber,
		FindInAppNameSubdirDescriber,
		FindInHomeDirDescriber,
		FindInXDGDescriber,
	}
}

func IsFinder(f1 any, f2 any) bool {
	v1 := reflect.ValueOf(f1)
	v2 := reflect.ValueOf(f2)
	return v1.Pointer() == v2.Pointer()
}

func FindDirectDescriber(cfg Config, finder Finder) []string {
	if !IsFinder(finder, FindDirect) || cfg.File == "" {
		return nil
	}
	file, err := homedir.Expand(cfg.File)
	if err != nil {
		cfg.Logger.Debugf("unable to expand path: %s", cfg.File)
		file = cfg.File
	}
	return []string{file}
}

func FindConfigYamlInCwdDescriber(_ Config, finder Finder) []string {
	if !IsFinder(finder, FindConfigYamlInCwd) {
		return nil
	}
	return []string{"./config.yaml"}
}

func FindInCwdDescriber(cfg Config, finder Finder) []string {
	if !IsFinder(finder, FindInCwd) {
		return nil
	}
	return describeConfigFiles(".", "."+cfg.AppName)
}

func FindInAppNameSubdirDescriber(cfg Config, finder Finder) []string {
	if !IsFinder(finder, FindInCwd) {
		return nil
	}
	return describeConfigFiles("."+cfg.AppName, "config")
}

func FindInHomeDirDescriber(cfg Config, finder Finder) []string {
	if !IsFinder(finder, FindInHomeDir) {
		return nil
	}
	home, err := homedir.Dir()
	if err != nil {
		cfg.Logger.Debugf("unable to determine home dir: %w", err)
		return nil
	}
	return describeConfigFiles(home, "."+cfg.AppName)
}

func FindInXDGDescriber(cfg Config, finder Finder) (out []string) {
	if !IsFinder(finder, FindInXDG) {
		return nil
	}
	dirs := []string{path.Join(xdg.ConfigHome, cfg.AppName)}
	for _, dir := range xdg.ConfigDirs {
		dirs = append(dirs, path.Join(dir, cfg.AppName))
	}
	for _, dir := range dirs {
		out = append(out, describeConfigFiles(dir, "config")...)
	}
	return
}

func describeConfigFiles(dir string, base string) (out []string) {
	for _, ext := range viper.SupportedExts {
		name := path.Join(dir, fmt.Sprintf("%s.%s", base, ext))
		out = append(out, name)
	}
	return
}
