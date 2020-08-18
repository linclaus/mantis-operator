package kubernetes

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/linclaus/mantis-opeartor/pkg/conf"
	"gopkg.in/yaml.v2"

	logmonitorv1 "github.com/linclaus/mantis-opeartor/api/v1"
	alertmangerconfig "github.com/prometheus/alertmanager/config"
	prometheuscfg "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

// Config is the top-level configuration for Alertmanager's config files.
type Config struct {
	Global       *alertmangerconfig.GlobalConfig  `yaml:"global,omitempty" json:"global,omitempty"`
	Route        *alertmangerconfig.Route         `yaml:"route,omitempty" json:"route,omitempty"`
	InhibitRules []*alertmangerconfig.InhibitRule `yaml:"inhibit_rules,omitempty" json:"inhibit_rules,omitempty"`
	Receivers    []*Receiver                      `yaml:"receivers,omitempty" json:"receivers,omitempty"`
	Templates    []string                         `yaml:"templates" json:"templates"`

	// original is the input from which the config was parsed.
	original string
}

func (c Config) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("<error creating config string: %s>", err)
	}
	return string(b)
}

// Receiver configuration provides configuration on how to contact a receiver.
type Receiver struct {
	// A unique identifier for this receiver.
	Name string `yaml:"name" json:"name"`

	EmailConfigs     []*EmailConfig                       `yaml:"email_configs,omitempty" json:"email_configs,omitempty"`
	PagerdutyConfigs []*alertmangerconfig.PagerdutyConfig `yaml:"pagerduty_configs,omitempty" json:"pagerduty_configs,omitempty"`
	HipchatConfigs   []*alertmangerconfig.HipchatConfig   `yaml:"hipchat_configs,omitempty" json:"hipchat_configs,omitempty"`
	SlackConfigs     []*alertmangerconfig.SlackConfig     `yaml:"slack_configs,omitempty" json:"slack_configs,omitempty"`
	WebhookConfigs   []*alertmangerconfig.WebhookConfig   `yaml:"webhook_configs,omitempty" json:"webhook_configs,omitempty"`
	OpsGenieConfigs  []*alertmangerconfig.OpsGenieConfig  `yaml:"opsgenie_configs,omitempty" json:"opsgenie_configs,omitempty"`
	WechatConfigs    []*alertmangerconfig.WechatConfig    `yaml:"wechat_configs,omitempty" json:"wechat_configs,omitempty"`
	PushoverConfigs  []*alertmangerconfig.PushoverConfig  `yaml:"pushover_configs,omitempty" json:"pushover_configs,omitempty"`
	VictorOpsConfigs []*alertmangerconfig.VictorOpsConfig `yaml:"victorops_configs,omitempty" json:"victorops_configs,omitempty"`
}

// EmailConfig configures notifications via mail.
type EmailConfig struct {
	alertmangerconfig.NotifierConfig `yaml:",inline" json:",inline"`

	// Email address to notify.
	To            string                     `yaml:"to,omitempty" json:"to,omitempty"`
	From          string                     `yaml:"from,omitempty" json:"from,omitempty"`
	Hello         string                     `yaml:"hello,omitempty" json:"hello,omitempty"`
	Smarthost     alertmangerconfig.HostPort `yaml:"smarthost,omitempty" json:"smarthost,omitempty"`
	AuthUsername  string                     `yaml:"auth_username,omitempty" json:"auth_username,omitempty"`
	AuthPassword  alertmangerconfig.Secret   `yaml:"auth_password,omitempty" json:"auth_password,omitempty"`
	AuthSecret    alertmangerconfig.Secret   `yaml:"auth_secret,omitempty" json:"auth_secret,omitempty"`
	AuthIdentity  string                     `yaml:"auth_identity,omitempty" json:"auth_identity,omitempty"`
	Headers       map[string]string          `yaml:"headers,omitempty" json:"headers,omitempty"`
	HTML          string                     `yaml:"html,omitempty" json:"html,omitempty"`
	Text          string                     `yaml:"text,omitempty" json:"text,omitempty"`
	RequireTLS    *bool                      `yaml:"require_tls,omitempty" json:"require_tls,omitempty"`
	TLSConfig     prometheuscfg.TLSConfig    `yaml:"tls_config,omitempty" json:"tls_config,omitempty"`
	StatusWebhook string                     `yaml:"status_webhook,omitempty" json:"status_webhook,omitempty"`
}

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
