package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ElasticMetricCountVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "elastic_metric_gauge_vec",
		Help: "elastic count",
	}, []string{"strategy_id"})
)

//Init metric
func init() {
	prometheus.MustRegister(ElasticMetricCountVec)
}
