package source

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/softonic/ip-blocker/app"
	"k8s.io/klog"
)

type Response struct {
	Ip       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
	Readme   string `json:"readme"`
}

func transformCSVtoArray() ([]string, error) {

	var records []string

	csvfile, err := os.Open("/etc/config/input-csv-rules-cloud-providers.csv")
	if err != nil {
		klog.Info("Couldn't open the csv file", err)
		return nil, err

	}

	r := csv.NewReader(csvfile)

	for {
		// Read each record from csv
		records, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			klog.Info(err)
		}
		fmt.Printf("Org: %s City %s Country: %s Hostname: %s\n", records[0], records[1], records[2], records[3])
	}

	return records, nil

}

func checkIPFromCloudProviders(listIPsCloudProviders []app.IPCount, records []string) []app.IPCount {

	var cResp Response
	var filteredIPCloudProviders []app.IPCount

	for _, ipCount := range listIPsCloudProviders {

		url := "https://ipinfo.io/" + ipCount.IP
		fmt.Println("URL:>", url)

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		fmt.Println("response Status:", resp.Status)

		body, _ := ioutil.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &cResp); err != nil { // unmarshall body contents as a type Candidate
			fmt.Println("Can not unmarshal JSON")
		}

		fmt.Println(cResp.Org)

		for _, entry := range records {
			// extract org from csv
			// ORG, CITY, COUNTRY, HOSTNAME
			s := strings.Split(entry, ",")

			matchOrg, _ := regexp.MatchString(s[0], cResp.Org)
			matchHostname, _ := regexp.MatchString(s[3], cResp.Hostname)
			if matchOrg && matchHostname {
				filteredIPCloudProviders = append(filteredIPCloudProviders, ipCount)
			} else {
				fmt.Println("Not found")
			}

		}

	}

	return filteredIPCloudProviders
}
