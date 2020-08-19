package alertmanager

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v2"

	alertmangerconfig "github.com/prometheus/alertmanager/config"
	prometheuscfg "github.com/prometheus/common/config"
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

func Load(s string) (*Config, error) {
	cfg := &Config{}
	err := yaml.UnmarshalStrict([]byte(s), cfg)
	if err != nil {
		return nil, err
	}
	// Check if we have a root route. We cannot check for it in the
	// UnmarshalYAML method because it won't be called if the input is empty
	// (e.g. the config file is empty or only contains whitespace).
	if cfg.Route == nil {
		return nil, errors.New("no route provided in config")
	}

	// Check if continue in root route.
	if cfg.Route.Continue {
		return nil, errors.New("cannot have continue in root route")
	}

	cfg.original = s
	return cfg, nil
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
