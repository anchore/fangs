package fangs

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Describer interface {
	Describe(value reflect.Value, field reflect.StructField) string
}

type FinderDescriber func(cfg Config, finder Finder) []string

func Summarize(cfg Config, value interface{}, describers ...Describer) string {
	describers = append(describers, &structFieldDescriber{})
	out := summarize(cfg, reflect.ValueOf(value), nil, describers, "")
	return strings.TrimSpace(out)
}

func SummarizeLocations(cfg Config, describers ...FinderDescriber) (out []string) {
	for _, f := range cfg.Finders {
		for _, d := range describers {
			out = append(out, d(cfg, f)...)
		}
	}
	return
}

//nolint:gocognit
func summarize(cfg Config, value reflect.Value, path []string, describers []Describer, indent string) string {
	out := bytes.Buffer{}

	v, t := base(value)

	if !isStruct(t) {
		panic(fmt.Sprintf("Summarize requires struct types, got: %+v", value.Interface()))
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		path := path
		name := field.Name

		if tag, ok := field.Tag.Lookup(cfg.TagName); ok {
			parts := strings.Split(tag, ",")
			tag = parts[0]
			if tag == "-" {
				continue
			}
			switch {
			case contains(parts, "squash"):
				name = ""
			case tag == "":
				path = append(path, name)
			default:
				name = tag
				path = append(path, tag)
			}
		} else {
			path = append(path, name)
		}

		v, t := base(v.Field(i))

		var section string
		if isStruct(t) {
			if name == "" {
				section = summarize(cfg, v, path, describers, indent)
			} else {
				section = fmt.Sprintf("%s:\n%s",
					name,
					summarize(cfg, v, path, describers, indent+"  "))
			}
		} else {
			envVar := envVar(cfg.AppName, path)

			description := ""
			for _, d := range describers {
				description = d.Describe(v, field)
				if description != "" {
					break
				}
			}

			section = fmt.Sprintf("%s: %s # %s (env var: %s)\n\n", name, printVal(v), description, envVar)
		}

		section = Indent(section, indent)

		out.WriteString(section)
	}

	return out.String()
}

func printVal(value reflect.Value) string {
	v, _ := base(value)
	if v.CanInterface() {
		v := v.Interface()
		switch v.(type) {
		case string:
			return fmt.Sprintf("'%s'", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func base(v reflect.Value) (reflect.Value, reflect.Type) {
	if isPtr(v.Type()) {
		v = v.Elem()
		return v, v.Type()
	}
	return v, v.Type()
}

type structFieldDescriber struct{}

var _ Describer = (*structFieldDescriber)(nil)

func (*structFieldDescriber) Describe(_ reflect.Value, field reflect.StructField) string {
	return field.Tag.Get("description")
}

type commandDescriber struct {
	tag      string
	flagRefs flagRefs
}

var _ Describer = (*commandDescriber)(nil)

func NewCommandDescriber(cfg Config, cmd *cobra.Command) Describer {
	return &commandDescriber{
		tag:      cfg.TagName,
		flagRefs: collectFlagRefs(cmd),
	}
}

func (d *commandDescriber) Describe(v reflect.Value, _ reflect.StructField) string {
	if v.CanAddr() {
		v = v.Addr()
		f := d.flagRefs[v.Pointer()]
		if f != nil {
			return f.Usage
		}
	}
	return ""
}

func collectFlagRefs(cmd *cobra.Command) flagRefs {
	out := getFlagRefs(cmd.PersistentFlags(), cmd.Flags())
	for _, c := range cmd.Commands() {
		for k, v := range collectFlagRefs(c) {
			out[k] = v
		}
	}
	return out
}

type DirectDescriber struct {
	flagRefs flagRefs
}

var _ Describer = (*DirectDescriber)(nil)

func NewDescriber() *DirectDescriber {
	return &DirectDescriber{
		flagRefs: flagRefs{},
	}
}

func (d *DirectDescriber) Add(ptr any, description string) {
	v := reflect.ValueOf(ptr)
	if !isPtr(v.Type()) {
		panic(fmt.Sprintf("Descriptions.Add requires a pointer, but got: %+v", ptr))
	}
	p := v.Pointer()
	d.flagRefs[p] = &pflag.Flag{
		Usage: description,
	}
}

func (d *DirectDescriber) Describe(v reflect.Value, _ reflect.StructField) string {
	if v.CanAddr() {
		v = v.Addr()
	}
	if isPtr(v.Type()) {
		f := d.flagRefs[v.Pointer()]
		if f != nil {
			return f.Usage
		}
	}
	return ""
}
