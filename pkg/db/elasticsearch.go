package db

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v6/esapi"
	"github.com/linclaus/mantis-opeartor/pkg/model"

	es6 "github.com/elastic/go-elasticsearch/v6"
)

type ElasticDB struct {
	esClient *es6.Client
}

func ConnectES(addresses []string) (*ElasticDB, error) {
	cfg := es6.Config{
		Addresses: addresses,
	}
	client, err := es6.NewClient(cfg)
	if err != nil {
		return nil, err
	} else {
		return &ElasticDB{esClient: client}, nil
	}
}

func (es ElasticDB) GetVersion() error {
	res, err := es.esClient.Info()
	if err != nil {
		return err
	}
	defer res.Body.Close()
	log.Println(res)
	return nil
}

func (es ElasticDB) GetMetric(sm model.StrategyMetric) model.ElasticMetric {
	count := es.countByKeyword(sm.Container, sm.Keyword)
	log.Printf("count : %f", count)
	em := model.ElasticMetric{
		Keyword:    sm.Keyword,
		StrategyId: sm.StrategyId,
		Count:      count,
	}
	return em
}

func (es ElasticDB) countByKeyword(container string, keyword string) float64 {
	now := time.Now().UTC()
	from := now.Add(-1 * time.Duration(1)).UTC()
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"kubernetes.container.name": container,
						}},
					map[string]interface{}{"match_phrase": map[string]interface{}{
						"message": keyword,
					}},
					map[string]interface{}{"range": map[string]interface{}{
						"@timestamp": map[string]interface{}{
							"gt": from.Format(dateTemplate),
							"lt": now.Format(dateTemplate),
						}},
					},
				},
			},
		},
	}
	jsonBody, _ := json.Marshal(query)

	log.Printf("jsonBody: %s", jsonBody)

	req := esapi.CountRequest{
		Index:        []string{strings.Join([]string{indexPrefix, from.Format(indexDateTemplate)}, ""), strings.Join([]string{indexPrefix, now.Format(indexDateTemplate)}, "")},
		DocumentType: []string{"doc"},
		Body:         bytes.NewReader(jsonBody),
	}
	res, err := req.Do(context.Background(), es.esClient)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	var jsonData map[string]interface{}
	if res.StatusCode == 200 {
		jsonResp, _ := ioutil.ReadAll(res.Body)
		json.Unmarshal([]byte(jsonResp), &jsonData)
		count, _ := jsonData["count"].(float64)
		return count
	}
	log.Println(res.String())
	return 0
}
