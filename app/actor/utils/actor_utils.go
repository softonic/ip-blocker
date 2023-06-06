package utils

import (
	"errors"
	"regexp"
	"strings"

	"k8s.io/klog"
)

// ExtractFromDescription returns the string that matches the regular expression
func ExtractFromDescription(description string) string {
	stringToParse := "ipblocker:"
	reg := regexp.MustCompile(stringToParse)
	res := reg.ReplaceAllString(description, "${1}")
	return res
}

func ConvertCSVToArray(excludeIps string) ([]string, error) {

	listIPs := strings.Split(excludeIps, ",")
	// if returns an empty slice, it means that the excludeIps is empty
	if listIPs == nil && excludeIps != "" {
		klog.Error("Error parsing excludeIps")
		return nil, errors.New("Error parsing excludeIps")
	}
	return listIPs, nil
}
