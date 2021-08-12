package istio

import (
	"istio.io/istio/galley/pkg/config/analysis/analyzers/injection"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var  (
	WorkloadNamespace = "bookinfo"

	BookinfoNamespaceObj = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: WorkloadNamespace,
				Labels: map[string]string{
					injection.RevisionInjectionLabelName: IstioRevsion,
				},
			},
		}
)