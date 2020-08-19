package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"k8s.io/client-go/kubernetes"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	logmonitorv1 "gitlab.moebius.com/mantis/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	apiclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	monitoringclient "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
)

type Framework struct {
	KubeClient      kubernetes.Interface
	MonClientV1     monitoringclient.MonitoringV1Interface
	APIServerClient apiclient.Interface
	HTTPClient      *http.Client
	MasterHost      string
	DefaultTimeout  time.Duration
}

func (f *Framework) MakeLogMonitorRule(namespace, strategyId string, lm *logmonitorv1.LogMonitor) *monitoringv1.PrometheusRule {
	l := lm.Spec.Labels
	groups := []monitoringv1.RuleGroup{
		{
			Name: strategyId,
			Rules: []monitoringv1.Rule{
				{
					Alert: strategyId,
					Annotations: map[string]string{
						"link_prefix": l.LinkPrefix,
					},
					Expr: intstr.FromString(lm.Spec.Promql),
					For:  lm.Spec.Duration,
					Labels: map[string]string{
						"alarm_content":      l.AlarmContent,
						"alarm_source":       l.AlarmSource,
						"application":        l.Application,
						"contact":            l.Contact,
						"container_name":     l.ContainerName,
						"metric_name":        l.MetricName,
						"metric_instance_id": l.MetricInstanceId,
						"strategy_id":        l.StrategyId,
						"strategy_name":      l.StrategyName,
					},
				},
			},
		},
	}

	rule := f.MakeBasicRule(namespace, strategyId, groups)
	return rule
}

func (f *Framework) MakeBasicRule(ns, name string, groups []monitoringv1.RuleGroup) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]string{
				"app": "prometheus-operator",
			},
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: groups,
		},
	}
}

func (f *Framework) CreateRule(ns string, ar *monitoringv1.PrometheusRule) (*monitoringv1.PrometheusRule, error) {
	result, err := f.MonClientV1.PrometheusRules(ns).Create(context.TODO(), ar, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("creating %v RuleFile failed: %v", ar.Name, err)
	}

	return result, nil
}

func (f *Framework) GetRule(ns, name string) (*monitoringv1.PrometheusRule, error) {
	result, err := f.MonClientV1.PrometheusRules(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting %v Rule failed: %v", name, err)
	}

	return result, nil
}

func (f *Framework) UpdateRule(ns string, ar *monitoringv1.PrometheusRule) (*monitoringv1.PrometheusRule, error) {
	result, err := f.MonClientV1.PrometheusRules(ns).Update(context.TODO(), ar, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("updating %v RuleFile failed: %v", ar.Name, err)
	}

	return result, nil
}

func (f *Framework) DeleteRule(ns string, r string) error {
	err := f.MonClientV1.PrometheusRules(ns).Delete(context.TODO(), r, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("deleteing %v Prometheus rule in namespace %v failed: %v", r, ns, err.Error())
	}

	return nil
}
