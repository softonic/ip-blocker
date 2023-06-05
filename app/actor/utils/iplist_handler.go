package utils

import (
	globalUtils "github.com/softonic/ip-blocker/app/utils"
)

type IPListHandler interface {
	UniqueItems(a []string, b []string) []string
	RemoveDuplicateStr(items []string) []string
}

type UtilsIPListHandler struct{}

func (h UtilsIPListHandler) UniqueItems(a []string, b []string) []string {
	return globalUtils.UniqueItems(a, b)
}

func (h UtilsIPListHandler) RemoveDuplicateStr(items []string) []string {
	return globalUtils.RemoveDuplicateStr(items)
}
