package utils

import (
	"regexp"
	"strings"
)

func VariableMatchesRegexIn(variable string, list []string) bool {
	for _, value := range list {
		value = strings.TrimSuffix(value, "/")
		if matched, _ := regexp.MatchString(value, variable); matched {
			return true
		}
	}
	return false
}
