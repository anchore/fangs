package config

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func Load(cfg Config, cmd *cobra.Command, configurations ...any) error {
	// allow for nested options to be specified via environment variables
	// e.g. pod.context = APPNAME_POD_CONTEXT
	v := viper.NewWithOptions(viper.EnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_")))

	return load(cfg, v, cmd, configurations...)
}

func LoadAt(cfg Config, cmd *cobra.Command, path string, configuration any) error {
	t := reflect.TypeOf(configuration)
	config := reflect.StructOf([]reflect.StructField{{
		Name: upperFirst(path),
		Type: t,
		Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s" yaml:"%s" mapstructure:"%s"`, path, path, path)),
	}})

	value := reflect.New(config)
	value.Elem().Field(0).Set(reflect.ValueOf(configuration))

	return Load(cfg, cmd, value.Interface())
}

func upperFirst(p string) string {
	if len(p) < 2 {
		return strings.ToUpper(p)
	}
	return strings.ToUpper(p[0:1]) + p[1:]
}

func load(cfg Config, v *viper.Viper, cmd *cobra.Command, configurations ...any) error {
	for _, cfg := range configurations {
		if reflect.TypeOf(cfg).Kind() != reflect.Ptr {
			return fmt.Errorf("config.Load configuration parameters must be a pointers, got: %s -- %v", reflect.TypeOf(cfg).Name(), cfg)
		}
	}

	// priority order: viper.Set, flag, env, config, kv, defaults
	// flags have already been loaded into viper by command construction

	// check if user specified config; otherwise read all possible paths
	if err := loadConfig(cfg, v); err != nil {
		if isNotFoundErr(err) {
			cfg.Logger.Debug("no config file found, using defaults")
		} else {
			return fmt.Errorf("unable to load config: %w", err)
		}
	}

	// load environment variables
	v.SetEnvPrefix(cfg.AppName)
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	appPrefix := cfg.AppName
	if appPrefix != "" {
		appPrefix += "."
	}

	flags := getFlagRefs(cmd)

	for _, configuration := range configurations {
		configureViper(cfg, v, reflect.ValueOf(configuration), flags, appPrefix, "")

		// unmarshal fully populated viper object onto config
		err := v.Unmarshal(configuration)
		if err != nil {
			return err
		}

		// Convert all populated config options to their internal application values ex: scope string => scopeOpt source.Scope
		err = postLoad(configuration)
		if err != nil {
			return err
		}
	}

	return nil
}

// configureViper loads the default configuration values into the viper instance, before the config values are read and parsed
func configureViper(cfg Config, v *viper.Viper, value reflect.Value, flags flagRefs, appPrefix string, path string) {
	if value.Type().Kind() == reflect.Ptr && value.Type().Elem().Kind() != reflect.Struct {
		if flag, ok := flags[value.Pointer()]; ok {
			cfg.Logger.Tracef("binding: %s = %v (flag)\n", strings.ToUpper(regexp.MustCompile("[^a-zA-Z0-9]").ReplaceAllString(appPrefix+path, "_")), value.Elem().Interface())
			_ = v.BindPFlag(path, flag)
			return
		}
	}

	if value.Type().Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Type().Kind() != reflect.Struct {
		cfg.Logger.Tracef("binding: %s = %v\n", strings.ToUpper(regexp.MustCompile("[^a-zA-Z0-9]").ReplaceAllString(appPrefix+path, "_")), value.Interface())
		v.SetDefault(path, value.Interface())
		return
	}

	if path != "" {
		path += "."
	}

	// for each field in the configuration struct, see if the field implements the defaultValueLoader interface and invoke it if it does
	for i := 0; i < value.NumField(); i++ {
		fieldValue := value.Field(i)
		field := value.Type().Field(i)

		mapStructTag := field.Tag.Get("mapstructure")

		if mapStructTag == "-" {
			continue
		}

		if !field.Anonymous && mapStructTag == "" {
			cfg.Logger.Tracef("not binding field due to lacking mapstructure tag: %s.%s", value.Type().Name(), field.Name)
			continue
		}

		if fieldValue.Type().Kind() != reflect.Ptr {
			fieldValue = fieldValue.Addr()
		}

		configureViper(cfg, v, fieldValue, flags, appPrefix, path+mapStructTag)
	}
}

type flagRefs map[uintptr]*pflag.Flag

func getFlagRefs(cmd *cobra.Command) flagRefs {
	refs := flagRefs{}
	for _, flags := range []*pflag.FlagSet{cmd.PersistentFlags(), cmd.Flags()} {
		flags.VisitAll(func(flag *pflag.Flag) {
			v := reflect.ValueOf(flag.Value)
			// check for struct types like stringArrayValue
			if v.Type().Kind() == reflect.Ptr {
				vf := v.Elem()
				if vf.Type().Kind() == reflect.Struct {
					if _, ok := vf.Type().FieldByName("value"); ok {
						vf = vf.FieldByName("value")
						if vf.IsValid() {
							v = vf
						}
					}
				}
			}
			refs[v.Pointer()] = flag
		})
	}
	return refs
}

func loadConfig(cfg Config, v *viper.Viper) error {
	for _, finder := range cfg.Finders {
		files := finder(cfg)
		if files == nil {
			continue
		}
		for _, file := range files {
			v.SetConfigFile(file)
			err := v.ReadInConfig()
			if isNotFoundErr(err) {
				continue
			}
			if err != nil {
				return err
			}
			v.Set("config", v.ConfigFileUsed())
			return nil
		}
	}
	return &viper.ConfigFileNotFoundError{}
}

func postLoad(obj any) error {
	value := reflect.ValueOf(obj)
	typ := value.Type()
	if typ.Kind() == reflect.Ptr {
		if p, ok := obj.(PostLoad); ok {
			// the field implements parser, call it
			if err := p.PostLoad(); err != nil {
				return err
			}
		}
		value = value.Elem()
		typ = value.Type()
	}

	if typ.Kind() != reflect.Struct {
		return nil
	}

	// parse nested config options
	// for each field in the configuration struct, see if the field implements the parser interface
	// note: the app config is a pointer, so we need to grab the elements explicitly (to traverse the address)
	for i := 0; i < value.NumField(); i++ {
		f := value.Field(i)
		ft := f.Type()
		if ft.Kind() == reflect.Ptr {
			f = f.Elem()
			ft = f.Type()
		}
		if !f.CanAddr() || ft.Kind() != reflect.Struct {
			continue
		}
		// note: since the interface method of parser is a pointer receiver we need to get the value of the field as a pointer.
		// the field implements parser, call it
		if err := postLoad(f.Addr().Interface()); err != nil {
			return err
		}
	}

	return nil
}

func isNotFoundErr(err error) bool {
	var notFound *viper.ConfigFileNotFoundError
	return err != nil && errors.As(err, &notFound)
}
