package fangs

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"github.com/anchore/go-logger"
)

func Summarize(cfg Config, descriptions DescriptionProvider, values ...any) string {
	root := &section{}
	for _, value := range values {
		v := reflect.ValueOf(value)
		summarize(cfg, descriptions, root, v, nil)
	}
	return root.stringify(cfg)
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
		if !includeField(f) {
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
			env := envVar(cfg.AppName, path...)
			// for slices of structs, do not output an env var
			if t.Kind() == reflect.Slice && baseType(t.Elem()).Kind() == reflect.Struct {
				env = ""
			}
			s.add(cfg.Logger,
				name,
				v,
				descriptions.GetDescription(v, f),
				env)
		}
	}
}

// printVal prints a value in YAML format
// nolint:gocognit
func printVal(cfg Config, value reflect.Value, indent string) string {
	buf := bytes.Buffer{}

	v, t := base(value)
	switch {
	case isSlice(t):
		if v.Len() == 0 {
			return "[]"
		}

		for i := 0; i < v.Len(); i++ {
			v := v.Index(i)
			buf.WriteString("\n")
			buf.WriteString(indent)
			buf.WriteString("- ")

			val := printVal(cfg, v, indent+"  ")
			val = strings.TrimSpace(val)
			buf.WriteString(val)

			// separate struct entries by an empty line
			_, t := base(v)
			if isStruct(t) {
				buf.WriteString("\n")
			}
		}

	case isStruct(t):
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !includeField(f) {
				continue
			}

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
				default:
					name = tag
				}
			}

			v := v.Field(i)

			buf.WriteString("\n")
			buf.WriteString(indent)

			val := printVal(cfg, v, indent+"  ")

			val = fmt.Sprintf("%s: %s", name, val)

			buf.WriteString(val)
		}

	case v.CanInterface():
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return ""
		}
		v := v.Interface()
		switch v.(type) {
		case string:
			return fmt.Sprintf("'%s'", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}

	val := buf.String()
	// for slices, there will be an extra newline, which we want to remove
	val = strings.TrimSuffix(val, "\n")
	return val
}

func base(v reflect.Value) (reflect.Value, reflect.Type) {
	t := v.Type()
	for isPtr(t) {
		t = t.Elem()
		if v.IsNil() {
			newV := reflect.New(t)
			// If the field we're looking at is nil, and is a pointer to a struct,
			// change it to point to an empty instance of the struct, so that we can
			// continue recursing on the config structure. However, if it's a nil pointer
			// to a primitive type, leave it as nil so that we can tell later in the summary
			// that it wasn't set.
			if newV.Kind() == reflect.Struct {
				v = newV
			}
		} else {
			v = v.Elem()
		}
	}
	return v, t
}

func baseType(t reflect.Type) reflect.Type {
	for isPtr(t) {
		t = t.Elem()
	}
	return t
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

func (s *section) stringify(cfg Config) string {
	out := &bytes.Buffer{}
	stringifySection(cfg, out, s, "")
	return out.String()
}

func stringifySection(cfg Config, out *bytes.Buffer, s *section, indent string) {
	nextIndent := indent

	if s.name != "" {
		nextIndent += "  "

		desc := &bytes.Buffer{}

		if s.description != "" {
			// support multi-line descriptions
			lines := strings.Split(strings.TrimSpace(s.description), "\n")
			for idx, line := range lines {
				desc.WriteString(indent + "# " + line)
				if idx < len(lines)-1 {
					desc.WriteString("\n")
				}
			}
		}
		if s.env != "" {
			value := fmt.Sprintf("(env: %s)", s.env)
			if s.description == "" {
				// since there is no description, we need to start the comment
				desc.WriteString(indent + "# ")
			} else {
				// buffer between description and env hint
				desc.WriteString(" ")
			}
			desc.WriteString(value)
		}
		if s.description != "" || s.env != "" {
			desc.WriteString("\n")
		}

		out.WriteString(wrapCommentLines(desc.String(), indent, 95))

		out.WriteString(indent)

		out.WriteString(s.name)
		out.WriteString(":")

		if s.value.IsValid() {
			val := printVal(cfg, s.value, indent+"  ")
			if val != "" {
				out.WriteString(" ")
			}
			out.WriteString(val)
		}

		out.WriteString("\n")
	}

	for _, s := range s.subsections {
		stringifySection(cfg, out, s, nextIndent)
		if len(s.subsections) == 0 {
			out.WriteString(nextIndent)
			out.WriteString("\n")
		}
	}
}

func wrapCommentLines(commentLine, indent string, limit int) string {
	commentLine = strings.TrimSpace(removeComment(commentLine))

	ogLines := strings.Split(commentLine, "(env: ")
	if len(ogLines) == 0 {
		return ""
	}

	if len(ogLines) > 1 {
		for i := 1; i < len(ogLines); i++ {
			ogLines[i] = "(env: " + ogLines[i]
		}
	}

	ogLines[0] = strings.TrimSpace(ogLines[0])
	if len(ogLines[0]) == 0 {
		if len(ogLines) > 1 {
			ogLines = ogLines[1:]
		} else {
			return ""
		}
	}

	var lines []string
	for _, line := range ogLines {
		wrappedLine := wordWrap(line, limit)
		lines = append(lines, strings.Split(wrappedLine, "\n")...)
	}

	return indent + "# " + strings.Join(lines, "\n"+indent+"# ") + "\n"
}

// source: https://codereview.stackexchange.com/questions/244435/word-wrap-in-go
func wordWrap(text string, lineWidth int) string {
	wrap := make([]byte, 0, len(text)+2*len(text)/lineWidth)
	eoLine := lineWidth
	inWord := false
	for i, j := 0, 0; ; {
		r, size := utf8.DecodeRuneInString(text[i:])
		if size == 0 && r == utf8.RuneError {
			r = ' '
		}
		if unicode.IsSpace(r) {
			if inWord {
				if i >= eoLine {
					wrap = append(wrap, '\n')
					eoLine = len(wrap) + lineWidth
				} else if len(wrap) > 0 {
					wrap = append(wrap, ' ')
				}
				wrap = append(wrap, text[j:i]...)
			}
			inWord = false
		} else if !inWord {
			inWord = true
			j = i
		}
		if size == 0 && r == ' ' {
			break
		}
		i += size
	}
	return string(wrap)
}

func removeComment(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "# ", ""))
}
