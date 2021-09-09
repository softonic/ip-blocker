package source

import (
	"fmt"

	"github.com/softonic/ip-blocker/app/source/elasticsearch"
)

type Source struct {
	Elasticsearch elasticsearch.ElasticSearch
}

func NewSource() *Source {
	return &Source{
		Elasticsearch: elasticsearch.ElasticSearch{},
	}
}

func (s *Source) InitConnectiontoSource(address string, username string, password string) elasticsearch.ElasticSearch {

	s.Elasticsearch = elasticsearch.ElasticSearchInit(address, username, password)

	return s.Elasticsearch

}

func (s *Source) SearchSource() map[string]int {

	botMap := elasticsearch.ElasticSearchSearch(s.Elasticsearch)

	fmt.Println("this is the map called botMap", botMap)

	return botMap

}
