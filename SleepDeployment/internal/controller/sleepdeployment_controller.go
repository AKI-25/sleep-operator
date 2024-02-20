/*
Copyright 2024 Ismailabdelkefi.

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

package controller

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	// v1 "k8s.io/client-go/applyconfigurations/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	demov1 "wecraft-operator/api/v1"
)

// SleepDeploymentReconciler reconciles a SleepDeployment object
type SleepDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=demo.demo.wecraft.tn,resources=sleepdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=demo.demo.wecraft.tn,resources=sleepdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=demo.demo.wecraft.tn,resources=sleepdeployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SleepDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *SleepDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Enter Reconcile", "Req", req)
	sleepDeployment := &demov1.SleepDeployment{}
	err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, sleepDeployment)
	l.Info(sleepDeployment.Spec.EndSleep)
	// l.Info("SleepAt: ", sleepDeployment.Spec.StartSleep)
	// l.Info("EndSleepAt: ", sleepDeployment.Spec.EndSleep)
	if sleepDeployment.Spec.StartSleep != sleepDeployment.Status.StartSleep {
		sleepDeployment.Status.StartSleep = sleepDeployment.Spec.StartSleep
		r.Status().Update(ctx, sleepDeployment)
		return ctrl.Result{}, err
	}
	err = r.reconcileDeploy(ctx, sleepDeployment, l)
	if err != nil {
		l.Info("unable to reconcile Deployment", err)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *SleepDeploymentReconciler) reconcileDeploy(ctx context.Context, sleepDeployment *demov1.SleepDeployment, l logr.Logger) error {
	deploy := &appsv1.Deployment{}
	l.Info(sleepDeployment.Spec.DeploymentRef)
	err := r.Get(ctx, types.NamespacedName{Name: sleepDeployment.Spec.DeploymentRef, Namespace: sleepDeployment.Namespace}, deploy)
	if err == nil {
		replicas := int32(0)
		def := int32(0)
		if deploy.Spec.Replicas != nil {
			replicas = *deploy.Spec.Replicas
		}
		l.Info("Deployment Found", "replicas", replicas)
		l.Info("Making deployment go to sleep...")
		deploy.Spec.Replicas = &def
		r.Update(ctx, deploy)
		l.Info("Deployment went to sleep...")
		return nil
	}
	if !errors.IsNotFound(err) {
		return err
	}
	l.Info("Deployment Not Found")

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SleepDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demov1.SleepDeployment{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
