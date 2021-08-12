package istio

import (
	istiov1alpha1 "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pkg/util/protomarshal"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	IstioNamespace = "istio-namespace"
	IstioRevsion = "best-revision-ever"
	IstioServiceAccount = "service-account-name"
	IstiodDeploymentName = "istiod"
	IstioTrustDomain = "cluster.suffix"


	// istio namespace
	IstioNamespaceObj = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: IstioNamespace,
		},
	}

	// istiod deployment
	IstiodDeploymentObj = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: IstioNamespace,
			Name:      IstiodDeploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "istiod"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "istiod",
							Image: "istio-pilot:latest",
						},
					},
					ServiceAccountName: IstioServiceAccount,
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "istiod"},
			},
		},
	}

	// istio meshconfig configmap
	IstioMeshConfigConfigMapObj = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio",
			Namespace: IstioNamespace,
		},
		Data: map[string]string{
			"mesh": func()string {
				meshConfig := &istiov1alpha1.MeshConfig{
					TrustDomain: IstioTrustDomain,
				}
				meshConfigYaml, err := protomarshal.ToYAML(meshConfig)
				if err != nil {
					panic(err)
				}
				return meshConfigYaml
			}(),
		},
	}

	// istio ingress gateway
	IstioIngressGatewayServiceObj = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "istio-ingress",
			Namespace: IstioNamespace,
		},
		Spec: corev1.ServiceSpec{
			ExternalIPs: []string{"12.34.56.78"},
			Ports: []corev1.ServicePort{{
				Name:     "tls",
				Protocol: "TCP",
				Port:     1234,
			}},
			Selector: map[string]string{"istio": "ingressgateway"},
			Type:     corev1.ServiceTypeLoadBalancer,
		},
	}
)
