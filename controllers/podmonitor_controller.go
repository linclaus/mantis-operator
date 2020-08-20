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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitoringv1 "gitlab.moebius.com/mantis/api/v1"
)

// PodMonitorReconciler reconciles a PodMonitor object
type PodMonitorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=podmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=podmonitors/status,verbs=get;update;patch

func (r *PodMonitorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("podmonitor", req.NamespacedName)
	pm := &monitoringv1.PodMonitor{}

	if err := r.Get(ctx, req.NamespacedName, pm); err != nil {
		r.Log.V(1).Info("Deleted LogMonitor")
	} else {
		r.Log.V(1).Info("Successfully get LogMonitor", "LogMonitor", pm.Spec)
	}

	return ctrl.Result{}, nil
}

func (r *PodMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1.PodMonitor{}).
		Complete(r)
}
