package db

import "github.com/linclaus/mantis-opeartor/pkg/logmetric/model"

var (
	dateTemplate      = "2006-01-02T15:04:05"
	indexDateTemplate = "2006.01.02"
	indexPrefix       = "filebeat-6.8.3-"
)

type Storer interface {
	GetVersion() error
	GetMetric(sm model.StrategyMetric) model.ElasticMetric
}
