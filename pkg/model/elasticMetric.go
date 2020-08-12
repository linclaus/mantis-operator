package model

import (
	"sync"
)

type ElasticMetric struct {
	StrategyId string
	Count      float64
}

type StrategyMetric struct {
	StrategyId string
	Dsl        string
}

type ElasticMetricMap struct {
	elasticMetricMap map[string]*StrategyMetric
	lock             sync.RWMutex
}

func (m ElasticMetricMap) Get(k string) *StrategyMetric {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if v, exit := m.elasticMetricMap[k]; exit {
		return v
	}
	return nil
}

func (m *ElasticMetricMap) Set(k string, v *StrategyMetric) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.elasticMetricMap == nil {
		m.elasticMetricMap = make(map[string]*StrategyMetric)
	}
	m.elasticMetricMap[k] = v
}

func (m *ElasticMetricMap) Delete(k string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.elasticMetricMap == nil {
		return
	}
	delete(m.elasticMetricMap, k)
}

func (m *ElasticMetricMap) Range(f func(k string, v *StrategyMetric) bool) {
	m.lock.RLock()
	tem := m.elasticMetricMap
	m.lock.RUnlock()
	for mk, mv := range tem {
		if f(mk, mv) {
			continue
		} else {
			break
		}
	}
}
