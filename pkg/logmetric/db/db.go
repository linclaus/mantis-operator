package db

import "gitlab.moebius.com/mantis/pkg/logmetric/model"

var (
	dateTemplate      = "2006-01-02T15:04:05"
	indexDateTemplate = "2006.01.02"
	indexPrefix       = "filebeat-6.8.3-"
)

type Storer interface {
	GetVersion() error
	GetMetric(sm model.StrategyMetric) model.ElasticMetric
}
