package utils

import (
	"errors"
	"strings"

	"k8s.io/klog"
)

// This function is used to remove duplicates
func RemoveDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// This funcion return true if the element is in the slice
func Find(slice []int32, val int32) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// This function is used to compare two slices and return the elements that are not in the second slice
func UniqueItems(sourceIPs []string, exceptionsIPs []string) []string {

	var ipWithMaskES string
	candidateIPsBlocked := []string{}

	for _, elasticIps := range sourceIPs {
		count := 0
		for _, armorIps := range exceptionsIPs {
			ipWithMaskES = elasticIps
			if ipWithMaskES == armorIps || ipWithMaskES == armorIps+"/32" || ipWithMaskES+"/32" == armorIps {
				count++
			}
		}
		if count == 0 {
			candidateIPsBlocked = append(candidateIPsBlocked, elasticIps)
		}

	}

	return candidateIPsBlocked

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
