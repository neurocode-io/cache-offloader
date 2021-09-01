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

func IsValidHTTPMethod(maybeMethod string) bool {
	methods := []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}

	for _, m := range methods {
		if m == strings.ToUpper(maybeMethod) {
			return true
		}
	}

	return false
}
