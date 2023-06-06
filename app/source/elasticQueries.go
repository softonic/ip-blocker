package source

import (
	"bytes"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v7"
	"k8s.io/klog"
)

func elasticSearchQuery(client *elasticsearch.Client, index string, query *bytes.Reader) (map[string]interface{}, error) {

	res, err := client.Search(
		client.Search.WithIndex(index),
		client.Search.WithBody(query),
		client.Search.WithTrackTotalHits(true),
		client.Search.WithPretty(),
		client.Search.WithSize(10000),
	)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			klog.Fatalf("Error parsing the respnse body: %s", err)
		} else {
			klog.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)

		}
	}

	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		klog.Fatalf("Error parsing the response body: %s", err)
	}
	klog.Infof(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(r["took"].(float64)),
	)

	//r = res.Body

	return r, nil

}
