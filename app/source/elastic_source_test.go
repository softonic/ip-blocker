package source

import (
	"reflect"
	"testing"

	"github.com/softonic/ip-blocker/app"
	elasticUtils "github.com/softonic/ip-blocker/app/source/utils"
)

func TestOrderAndTrimIPsMostCounts(t *testing.T) {

	threshold := 25

	ipCounter := map[string]int{
		"1.1.1.1": 1,
		"2.2.2.2": 30,
	}

	result := elasticUtils.OrderAndTrimIPs(ipCounter, threshold)

	expectedBots := []app.IPCount{
		{
			IP:    "2.2.2.2",
			Count: 30,
		},
	}

	if !reflect.DeepEqual(expectedBots, result) {
		t.Errorf("Error actual = %v, and Expected = %v.", result, expectedBots)
	}

}

func TestOrderAndTrimIPsNumberIPsToBlock(t *testing.T) {

	threshold := 25

	ipCounter := map[string]int{
		"1.1.1.1":     42,
		"2.2.2.2":     41,
		"3.3.3.3":     40,
		"4.4.4.4":     39,
		"5.5.5.5":     38,
		"6.6.6.6":     37,
		"7.7.7.7":     36,
		"8.8.8.8":     35,
		"9.9.9.9":     34,
		"10.10.10.10": 33,
		"11.11.11.11": 32,
		"12.12.12.12": 31,
	}

	result := elasticUtils.OrderAndTrimIPs(ipCounter, threshold)

	expectedBots := []app.IPCount{
		{
			IP:    "1.1.1.1",
			Count: 42,
		},
		{
			IP:    "2.2.2.2",
			Count: 41,
		},
		{
			IP:    "3.3.3.3",
			Count: 40,
		},
		{
			IP:    "4.4.4.4",
			Count: 39,
		},
		{
			IP:    "5.5.5.5",
			Count: 38,
		},
		{
			IP:    "6.6.6.6",
			Count: 37,
		},
		{
			IP:    "7.7.7.7",
			Count: 36,
		},
		{
			IP:    "8.8.8.8",
			Count: 35,
		},
		{
			IP:    "9.9.9.9",
			Count: 34,
		},
		{
			IP:    "10.10.10.10",
			Count: 33,
		},
	}

	if !reflect.DeepEqual(expectedBots, result) {
		t.Errorf("Error actual = %v, and Expected = %v.", result, expectedBots)
	}

}
