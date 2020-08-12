package model

type StrategyMetricRequest struct {
	StrategyId string `json:strategyId`
	Dsl        string `json:dsl`
}
