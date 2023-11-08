/*
Copyright 2023.

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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	frontierv1 "github.com/yukiouma/frontapp/api/v1"
	"github.com/yukiouma/frontapp/internal/controller/template"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
)

// FrontAppReconciler reconciles a FrontApp object
type FrontAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=frontier.demo.com,resources=frontapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=frontier.demo.com,resources=frontapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=frontier.demo.com,resources=frontapps/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FrontApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *FrontAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	logger := log.FromContext(ctx)

	// TODO(user): your logic here

	// get frontapp object
	frontapp := &frontierv1.FrontApp{}
	if err = r.Get(ctx, req.NamespacedName, frontapp); err != nil {
		err = client.IgnoreNotFound(err)
		return
	}
	// handle configmap
	if err = r.createOrUpdateConfigMap(ctx, frontapp, req.NamespacedName); err != nil {
		logger.Error(err, "create or update configmap failed")
		return
	}

	// handle deployment
	if err = r.createOrUpdateDeployment(ctx, frontapp, req.NamespacedName); err != nil {
		logger.Error(err, "create or update deployment failed")
		return
	}

	// handle service
	if err = r.createOrUpdateService(ctx, frontapp, req.NamespacedName); err != nil {
		logger.Error(err, "create or update service failed")
		return
	}

	// handle ingress
	if err = r.createOrUpdateIngress(ctx, frontapp, req.NamespacedName); err != nil {
		logger.Error(err, "create or update ingress failed")
		return
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FrontAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&frontierv1.FrontApp{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&netv1.Ingress{}).
		WithEventFilter(predicate.Funcs{
			DeleteFunc: func(e event.DeleteEvent) bool {
				_, ok := e.Object.(*frontierv1.FrontApp)
				return !ok
			},
		}).
		Complete(r)
}

// handle deployment
func (r *FrontAppReconciler) createOrUpdateDeployment(
	ctx context.Context,
	frontapp *frontierv1.FrontApp,
	namespaceName types.NamespacedName,
) (err error) {
	// create deployment
	logger := log.FromContext(ctx)
	deployment, err := template.NewDeployment(frontapp)
	if err != nil {
		return
	}
	if err = controllerutil.SetControllerReference(frontapp, deployment, r.Scheme); err != nil {
		return
	}
	// find if deployment existed
	oldDeployment := &appsv1.Deployment{}
	if err = r.Get(ctx, namespaceName, oldDeployment); err != nil {
		if !errors.IsNotFound(err) {
			return
		}
		// not found, create deployment
		logger.Info("create deployment", "name", deployment.ObjectMeta.Name, "namespace", deployment.ObjectMeta.Namespace)
		return r.Create(ctx, deployment)
	}
	if oldDeployment.Spec.Template.Spec.Containers[0].Image != deployment.Spec.Template.Spec.Containers[0].Image {
		logger.Info("update deployment", "name", deployment.ObjectMeta.Name, "namespace", deployment.ObjectMeta.Namespace)
		r.Update(ctx, deployment)
	}
	return nil
}

// handle service
func (r *FrontAppReconciler) createOrUpdateService(
	ctx context.Context,
	frontapp *frontierv1.FrontApp,
	namespaceName types.NamespacedName,
) (err error) {
	logger := log.FromContext(ctx)
	// create service
	service, err := template.NewService(frontapp)
	if err != nil {
		return
	}
	if err = controllerutil.SetControllerReference(frontapp, service, r.Scheme); err != nil {
		return
	}
	// find if service existed
	oldService := &corev1.Service{}
	if err = r.Get(ctx, namespaceName, oldService); err != nil {
		if !errors.IsNotFound(err) {
			return
		}
		// not found, create service
		logger.Info("create service", "name", service.ObjectMeta.Name, "namespace", service.ObjectMeta.Namespace)
		return r.Create(ctx, service)
	}
	// no need to update service
	return nil
}

// handle ingress
func (r *FrontAppReconciler) createOrUpdateIngress(
	ctx context.Context,
	frontapp *frontierv1.FrontApp,
	namespaceName types.NamespacedName,
) (err error) {
	logger := log.FromContext(ctx)
	// create ingress
	ingress, err := template.NewIngress(frontapp)
	if err != nil {
		return
	}
	if err = controllerutil.SetControllerReference(frontapp, ingress, r.Scheme); err != nil {
		return
	}
	// find if ingress existed
	oldIngress := &netv1.Ingress{}
	if err = r.Get(ctx, namespaceName, oldIngress); err != nil {
		if !errors.IsNotFound(err) {
			return
		}
		// not found, create ingress
		logger.Info("create ingress", "name", ingress.ObjectMeta.Name, "namespace", ingress.ObjectMeta.Namespace)
		return r.Create(ctx, ingress)
	}
	if oldIngress.Spec.Rules[0].Host != ingress.Spec.Rules[0].Host {
		logger.Info("udpate ingress", "name", ingress.ObjectMeta.Name, "namespace", ingress.ObjectMeta.Namespace)
		r.Update(ctx, ingress)
	}
	return nil
}

// handle configmap
func (r *FrontAppReconciler) createOrUpdateConfigMap(
	ctx context.Context,
	frontapp *frontierv1.FrontApp,
	namespaceName types.NamespacedName,
) (err error) {
	logger := log.FromContext(ctx)
	// create configmap
	configMap, err := template.NewConfig(frontapp)
	if err != nil {
		return
	}
	if err = controllerutil.SetControllerReference(frontapp, configMap, r.Scheme); err != nil {
		return
	}
	// find if configmap existed
	oldConfigMap := &corev1.ConfigMap{}
	if err = r.Get(ctx, namespaceName, oldConfigMap); err != nil {
		if !errors.IsNotFound(err) {
			return
		}
		// not found, create config map
		logger.Info("create configMap", "name", configMap.ObjectMeta.Name, "namespace", configMap.ObjectMeta.Namespace)
		return r.Create(ctx, configMap)
	}
	if oldConfigMap.Data["Caddyfile"] != configMap.Data["Caddyfile"] {
		logger.Info("udpate ingress", "name", configMap.ObjectMeta.Name, "namespace", configMap.ObjectMeta.Namespace)
		return r.Update(ctx, configMap)
	}
	return nil
}
