package istio

import (
	"istio.io/istio/galley/pkg/config/analysis/analyzers/injection"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	BookinfoNamespace = "bookinfo"
	ProductpageName   = "productpage"
	ProductpageLabels = map[string]string{"app": ProductpageName}

	BookinfoNamespaceObj = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: BookinfoNamespace,
			Labels: map[string]string{
				injection.RevisionInjectionLabelName: IstioRevsion,
			},
		},
	}

	// productpage deployment
	ProductpageDeploymentObj = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: BookinfoNamespace,
			Name:      ProductpageName,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ProductpageLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "productpage",
							Image: "docker.io/productpage:latest",
						},
					},
					ServiceAccountName: ProductpageName,
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: ProductpageLabels,
			},
		},
	}

	// productpage service
	ProductpageServiceObj = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ProductpageName,
			Namespace: BookinfoNamespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:     "http",
				Protocol: "TCP",
				Port:     9080,
			}},
			Selector: ProductpageLabels,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
)
