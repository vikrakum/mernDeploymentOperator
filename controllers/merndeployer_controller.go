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

package controllers

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/vikrakum/mernDeploymentOperator/api/v1alpha1"
	cachev1alpha1 "github.com/vikrakum/mernDeploymentOperator/api/v1alpha1"
	"github.com/vikrakum/mernDeploymentOperator/common"
	"github.com/vikrakum/mernDeploymentOperator/controllers/cf"
)

// MernDeployerReconciler reconciles a MernDeployer object
type MernDeployerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func dbDeployer(r *MernDeployerReconciler, ctx context.Context, req ctrl.Request) {
	log.Log.Info("DB deployment init", "in Namespace", req.Namespace, "Request.Name", req.Name)

	isDepExist := &appsv1.Deployment{}
	r.Get(ctx, types.NamespacedName{
		Name:      common.DB_DEPLOYMENT_NAME,
		Namespace: req.Namespace,
	}, isDepExist)
	// if err != nil {
	// 	return
	// }
	if isDepExist.ObjectMeta.Name == common.DB_DEPLOYMENT_NAME && *isDepExist.Spec.Replicas >= 1 {
		log.Log.Info("deployment: ", common.DB_DEPLOYMENT_NAME, " already exist")
		return
	}

	mernDeployer := &v1alpha1.MernDeployer{}
	dbReplica := new(int32)
	*dbReplica = 1
	if mernDeployer.Spec.DbReplicas > 0 {
		*dbReplica = mernDeployer.Spec.DbReplicas
	}

	mongodbDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.DB_DEPLOYMENT_NAME,
			Namespace: req.Namespace,
			Labels: map[string]string{
				"app": common.DB_APP_NAME,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: dbReplica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": common.DB_APP_NAME,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": common.DB_APP_NAME,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: common.APP_NAME + "-sa",
					Containers: []corev1.Container{{
						Image: "mongo",
						Name:  common.DB_APP_NAME,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 27017,
							Name:          common.DB_APP_NAME,
						}},
						Env: []corev1.EnvVar{{
							Name:  common.DB_ENV_USERNAME,
							Value: mernDeployer.Spec.DbSecrets.Username,
						}, {
							Name:  common.DB_ENV_PASSWORD,
							Value: mernDeployer.Spec.DbSecrets.Password,
						}},
					}},
				},
			},
		},
	}

	err := r.Create(ctx, mongodbDeployment)
	if err != nil {
		log.Log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", req.Namespace, "Deployment.Name", req.Name)
		return
	}
}

// +kubebuilder:rbac:groups=cache.my.domain,resources=merndeployers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cache.my.domain,resources=merndeployers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cache.my.domain,resources=merndeployers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MernDeployer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *MernDeployerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	appName := req.Name
	appNamespace := req.Namespace
	// deploymentsAvail := &appsv1.Deployment{}
	// // check for deployment exits or not
	// err := r.Get(ctx, types.NamespacedName{
	// 	Name:      appName,
	// 	Namespace: appNamespace,
	// }, deploymentsAvail)
	// if err != nil {
	// 	return ctrl.Result{}, err
	// }

	// create db deployment
	dbDeployer(r, ctx, req)

	// db internal services
	// isSVCexist := &corev1.Service{}
	// err := r.Get(ctx, types.NamespacedName{
	// 	Name:      common.DB_DEPLOYMENT_NAME + "-service",
	// 	Namespace: appNamespace,
	// }, isSVCexist)
	// if err != nil {
	// 	log.Log.Error(err, "fial to retrieve SVC")
	// 	return ctrl.Result{}, err
	// }
	// if isSVCexist.ObjectMeta.Name != common.DB_DEPLOYMENT_NAME+"-service" {
	// }

	dbInternalService := cf.DbService(appName, appNamespace)

	if err := r.Create(ctx, dbInternalService); err != nil {
		log.Log.Error(err, "Failed to create new service", "Deployment.OperatorNamespace", appNamespace, "Deployment.Name", appName)
		return ctrl.Result{}, err
	}

	// creating server deployment
	isDepExist := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      common.SERVER_DEPLOYMENT_NAME,
		Namespace: req.Namespace,
	}, isDepExist)
	if isDepExist.ObjectMeta.Name == common.SERVER_DEPLOYMENT_NAME && *isDepExist.Spec.Replicas >= 1 {
		log.Log.Info("deployment server: ", common.SERVER_DEPLOYMENT_NAME, " already exist")

	}
	if err != nil {
		log.Log.Error(err, "fail to retrieve mongo-express")
		serverDeployment := cf.ServerDeployer(appName, appNamespace)
		if err := r.Create(ctx, serverDeployment); err != nil {
			log.Log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", appNamespace, "Deployment.Name", common.SERVER_DEPLOYMENT_NAME)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: time.Duration(60000 * time.Second)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MernDeployerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.MernDeployer{}).
		Complete(r)
}
