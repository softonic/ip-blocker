package actor

import (
	"reflect"
	"testing"
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

	result := uniqueItems(elasticIPs, armorIPs)

	result = removeDuplicateStr(result)

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

	result := uniqueItems(elasticIPs, exceptionsIPs)

	result = removeDuplicateStr(result)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Error actual = %v, and Expected = %v.", result, expected)
	}

}
