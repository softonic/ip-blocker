package actor

import (
	"reflect"
	"testing"

	"github.com/softonic/ip-blocker/app"
)

func TestDetectWhichOfTheseIPsAreNotBlocked(t *testing.T) {

	elasticIPs := []app.IPCount{
		{
			IP:    "1.1.1.1",
			Count: 2,
		},
		{
			IP:    "2.2.2.2",
			Count: 2,
		},
	}

	armorIPs := []string{
		"1.1.1.1/32",
		"3.3.3.3/32",
		"4.4.4.4/32",
	}

	expected := []string{
		"2.2.2.2/32",
	}

	result := detectWhichOfTheseIPsAreNotBlocked(elasticIPs, armorIPs)

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Error actual = %v, and Expected = %v.", result, expected)
	}

}