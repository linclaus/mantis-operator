package model

type StrategyMetricRequest struct {
	StrategyId string `json:strategyId`
	Container  string `json:container`
	Keyword    string `json:keyword`
}
