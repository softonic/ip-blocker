package actor

import (
	"reflect"
	"testing"

	"github.com/softonic/ip-blocker/app/utils"
)

func TestDetectWhichOfTheseIPsAreNotBlocked(t *testing.T) {

	/* 	elasticIPs := []app.IPCount{
		{
			IP:    "1.1.1.1",
			Count: 2,
		},
		{
			IP:    "2.2.2.2",
			Count: 2,
		},
	} */
	elasticIPs := []string{
		"1.1.1.1",
		"2.2.2.2",
		"2.2.2.2",
	}

	armorIPs := []string{
		"1.1.1.1",
		"3.3.3.3",
		"4.4.4.4",
	}

	expected := []string{
		"2.2.2.2",
	}

	result := utils.UniqueItems(elasticIPs, armorIPs)

	result = utils.RemoveDuplicateStr(result)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Error actual = %v, and Expected = %v.", result, expected)
	}

}

func TestDetectWhichOfTheseIPsAreNotExcluded(t *testing.T) {

	/* 	elasticIPs := []app.IPCount{
		{
			IP:    "1.1.1.1",
			Count: 2,
		},
		{
			IP:    "2.2.2.2",
			Count: 2,
		},
	} */
	elasticIPs := []string{
		"1.1.1.1",
		"2.2.2.2",
		"2.2.2.2",
	}

	exceptionsIPs := []string{
		"1.1.1.1",
		"3.3.3.3",
		"4.4.4.4",
	}

	expected := []string{
		"2.2.2.2",
	}

	result := utils.UniqueItems(elasticIPs, exceptionsIPs)

	result = utils.RemoveDuplicateStr(result)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Error actual = %v, and Expected = %v.", result, expected)
	}

}
