package utils

import (
	"regexp"
	"strings"
)

func VariableMatchesRegexIn(variable string, list []string) bool {
	for _, value := range list {
		value = strings.TrimSuffix(value, "/")
		matched, err := regexp.MatchString(value, variable)
		if err != nil {
			return false
		}

		if matched {
			return true
		}
	}

	return false
}
