package db

import (
	"log"

	"github.com/linclaus/mantis-opeartor/pkg/model"
)

type NullDB struct{}

func (db NullDB) GetVersion() error {
	log.Println("this is null db")
	return nil
}

func (db NullDB) GetMetric(sm model.StrategyMetric) model.ElasticMetric {
	count := 123.0
	log.Printf("count : %f", count)
	em := model.ElasticMetric{
		Keyword:    sm.Keyword,
		StrategyId: sm.StrategyId,
		Count:      count,
	}
	return em
}
