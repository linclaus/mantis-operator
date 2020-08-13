package kubernetes

import (
	"net/url"

	logmonitorv1 "github.com/linclaus/mantis-opeartor/api/v1"
	alertmangerconfig "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
)

func UpdatedReceivers(rvs []*alertmangerconfig.Receiver, strategyId string, lm *logmonitorv1.LogMonitor) []*alertmangerconfig.Receiver {
	//TODO multiEmailConfig if multiContactValues
	var rv *alertmangerconfig.Receiver
	index := -1
	for i, receive := range rvs {
		if receive.Name == strategyId {
			index = i
			break
		}
	}
	ec := &alertmangerconfig.EmailConfig{
		To:      lm.Spec.ContactValue,
		HTML:    "{{ template \"" + "email-alert-content" + "\" . }}",
		Headers: map[string]string{"subject": "{{ template \"" + "email-alert-subject" + "\" . }}"},
		//TODO add status_webhook
	}
	rawurl, _ := url.Parse("http://mantis-api.moebius-system:8000/api/v2/webhook")
	wc := &alertmangerconfig.WebhookConfig{
		URL: &alertmangerconfig.URL{
			URL: rawurl,
		},
	}
	rv = &alertmangerconfig.Receiver{
		Name:           strategyId,
		EmailConfigs:   []*alertmangerconfig.EmailConfig{ec},
		WebhookConfigs: []*alertmangerconfig.WebhookConfig{wc},
	}
	if index == -1 {
		return append(rvs, rv)
	} else {
		rvs[index] = rv
		return rvs
	}
}

func DeletedReceivers(rvs []*alertmangerconfig.Receiver, strategyId string) []*alertmangerconfig.Receiver {
	index := -1
	for i, receive := range rvs {
		if receive.Name == strategyId {
			index = i
			break
		}
	}

	if index != -1 {
		return append(rvs[:index], rvs[index+1:]...)
	} else {
		return rvs
	}
}

func UpdatedRoutes(rts []*alertmangerconfig.Route, strategyId string, lm *logmonitorv1.LogMonitor) []*alertmangerconfig.Route {
	var rt *alertmangerconfig.Route
	index := -1
	for i, route := range rts {
		if route.Receiver == strategyId {
			index = i
			break
		}
	}
	ri, _ := model.ParseDuration("1s")
	gi, _ := model.ParseDuration(lm.Spec.Duration)
	rt = &alertmangerconfig.Route{
		Match:          map[string]string{"strategy_id": lm.Spec.Labels.StrategyId},
		Receiver:       strategyId,
		RepeatInterval: &ri,
		GroupInterval:  &gi,
		GroupByStr:     []string{"strategy_id"},
	}
	if index == -1 {
		return append(rts, rt)
	} else {
		rts[index] = rt
		return rts
	}
}

func DeletedRoutes(rts []*alertmangerconfig.Route, strategyId string) []*alertmangerconfig.Route {
	index := -1
	for i, route := range rts {
		if route.Receiver == strategyId {
			index = i
			break
		}
	}
	if index != -1 {
		return append(rts[:index], rts[index+1:]...)
	} else {
		return rts
	}
}
