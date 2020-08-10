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

	"github.com/linclaus/mantis-opeartor/pkg/model"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	logmonitorv1 "github.com/linclaus/mantis-opeartor/api/v1"
)

// LogMonitorSumReconciler reconciles a LogMonitorSum object
type LogMonitorSumReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	ElasticMetricMap *model.ElasticMetricMap
}

// +kubebuilder:rbac:groups=logmonitor.monitoring.coreos.com,resources=logmonitorsums,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=logmonitor.monitoring.coreos.com,resources=logmonitorsums/status,verbs=get;update;patch

func (r *LogMonitorSumReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("logmonitorsum", req.NamespacedName)

	lms := &logmonitorv1.LogMonitorSum{}
	strategyId := req.NamespacedName.Name
	if err := r.Get(ctx, req.NamespacedName, lms); err != nil {
		r.Log.V(1).Info("Deleted LogMonitorSum")
		r.DeleteCRD(strategyId)
	} else {
		r.Log.V(1).Info("Successfully get LogMonitorSum", "LogMonitorSum", lms.Spec)
		r.CreateOrUpdateCRD(strategyId, lms)
		if lms.Status.Created {
			lms.Status.Created = false
		} else {
			lms.Status.Created = true
		}
		// r.Status().Update(ctx, lms)
		r.Update(ctx, lms)
	}

	return ctrl.Result{}, nil
}

func (r *LogMonitorSumReconciler) CreateOrUpdateCRD(strategyId string, lm *logmonitorv1.LogMonitorSum) {
	sm := &model.StrategyMetric{
		StrategyId: strategyId,
		Container:  lm.Spec.Labels.ContainerName,
		Keyword:    lm.Spec.Keyword,
	}
	r.ElasticMetricMap.Set(sm.StrategyId, sm)
}

func (r *LogMonitorSumReconciler) DeleteCRD(strategyId string) {
	r.ElasticMetricMap.Delete(strategyId)
}

func (r *LogMonitorSumReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&logmonitorv1.LogMonitorSum{}).
		Complete(r)
}
