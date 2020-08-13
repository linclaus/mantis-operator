/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/linclaus/mantis-opeartor/pkg/kubernetes"
	"github.com/linclaus/mantis-opeartor/pkg/model"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	logmonitorv1 "github.com/linclaus/mantis-opeartor/api/v1"
	alertmangerconfig "github.com/prometheus/alertmanager/config"
)

// LogMonitorReconciler reconciles a LogMonitor object
type LogMonitorReconciler struct {
	client.Client
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	ElasticMetricMap    *model.ElasticMetricMap
	HttpClient          *http.Client
	ElasticExportorAddr string
	Framework           *kubernetes.Framework
}

// +kubebuilder:rbac:groups=logmonitor.monitoring.coreos.com,resources=logmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=logmonitor.monitoring.coreos.com,resources=logmonitors/status,verbs=get;update;patch

func (r *LogMonitorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("logmonitor", req.NamespacedName)

	lms := &logmonitorv1.LogMonitor{}
	strategyId := req.NamespacedName.Name
	ns := req.Namespace
	if err := r.Get(ctx, req.NamespacedName, lms); err != nil {
		r.Log.V(1).Info("Deleted LogMonitor")
		r.DeleteCRD(ns, strategyId)
	} else {
		r.Log.V(1).Info("Successfully get LogMonitor", "LogMonitor", lms.Spec)
		err := r.CreateOrUpdateCRD(ns, strategyId, lms)
		if err == nil {
			lms.Status.Status = "Success"
			lms.Status.RetryTimes = 0
		} else {
			rty := lms.Status.RetryTimes
			if rty >= 100 {
				return ctrl.Result{}, nil
			}
			lms.Status.Status = "Failed"
			lms.Status.RetryTimes = rty + 1
		}
	}
	r.Status().Update(ctx, lms)

	return ctrl.Result{}, nil
}

func (r *LogMonitorReconciler) CreateOrUpdateCRD(namespace, strategyId string, lm *logmonitorv1.LogMonitor) error {
	fmt.Printf("logmonitor: %s\n", lm.Spec)
	//update PrometheusRule
	rule := r.Framework.MakeLogMonitorRule(namespace, strategyId, lm)
	r.Framework.DeleteRule(namespace, strategyId)
	r.Framework.CreateRule(namespace, rule)

	//update Alertmanager secret
	//TODO update secret namespace
	secret, _ := r.Framework.GetSecret("moebius-system", "alertmanager-r-prometheus-operator-alertmanager")
	if secret != nil {
		b := secret.Data["alertmanager.yaml"]
		cfg, _ := alertmangerconfig.Load(string(b))

		cfg.Receivers = kubernetes.UpdatedReceivers(cfg.Receivers, strategyId, lm)
		cfg.Route.Routes = kubernetes.UpdatedRoutes(cfg.Route.Routes, strategyId, lm)
		fmt.Println(cfg)
		secret.Data["alertmanager.yaml"] = []byte(cfg.String())
		r.Framework.UpdateSecret("moebius-system", secret)
	}

	//create ElasticMetric
	sm := &model.StrategyMetric{
		StrategyId: strategyId,
		Dsl:        lm.Spec.Dsl,
	}
	r.ElasticMetricMap.Set(sm.StrategyId, sm)
	data := make(map[string]interface{})
	data["dsl"] = lm.Spec.Dsl
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/metric/%s", r.ElasticExportorAddr, sm.StrategyId), bytes.NewReader(jsonData))
	r.HttpClient.Do(req)
	return nil
}

func (r *LogMonitorReconciler) DeleteCRD(namespace, strategyId string) error {
	//delete prometheusRule
	r.Framework.DeleteRule(namespace, strategyId)

	//update Alertmanager secret
	secret, _ := r.Framework.GetSecret("moebius-system", "alertmanager-r-prometheus-operator-alertmanager")
	if secret != nil {
		b := secret.Data["alertmanager.yaml"]
		cfg, _ := alertmangerconfig.Load(string(b))

		cfg.Receivers = kubernetes.DeletedReceivers(cfg.Receivers, strategyId)
		cfg.Route.Routes = kubernetes.DeletedRoutes(cfg.Route.Routes, strategyId)
		fmt.Println(cfg)
		secret.Data["alertmanager.yaml"] = []byte(cfg.String())
		r.Framework.UpdateSecret("moebius-system", secret)
	}

	//delete ElasticMetric
	r.ElasticMetricMap.Delete(strategyId)
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/metric/%s", r.ElasticExportorAddr, strategyId), nil)
	r.HttpClient.Do(req)
	return nil
}

func (r *LogMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&logmonitorv1.LogMonitor{}).
		Complete(r)
}
