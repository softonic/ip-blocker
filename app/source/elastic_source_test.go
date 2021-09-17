package source

import (
	"reflect"
	"testing"

	"github.com/softonic/ip-blocker/app"
)

func TestOrderAndTrimIPsMostCounts(t *testing.T) {

	ipCounter := map[string]int{
		"1.1.1.1": 1,
		"2.2.2.2": 20,
	}

	result := orderAndTrimIPs(ipCounter)

	expectedBots := []app.IPCount{
		{
			IP:    "2.2.2.2",
			Count: 20,
		},
	}

	if !reflect.DeepEqual(expectedBots, result) {
		t.Errorf("Error actual = %v, and Expected = %v.", result, expectedBots)
	}

}

func TestOrderAndTrimIPsNumberIPsToBlock(t *testing.T) {

	ipCounter := map[string]int{
		"1.1.1.1":     30,
		"2.2.2.2":     29,
		"3.3.3.3":     28,
		"4.4.4.4":     27,
		"5.5.5.5":     26,
		"6.6.6.6":     25,
		"7.7.7.7":     24,
		"8.8.8.8":     23,
		"9.9.9.9":     22,
		"10.10.10.10": 21,
		"11.11.11.11": 20,
		"12.12.12.12": 20,
	}

	result := orderAndTrimIPs(ipCounter)

	expectedBots := []app.IPCount{
		{
			IP:    "1.1.1.1",
			Count: 30,
		},
		{
			IP:    "2.2.2.2",
			Count: 29,
		},
		{
			IP:    "3.3.3.3",
			Count: 28,
		},
		{
			IP:    "4.4.4.4",
			Count: 27,
		},
		{
			IP:    "5.5.5.5",
			Count: 26,
		},
		{
			IP:    "6.6.6.6",
			Count: 25,
		},
		{
			IP:    "7.7.7.7",
			Count: 24,
		},
		{
			IP:    "8.8.8.8",
			Count: 23,
		},
		{
			IP:    "9.9.9.9",
			Count: 22,
		},
		{
			IP:    "10.10.10.10",
			Count: 21,
		},
	}

	if !reflect.DeepEqual(expectedBots, result) {
		t.Errorf("Error actual = %v, and Expected = %v.", result, expectedBots)
	}

}
