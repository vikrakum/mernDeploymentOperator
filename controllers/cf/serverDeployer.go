package cf

import (
	"github.com/vikrakum/mernDeploymentOperator/api/v1alpha1"
	"github.com/vikrakum/mernDeploymentOperator/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func ServerDeployer(nameSpace string) *appsv1.Deployment {
	log.Log.Info("DB deployment init", "in Namespace", nameSpace, "Request.Name", nameSpace)

	mernDeployer := &v1alpha1.MernDeployer{}
	mongoExpressDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.SERVER_APP_NAME,
			Namespace: mernDeployer.Namespace,
			Labels: map[string]string{
				"app": common.SERVER_APP_NAME,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &mernDeployer.Spec.ServerReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": common.SERVER_APP_NAME,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": common.SERVER_APP_NAME,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: common.APP_NAME + "-sa",
					Containers: []corev1.Container{{
						Image: "mongo-express",
						Name:  common.SERVER_APP_NAME,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8081,
							Name:          common.SERVER_APP_NAME,
						}},
						Env: []corev1.EnvVar{{
							Name:  common.DB_ENV_USERNAME,
							Value: mernDeployer.Spec.DbSecrets.Username,
						}, {
							Name:  common.DB_ENV_PASSWORD,
							Value: mernDeployer.Spec.DbSecrets.Password,
						}, {
							Name:  common.DB_URL_VAR,
							Value: common.DB_DEPLOYMENT_NAME + "-service",
						}},
					}},
				},
			},
		},
	}

	return mongoExpressDeployment
}
