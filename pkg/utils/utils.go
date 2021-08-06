package utils

import (
	"regexp"
)

func VariableMatchesRegexIn(variable string, list []string) bool {
	for _, value := range list {

		if matched, _ := regexp.MatchString(value, variable); matched {
			return true
		}
	}
	return false
}
