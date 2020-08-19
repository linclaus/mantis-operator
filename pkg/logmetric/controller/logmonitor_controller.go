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
	"context"
	"fmt"

	"github.com/linclaus/mantis-opeartor/pkg/logmetric/metric"

	"github.com/linclaus/mantis-opeartor/pkg/logmetric/model"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	logmonitorv1 "github.com/linclaus/mantis-opeartor/api/v1"
)

// LogMonitorReconciler reconciles a LogMonitor object
type LogMonitorReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	ElasticMetricMap *model.ElasticMetricMap
}

// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=logmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=logmonitors/status,verbs=get;update;patch

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
		r.CreateOrUpdateCRD(ns, strategyId, lms)

	}
	return ctrl.Result{}, nil
}

func (r *LogMonitorReconciler) CreateOrUpdateCRD(namespace, strategyId string, lm *logmonitorv1.LogMonitor) error {
	fmt.Printf("logmonitor: %s\n", lm.Spec)
	//create ElasticMetric
	sm := &model.StrategyMetric{
		StrategyId: strategyId,
		Dsl:        lm.Spec.Dsl,
	}
	r.ElasticMetricMap.Set(sm.StrategyId, sm)
	return nil
}

func (r *LogMonitorReconciler) DeleteCRD(namespace, strategyId string) error {
	//delete ElasticMetric
	em := r.ElasticMetricMap.Get(strategyId)
	if em != nil {
		metric.ElasticMetricCountVec.DeleteLabelValues(em.StrategyId)
		r.ElasticMetricMap.Delete(em.StrategyId)
	}
	return nil
}

func (r *LogMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&logmonitorv1.LogMonitor{}).
		Complete(r)
}
