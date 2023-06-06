package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/softonic/ip-blocker/app"
)

// orderAndTrimIPs returns a slice of IPs ordered by count
// limit the output to the top n
// limit the output to count > threshold
func OrderAndTrimIPs(ipCounter map[string]int, maxCounter int) []app.IPCount {

	var ips []app.IPCount
	n := 10

	for ip, count := range ipCounter {
		if count > maxCounter {
			ips = append(ips, app.IPCount{
				IP:    ip,
				Count: int32(count),
			})
		}
	}

	sort.Slice(ips, func(i, j int) bool {
		return ips[i].Count > ips[j].Count
	})

	if len(ips) > n {
		ips = ips[:n]
	}

	return ips

}

func GetQueryFromFile(file string) *bytes.Reader {

	jsonFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	queryString, _ := ioutil.ReadAll(jsonFile)

	read := bytes.NewReader(queryString)

	return read

}
