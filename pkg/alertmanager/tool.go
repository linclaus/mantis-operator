package alertmanager

import (
	"net/url"
	"strings"

	"gitlab.moebius.com/mantis/pkg/conf"

	logmonitorv1 "gitlab.moebius.com/mantis/api/v1"
	alertmangerconfig "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
)

func UpdatedReceivers(rvs []*Receiver, strategyId string, lm *logmonitorv1.LogMonitor) []*Receiver {
	//TODO multiEmailConfig if multiContactValues
	var rv *Receiver
	index := -1
	for i, receive := range rvs {
		if receive.Name == strategyId {
			index = i
			break
		}
	}
	cvs := strings.Split(lm.Spec.ContactValue, ",")
	ecs := []*EmailConfig{}
	for _, cv := range cvs {
		ec := &EmailConfig{}
		ec.To = cv
		ec.HTML = "{{ template \"" + "email-alert-content" + "\" . }}"
		ec.Headers = map[string]string{"subject": "{{ template \"" + "email-alert-subject" + "\" . }}"}
		ec.StatusWebhook = conf.STATUS_WEBHOOK_URL
		ecs = append(ecs, ec)
	}

	rawurl, _ := url.Parse(conf.WEBHOOK_URL)
	wc := &alertmangerconfig.WebhookConfig{
		URL: &alertmangerconfig.URL{
			URL: rawurl,
		},
	}
	rv = &Receiver{
		Name:           strategyId,
		EmailConfigs:   ecs,
		WebhookConfigs: []*alertmangerconfig.WebhookConfig{wc},
	}
	if index == -1 {
		return append(rvs, rv)
	} else {
		rvs[index] = rv
		return rvs
	}
}

func DeletedReceivers(rvs []*Receiver, strategyId string) []*Receiver {
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
