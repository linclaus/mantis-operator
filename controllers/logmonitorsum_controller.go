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

	"github.com/linclaus/mantis-opeartor/pkg/model"
	"github.com/linclaus/mantis-opeartor/pkg/prometheus"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	logmonitorv1 "github.com/linclaus/mantis-opeartor/api/v1"
)

// LogMonitorSumReconciler reconciles a LogMonitorSum object
type LogMonitorSumReconciler struct {
	client.Client
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	ElasticMetricMap    *model.ElasticMetricMap
	HttpClient          *http.Client
	ElasticExportorAddr string
	Framework           *prometheus.Framework
}

// +kubebuilder:rbac:groups=logmonitor.monitoring.coreos.com,resources=logmonitorsums,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=logmonitor.monitoring.coreos.com,resources=logmonitorsums/status,verbs=get;update;patch

func (r *LogMonitorSumReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("logmonitorsum", req.NamespacedName)

	lms := &logmonitorv1.LogMonitorSum{}
	strategyId := req.NamespacedName.Name
	ns:=req.Namespace
	if err := r.Get(ctx, req.NamespacedName, lms); err != nil {
		r.Log.V(1).Info("Deleted LogMonitorSum")
		r.DeleteCRD(ns,strategyId)
	} else {
		r.Log.V(1).Info("Successfully get LogMonitorSum", "LogMonitorSum", lms.Spec)
		err := r.CreateOrUpdateCRD(ns,strategyId, lms)
		if err == nil {
			lms.Status.Status = "Success"
		}
	}
	if lms.Status.Status != "Success" {
		lms.Status.Status = "Running"
	}
	r.Update(ctx, lms)

	return ctrl.Result{}, nil
}

func (r *LogMonitorSumReconciler) CreateOrUpdateCRD(namespace,strategyId string, lm *logmonitorv1.LogMonitorSum) error {
	cn := lm.Spec.Labels.ContainerName
	kw := lm.Spec.Keyword
	l := lm.Spec.Labels
	//create PrometheusRule
	groups := []monitoringv1.RuleGroup{
		{
			Name: strategyId,
			Rules: []monitoringv1.Rule{
				{
					Alert: strategyId,
					Expr:  intstr.FromString("vector(1)"),
					Labels: map[string]string{
						"alarm_content":  "l.AlarmContent",
						"alarm_source":   l.AlarmSource,
						"application":    l.Application,
						"contact":        l.Contact,
						"container_name": l.ContainerName,
						"metric_name":    l.MetricName,
						"strategy_id":    l.StrategyId,
						"strategy_name":  l.StrategyName,
					},
				},
			},
		},
	}

	rule:=r.Framework.MakeBasicRule(namespace, strategyId, groups)
	r.Framework.DeleteRule(namespace,strategyId)
	r.Framework.CreateRule(namespace,rule)

	//update Alertmanager configMap

	//create ElasticMetric

	sm := &model.StrategyMetric{
		StrategyId: strategyId,
		Container:  cn,
		Keyword:    kw,
	}
	r.ElasticMetricMap.Set(sm.StrategyId, sm)
	data := make(map[string]interface{})
	data["container"] = cn
	data["keyword"] = kw
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/metric/%s", r.ElasticExportorAddr, sm.StrategyId), bytes.NewReader(jsonData))
	r.HttpClient.Do(req)
	return nil
}

func (r *LogMonitorSumReconciler) DeleteCRD(namespace,strategyId string) error {
	//delete prometheusRule
	r.Framework.DeleteRule(namespace,strategyId)

	//update Alertmanager configMap

	//delete ElasticMetric
	r.ElasticMetricMap.Delete(strategyId)
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/metric/%s", r.ElasticExportorAddr, strategyId), nil)
	r.HttpClient.Do(req)
	return nil
}

func (r *LogMonitorSumReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&logmonitorv1.LogMonitorSum{}).
		Complete(r)
}
