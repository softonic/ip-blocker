package source

import (
	"bytes"
	"fmt"
	"time"

	"k8s.io/klog"

	"github.com/softonic/ip-blocker/app"

	elasticUtils "github.com/softonic/ip-blocker/app/source/utils"
)

var (
	r map[string]interface{}
)

type ElasticSource struct {
	connection   *ElasticConnection  // client to connect to ElasticSearch
	queryConfig  *ElasticQueryConfig // queries to execute in ElasticSearch
	sourceConfig *SourceConfig       // source configuration (address, username, password, threshold)
}

// NewElasticSource returns a ElasticSource object
func NewElasticSource(config *SourceConfig) (*ElasticSource, error) {

	connection, err := NewElasticConnection(config.Address, config.Username, config.Password, config.Threshold, config.CACert)
	if err != nil {
		return nil, err
	}

	queryConfig := &ElasticQueryConfig{}
	err = queryConfig.LoadConfig()
	if err != nil {
		return nil, err
	}

	return &ElasticSource{
		connection:   connection,
		queryConfig:  queryConfig,
		sourceConfig: config,
	}, nil
}

// getElasticIndex returns the index name for the current day
func getElasticIndex(basename string) string {

	now := time.Now()

	suffix := fmt.Sprintf("%d.%02d.%02d",
		now.Year(), now.Month(), now.Day())

	index := basename + "-" + suffix

	return index

}

// runQueryAndGetIPs ejecuta la consulta a ElasticSearch y retorna un mapa con la cuenta de las IPs.
func (s *ElasticSource) runQueryAndGetIPs(indexName string, query *bytes.Reader, fieldToSearch string) (map[string]int, error) {
	ipCounter := make(map[string]int)

	r, err := elasticSearchQuery(s.connection.client, indexName, query)
	if err != nil {
		return nil, fmt.Errorf("error executing ElasticSearch query: %w", err)
	}

	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		ips := hit.(map[string]interface{})["_source"].(map[string]interface{})[fieldToSearch].(map[string]interface{})["ip"].(string)
		ipCounter[ips]++
	}

	return ipCounter, nil
}

// GetIPCount is a bussiness function to get the IPs from ElasticSearch with its count
func (s *ElasticSource) GetIPCount(interval int) app.Result {
	var listIPCandidates []app.IPCount
	var resultError error

	// loop over all queries
	for _, query := range s.queryConfig.Queries {
		todayIndexName := getElasticIndex(query.ElasticIndex)
		read := elasticUtils.GetQueryFromFile(query.QueryFile)

		ipCounter, err := s.runQueryAndGetIPs(todayIndexName, read, query.ElasticFieldtoSearch)
		if err != nil {
			resultError = fmt.Errorf("failed to run query and get IPs: %w", err)
			break
		}

		klog.Infof("This is the ipcounter: %v", ipCounter)

		maxCounter := interval * s.connection.threshold
		klog.Infof("This is the counter: %d", maxCounter)

		listIPs := elasticUtils.OrderAndTrimIPs(ipCounter, maxCounter)
		klog.Infof("These are the listIPs after orderAndTrim: %v", listIPs)

		listIPCandidates = append(listIPCandidates, listIPs...)
	}

	return app.Result{
		IPCounts: listIPCandidates,
		Error:    resultError,
	}
}
