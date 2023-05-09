package fangs

import (
	"os"
	"regexp"
	"strings"
)

func Indent(text, indent string) string {
	if len(strings.TrimSpace(text)) == 0 {
		return indent
	}
	if text[len(text)-1:] == "\n" {
		result := ""
		for _, j := range strings.Split(text[:len(text)-1], "\n") {
			result += indent + j + "\n"
		}
		return result
	}
	result := ""
	for _, j := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		result += indent + j + "\n"
	}
	return result[:len(result)-1]
}

func contains(parts []string, value string) bool {
	for _, v := range parts {
		if v == value {
			return true
		}
	}
	return false
}

var envVarRegex = regexp.MustCompile("[^a-zA-Z0-9_]")

func envVar(appName string, parts []string) string {
	v := strings.Join(parts, "_")
	if appName != "" {
		v = appName + "_" + v
	}
	v = envVarRegex.ReplaceAllString(v, "_")
	return strings.ToUpper(v)
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
