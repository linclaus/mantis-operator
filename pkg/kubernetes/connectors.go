package kubernetes

import (
	"net/http"
	"time"

	apiclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	monitoringclient "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	"github.com/pkg/errors"
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
func New(k8sMasterAddr string) (*Framework, error) {
	var err error
	var config *rest.Config
	if k8sMasterAddr == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		config = &rest.Config{
			Host: k8sMasterAddr,
		}
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
