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

	"github.com/linclaus/mantis-opeartor/pkg/alertmanager"

	"github.com/linclaus/mantis-opeartor/pkg/prometheus"

	"github.com/linclaus/mantis-opeartor/pkg/conf"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	logmonitorv1 "github.com/linclaus/mantis-opeartor/api/v1"
	"gopkg.in/yaml.v2"
)

// LogMonitorReconciler reconciles a LogMonitor object
type LogMonitorReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Framework *prometheus.Framework
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
	secret, _ := r.Framework.GetSecret(conf.PROMETHEUS_NAMESPACE, conf.ALERTMANAGER_SECRET_NAME)
	if secret != nil {
		b := secret.Data[conf.ALERTMANAGER_SECRET_DATA_NAME]
		cfg := &alertmanager.Config{}
		err := yaml.Unmarshal(b, cfg)
		if err != nil {
			fmt.Printf("load alertmanager config failed: %s", err)
			return err
		}

		cfg.Receivers = alertmanager.UpdatedReceivers(cfg.Receivers, strategyId, lm)
		cfg.Route.Routes = alertmanager.UpdatedRoutes(cfg.Route.Routes, strategyId, lm)
		fmt.Println(cfg)
		secret.Data[conf.ALERTMANAGER_SECRET_DATA_NAME] = []byte(cfg.String())
		r.Framework.UpdateSecret(conf.PROMETHEUS_NAMESPACE, secret)
	}
	return nil
}

func (r *LogMonitorReconciler) DeleteCRD(namespace, strategyId string) error {
	//delete prometheusRule
	r.Framework.DeleteRule(namespace, strategyId)

	//update Alertmanager secret
	secret, _ := r.Framework.GetSecret(conf.PROMETHEUS_NAMESPACE, conf.ALERTMANAGER_SECRET_NAME)
	if secret != nil {
		b := secret.Data[conf.ALERTMANAGER_SECRET_DATA_NAME]
		cfg, _ := alertmanager.Load(string(b))

		cfg.Receivers = alertmanager.DeletedReceivers(cfg.Receivers, strategyId)
		cfg.Route.Routes = alertmanager.DeletedRoutes(cfg.Route.Routes, strategyId)
		fmt.Println(cfg)
		secret.Data[conf.ALERTMANAGER_SECRET_DATA_NAME] = []byte(cfg.String())
		r.Framework.UpdateSecret(conf.PROMETHEUS_NAMESPACE, secret)
	}
	return nil
}

func (r *LogMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&logmonitorv1.LogMonitor{}).
		Complete(r)
}
