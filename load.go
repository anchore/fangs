package fangs

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
	return LoadConfig(cfg, commandFlagRefs(cmd), configurations...)
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

func LoadConfig(cfg Config, flags flagRefs, configurations ...any) error {
	// ensure the config is set up sufficiently
	if cfg.Logger == nil || cfg.Finders == nil {
		return fmt.Errorf("config.Load requires logger and finders to be set, but only has %+v", cfg)
	}

	// allow for nested options to be specified via environment variables
	// e.g. pod.context = APPNAME_POD_CONTEXT
	v := viper.NewWithOptions(viper.EnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_")))

	return load(cfg, v, flags, configurations...)
}

func load(cfg Config, v *viper.Viper, flags flagRefs, configurations ...any) error {
	for _, configuration := range configurations {
		if !isPtr(reflect.TypeOf(configuration)) {
			return fmt.Errorf("config.Load configuration parameters must be a pointers, got: %s -- %v", reflect.TypeOf(configuration).Name(), configuration)
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

// configureViper loads the default configuration values into the viper instance,
// before the config values are read and parsed. the value _must_ be a pointer but
// may be a pointer to a pointer
func configureViper(cfg Config, v *viper.Viper, value reflect.Value, flags flagRefs, appPrefix string, path string) {
	typ := value.Type()
	if !isPtr(typ) {
		panic(fmt.Sprintf("configureViper value must be a pointer, got: %+v", value))
	}

	// value is always a pointer, addr within a struct
	ptr := value.Pointer()
	value = value.Elem()
	typ = value.Type()

	// might be a pointer value
	if isPtr(typ) {
		typ = typ.Elem()
		value = value.Elem()
	}

	if !isStruct(typ) {
		envVar := strings.ToUpper(regexp.MustCompile("[^a-zA-Z0-9]").ReplaceAllString(appPrefix+path, "_"))

		if flag, ok := flags[ptr]; ok {
			cfg.Logger.Tracef("binding env var w/flag: %s", envVar)
			err := v.BindPFlag(path, flag)
			if err != nil {
				cfg.Logger.Debugf("unable to bind flag: %s to %+v", path, flag)
			}
			return
		}

		cfg.Logger.Tracef("binding env var: %s", envVar)

		v.SetDefault(path, nil) // no default value actually needs to be set for Viper to read config values
		return
	}

	prefix := path

	if prefix != "" {
		prefix += "."
	}

	// for each field in the configuration struct, see if the field implements the defaultValueLoader interface and invoke it if it does
	for i := 0; i < value.NumField(); i++ {
		fieldValue := value.Field(i)
		field := typ.Field(i)

		name := field.Name

		if tag, ok := field.Tag.Lookup("mapstructure"); ok {
			// handle ,squash mapstructure tags
			tag = strings.Split(tag, ",")[0]
			if tag == "-" {
				continue
			}
			if tag == "" {
				name = path
			} else {
				name = prefix + tag
			}
		} else {
			name = prefix + name
		}

		configureViper(cfg, v, fieldValue.Addr(), flags, appPrefix, name)
	}
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
	if isPtr(typ) {
		if p, ok := obj.(PostLoad); ok {
			// the field implements parser, call it
			if err := p.PostLoad(); err != nil {
				return err
			}
		}
		value = value.Elem()
		typ = value.Type()
	}

	if !isStruct(typ) {
		return nil
	}

	// parse nested config options
	// for each field in the configuration struct, see if the field implements the parser interface
	// note: the app config is a pointer, so we need to grab the elements explicitly (to traverse the address)
	for i := 0; i < value.NumField(); i++ {
		f := value.Field(i)
		ft := f.Type()
		if isPtr(ft) {
			f = f.Elem()
			ft = f.Type()
		}
		if !f.CanAddr() || !isStruct(ft) {
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

type flagRefs map[uintptr]*pflag.Flag

func commandFlagRefs(cmd *cobra.Command) flagRefs {
	return getFlagRefs(cmd.PersistentFlags(), cmd.Flags())
}

func getFlagRefs(flagSets ...*pflag.FlagSet) flagRefs {
	refs := flagRefs{}
	for _, flags := range flagSets {
		flags.VisitAll(func(flag *pflag.Flag) {
			refs[getFlagRef(flag)] = flag
		})
	}
	return refs
}

func getFlagRef(flag *pflag.Flag) uintptr {
	v := reflect.ValueOf(flag.Value)

	// check for struct types like stringArrayValue
	if isPtr(v.Type()) {
		vf := v.Elem()
		vt := vf.Type()
		if isStruct(vt) {
			if _, ok := vt.FieldByName("value"); ok {
				vf = vf.FieldByName("value")
				if vf.IsValid() {
					v = vf
				}
			}
		}
	}
	return v.Pointer()
}

func upperFirst(p string) string {
	if len(p) < 2 {
		return strings.ToUpper(p)
	}
	return strings.ToUpper(p[0:1]) + p[1:]
}

func isPtr(typ reflect.Type) bool {
	return typ.Kind() == reflect.Ptr
}

func isStruct(typ reflect.Type) bool {
	return typ.Kind() == reflect.Struct
}

func isNotFoundErr(err error) bool {
	var notFound *viper.ConfigFileNotFoundError
	return err != nil && errors.As(err, &notFound)
}
