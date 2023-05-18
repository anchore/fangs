package fangs

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
)

func Summarize(cfg Config, descriptions DescriptionProvider, values ...any) string {
	out := ""
	for _, value := range values {
		v := reflect.ValueOf(value)
		out += summarize(cfg, descriptions, v, nil, "")
	}
	return strings.TrimSpace(out)
}

func SummarizeCommand(cfg Config, cmd *cobra.Command, values ...any) string {
	root := cmd
	for root.Parent() != nil {
		root = root.Parent()
	}
	descriptions := DescriptionProviders(
		NewStructDescriber(values...),
		NewStructDescriptionTagProvider(),
		NewCommandDescriber(cfg.TagName, root),
	)
	out := Summarize(cfg, descriptions, values...)
	return strings.TrimSpace(out)
}

func SummarizeLocations(cfg Config) (out []string) {
	for _, f := range cfg.Finders {
		out = append(out, f(cfg)...)
	}
	return
}

//nolint:gocognit
func summarize(cfg Config, descriptions DescriptionProvider, value reflect.Value, path []string, indent string) string {
	out := bytes.Buffer{}

	v, t := base(value)

	if !isStruct(t) {
		panic(fmt.Sprintf("Summarize requires struct types, got: %#v", value.Interface()))
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
				section = summarize(cfg, descriptions, v, path, indent)
			} else {
				section = fmt.Sprintf("%s:\n%s",
					name,
					summarize(cfg, descriptions, v, path, indent+"  "))
			}
		} else {
			envVar := envVar(cfg.AppName, path)

			description := descriptions.GetDescription(v, field)

			section = fmt.Sprintf("%s: %s # %s (env: %s)\n\n", name, printVal(v), description, envVar)
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

type commandDescriber struct {
	tag      string
	flagRefs flagRefs
}

var _ DescriptionProvider = (*commandDescriber)(nil)

func NewCommandDescriber(tagName string, cmd *cobra.Command) DescriptionProvider {
	return &commandDescriber{
		tag:      tagName,
		flagRefs: collectFlagRefs(cmd),
	}
}

func (d *commandDescriber) GetDescription(v reflect.Value, _ reflect.StructField) string {
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
