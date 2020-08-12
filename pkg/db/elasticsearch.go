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
	count := es.countByDSL(sm.Dsl)
	log.Printf("count : %f", count)
	em := model.ElasticMetric{
		StrategyId: sm.StrategyId,
		Count:      count,
	}
	return em
}

func (es ElasticDB) countByDSL(dsl string) float64 {
	now := time.Now().UTC()
	indexString := now.Format(indexDateTemplate)
	from, _ := time.Parse(indexDateTemplate, indexString)
	from = from.UTC()

	log.Printf("dsl: %s", dsl)

	req := esapi.CountRequest{
		Index:        []string{strings.Join([]string{indexPrefix, now.Format(indexDateTemplate)}, "")},
		DocumentType: []string{"doc"},
		Body:         bytes.NewReader([]byte(dsl)),
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
