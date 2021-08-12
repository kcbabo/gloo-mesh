package robustness_test

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/skv2/pkg/controllerutils"
	"github.com/solo-io/skv2/pkg/resource"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Robustness", func() {
	var (
		// upsert an obj
		upsert = func(mgr manager.Manager, obj client.Object) {
			_, err := controllerutils.Upsert(ctx, mgr.GetClient(), obj.DeepCopyObject().(client.Object))
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
		}
	)

	It("do", func() {
		gateway := makeService(
			"gateway",
			"istio",
			map[string]string{"app": "istio-ingressgateway"},
			[]corev1.Container{
				{
					Name:  "doesntmatter",
					Image: "doesntmatter",
				},
			},
			80,
		)
		istiod := makeService(
			"istiod",
			"istio",
			map[string]string{"app": "istiod"},
			[]corev1.Container{
				{
					Name:  "doesntmatter",
					Image: "istio-pilot:latest",
				},
			},
			15001,
		)
		reviews := makeService(
			"reviews",
			"bookinfo",
			map[string]string{"app": "reviews"},
			[]corev1.Container{
				{
					Name:  "doesntmatter",
					Image: "istio-proxy:latest",
				},
			},
			9080,
		)

		discoverableResources := []resource.TypedObject{
			makeNamespace("istio", nil),
			makeNamespace("bookinfo", map[string]string{
				"sidecar.istio.io/inject": "true",
			}),
			gateway.Deployment,
			gateway.Service,
			istiod.Deployment,
			istiod.Service,
			reviews.Deployment,
			reviews.Service,
		}

		for _, res := range discoverableResources {
			upsert(mgmtMgr, res)
			upsert(remoteMgr, res)
		}

		time.Sleep(time.Minute * 5)
	})
})

type deployedService struct {
	Deployment *appsv1.Deployment
	Service    *corev1.Service
}

func makeNamespace(
	name string,
	labels map[string]string,
) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func makeService(
	name, namespace string,
	labels map[string]string,
	containers []corev1.Container,
	port int32,
) *deployedService {
	return &deployedService{
		Deployment: &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    labels,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
						Labels:    labels,
					},
					Spec: corev1.PodSpec{
						Containers: containers,
					},
				},
			},
		},
		Service: &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    labels,
			},
			Spec: corev1.ServiceSpec{
				Selector: labels,
				Ports: []corev1.ServicePort{{
					Port: port,
				}},
			},
		},
	}
}
