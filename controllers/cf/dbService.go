package cf

import (
	"github.com/vikrakum/mernDeploymentOperator/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func DbService(nameSpace string) *corev1.Service {
	dbService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.DB_DEPLOYMENT_NAME + "-service",
			Namespace: nameSpace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": common.DB_APP_NAME,
			},
			Ports: []corev1.ServicePort{
				{
					Protocol: corev1.ProtocolTCP,
					Port:     27017,
					TargetPort: intstr.IntOrString{
						IntVal: 27017,
					},
				},
			},
		},
	}
	return dbService
}
