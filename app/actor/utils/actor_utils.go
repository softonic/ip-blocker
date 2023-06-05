package utils

import (
	"regexp"
)

func ExtractFromDescription(description string) string {
	stringToParse := "ipblocker:"
	reg := regexp.MustCompile(stringToParse)
	res := reg.ReplaceAllString(description, "${1}")
	return res
}
