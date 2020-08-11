package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	apiclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monitoringclient "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	admissionHookSecretName                 = "prometheus-operator-admission"
	prometheusOperatorServiceDeploymentName = "prometheus-operator"
	operatorTLSDir                          = "/etc/tls/private"
)

type Framework struct {
	KubeClient      kubernetes.Interface
	MonClientV1     monitoringclient.MonitoringV1Interface
	APIServerClient apiclient.Interface
	HTTPClient      *http.Client
	MasterHost      string
	DefaultTimeout  time.Duration
}

// New setups a test framework and returns it.
func New(kubeconfigPath, masterUrl string) (*Framework, error) {
	config, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "build config from flags failed")
	}

	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating new kube-client failed")
	}

	apiCli, err := apiclient.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating new kube-client failed")
	}

	httpc := cli.CoreV1().RESTClient().(*rest.RESTClient).Client
	if err != nil {
		return nil, errors.Wrap(err, "creating http-client failed")
	}

	mClientV1, err := monitoringclient.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "creating v1 monitoring client failed")
	}

	f := &Framework{
		MasterHost:      config.Host,
		KubeClient:      cli,
		MonClientV1:     mClientV1,
		APIServerClient: apiCli,
		HTTPClient:      httpc,
		DefaultTimeout:  time.Minute,
	}

	return f, nil
}

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

func (f *Framework) MakeAndCreateFiringRule(ns, name, alertName string) (*monitoringv1.PrometheusRule, error) {
	groups := []monitoringv1.RuleGroup{
		{
			Name: alertName,
			Rules: []monitoringv1.Rule{
				{
					Alert: alertName,
					Expr:  intstr.FromString("vector(1)"),
				},
			},
		},
	}
	file := f.MakeBasicRule(ns, name, groups)

	result, err := f.CreateRule(ns, file)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (f *Framework) MakeAndCreateInvalidRule(ns, name, alertName string) (*monitoringv1.PrometheusRule, error) {
	groups := []monitoringv1.RuleGroup{
		{
			Name: alertName,
			Rules: []monitoringv1.Rule{
				{
					Alert: alertName,
					Expr:  intstr.FromString("vector(1))"),
				},
			},
		},
	}
	file := f.MakeBasicRule(ns, name, groups)

	result, err := f.CreateRule(ns, file)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// WaitForRule waits for a rule file with a given name to exist in a given
// namespace.
func (f *Framework) WaitForRule(ns, name string) error {
	return wait.Poll(time.Second, f.DefaultTimeout, func() (bool, error) {
		_, err := f.MonClientV1.PrometheusRules(ns).Get(context.TODO(), name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}
		return true, nil
	})
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
