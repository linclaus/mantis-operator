package kubernetes

import (
	"context"
	"fmt"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) MakeBasicRule(ns, name string, groups []monitoringv1.RuleGroup) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]string{
				"role": "rulefile",
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
