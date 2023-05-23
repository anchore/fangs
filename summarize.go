package fangs

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"

	"github.com/anchore/go-logger"
)

func Summarize(cfg Config, descriptions DescriptionProvider, values ...any) string {
	root := &section{}
	for _, value := range values {
		v := reflect.ValueOf(value)
		summarize(cfg, descriptions, root, v, nil)
	}
	return root.stringify()
}

func SummarizeCommand(cfg Config, cmd *cobra.Command, values ...any) string {
	root := cmd
	for root.Parent() != nil {
		root = root.Parent()
	}
	descriptions := DescriptionProviders(
		NewFieldDescriber(values...),
		NewStructDescriptionTagProvider(),
		NewCommandFlagDescriptionProvider(cfg.TagName, root),
	)
	return Summarize(cfg, descriptions, values...)
}

func SummarizeLocations(cfg Config) (out []string) {
	for _, f := range cfg.Finders {
		out = append(out, f(cfg)...)
	}
	return
}

//nolint:gocognit
func summarize(cfg Config, descriptions DescriptionProvider, s *section, value reflect.Value, path []string) {
	v, t := base(value)

	if !isStruct(t) {
		panic(fmt.Sprintf("Summarize requires struct types, got: %#v", value.Interface()))
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		path := path
		name := f.Name

		if tag, ok := f.Tag.Lookup(cfg.TagName); ok {
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

		v := v.Field(i)
		_, t := base(v)

		if isStruct(t) {
			sub := s
			if name != "" {
				sub = s.sub(name)
			}
			if isPtr(v.Type()) && v.IsNil() {
				v = reflect.New(t)
			}
			summarize(cfg, descriptions, sub, v, path)
		} else {
			s.add(cfg.Logger,
				name,
				v,
				descriptions.GetDescription(v, f),
				envVar(cfg.AppName, path))
		}
	}
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
	t := v.Type()
	for isPtr(t) {
		t = t.Elem()
		if v.IsNil() {
			v = reflect.New(t)
		} else {
			v = v.Elem()
		}
	}
	return v, t
}

type section struct {
	name        string
	value       reflect.Value
	description string
	env         string
	subsections []*section
}

func (s *section) get(name string) *section {
	for _, s := range s.subsections {
		if s.name == name {
			return s
		}
	}
	return nil
}

func (s *section) sub(name string) *section {
	sub := s.get(name)
	if sub == nil {
		sub = &section{
			name: name,
		}
		s.subsections = append(s.subsections, sub)
	}
	return sub
}

func (s *section) add(log logger.Logger, name string, value reflect.Value, description string, env string) *section {
	add := &section{
		name:        name,
		value:       value,
		description: description,
		env:         env,
	}
	sub := s.get(name)
	if sub != nil {
		if sub.name != name || !sub.value.CanConvert(value.Type()) || sub.description != description || sub.env != env {
			log.Warnf("multiple entries with different values: %#v != %#v", sub, add)
		}
		return sub
	}
	s.subsections = append(s.subsections, add)
	return add
}

func (s *section) stringify() string {
	out := &bytes.Buffer{}
	stringifySection(out, s, "")
	return out.String()
}

func stringifySection(out *bytes.Buffer, s *section, indent string) {
	nextIndent := indent

	if s.name != "" {
		nextIndent = "  "

		out.WriteString(indent)

		out.WriteString(s.name)
		out.WriteString(":")

		if s.value.IsValid() {
			out.WriteString(" ")
			out.WriteString(printVal(s.value))
		}

		if s.description != "" || s.env != "" {
			out.WriteString(" #")
			if s.description != "" {
				out.WriteString(" ")
				out.WriteString(s.description)
			}
			if s.env != "" {
				out.WriteString(" (env: ")
				out.WriteString(s.env)
				out.WriteString(")")
			}
		}

		out.WriteString("\n")
	}

	for _, s := range s.subsections {
		stringifySection(out, s, nextIndent)
		if len(s.subsections) == 0 {
			out.WriteString(nextIndent)
			out.WriteString("\n")
		}
	}
}
