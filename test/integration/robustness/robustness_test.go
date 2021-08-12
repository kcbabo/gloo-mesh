package robustness_test

import (
	v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	istiov1alpha1 "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/galley/pkg/config/analysis/analyzers/injection"
	"istio.io/istio/pkg/util/protomarshal"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Robustness", func() {

	It("upserts virtual mesh outputs to enable federation", func() {
		workloadNamespace := "bookinfo"
		istioNamespace := "istio-namespace"
		istioRevsion := "best-revision-ever"
		istioServiceAccount := "service-account-name"
		istiodDeploymentName := "istiod"
		trustDomain := "cluster.suffix"

		meshConfig := &istiov1alpha1.MeshConfig{
			TrustDomain: trustDomain,
		}
		meshConfigYaml, err := protomarshal.ToYAML(meshConfig)
		Expect(err).To(BeNil())

		testCase{states: []testState{
			{
				description: "virtual mesh outputs created when a virtual mesh is applied to multiple meshes",
				clusterStates: map[manager.Manager]configState{
					mgmtMgr: {
						clusterInputs: []client.Object{
							// gloo mesh namespace
							&corev1.Namespace{
								ObjectMeta: metav1.ObjectMeta{
									Name: defaults.GetPodNamespace(),
								},
							},
							// gloo mesh settings
							&settingsv1.Settings{
								ObjectMeta: metav1.ObjectMeta{
									Name:      defaults.DefaultSettingsName,
									Namespace: defaults.GetPodNamespace(),
								},
							},
						},
						clusterExpectedOutputs: []client.Object{
							// discovered mesh
							&v1.Mesh{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "istiod-istio-namespace-remote-east",
									Namespace: "gloo-mesh",
									Labels: map[string]string{
										"cluster.discovery.mesh.gloo.solo.io": "remote-east",
										"cluster.multicluster.solo.io":        "",
										"owner.discovery.mesh.gloo.solo.io":   "gloo-mesh",
									},
								},
								Spec: v1.MeshSpec{
									Type: &v1.MeshSpec_Istio_{Istio: &v1.MeshSpec_Istio{
										Installation:            &v1.MeshInstallation{
											Namespace: istioNamespace,
											Cluster:   "remote-east",
											PodLabels: map[string]string{
												"app": "istiod",
											},
											Version:   "latest",
											Region:    "",
										},
										TrustDomain:             trustDomain,
										IstiodServiceAccount:    istioServiceAccount,
										IngressGateways: []*v1.MeshSpec_Istio_IngressGatewayInfo{{
											Name:                "istio-ingress",
											Namespace:           istioNamespace,
											WorkloadLabels:       map[string]string{"istio": "ingressgateway"},
											ExternalAddress:     "12.34.56.78",
											ExternalAddressType: &v1.MeshSpec_Istio_IngressGatewayInfo_Ip{Ip: "12.34.56.78"},
											ExternalTlsPort:     1234,
											TlsContainerPort:    1234,
										}},
										SmartDnsProxyingEnabled: false,
									}},
								},
							},
						},
					},
					remoteEastMgr: {
						clusterInputs: []client.Object{
							// istio namespace
							&corev1.Namespace{
								ObjectMeta: metav1.ObjectMeta{
									Name: istioNamespace,
								},
							},
							// istiod deployment
							&appsv1.Deployment{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: istioNamespace,
									Name:      istiodDeploymentName,
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
											ServiceAccountName: istioServiceAccount,
										},
									},
									Selector: &metav1.LabelSelector{
										MatchLabels: map[string]string{"app": "istiod"},
									},
								},
							},
							// istio meshconfig configmap
							&corev1.ConfigMap{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "istio",
									Namespace: istioNamespace,
								},
								Data: map[string]string{
									"mesh": meshConfigYaml,
								},
							},
							// istio ingress gateway
							&corev1.Service{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "istio-ingress",
									Namespace: istioNamespace,
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
							},

							// injected workload namespace
							&corev1.Namespace{
								ObjectMeta: metav1.ObjectMeta{
									Name: workloadNamespace,
									Labels: map[string]string{
										injection.RevisionInjectionLabelName: istioRevsion,
									},
								},
							},
						},
						clusterExpectedOutputs: nil,
					},
					remoteWestMgr: {
						clusterInputs:          nil,
						clusterExpectedOutputs: nil,
					},
				},
			},
		}}.execute(rootCtx)

		Expect(err).NotTo(HaveOccurred())
		Eventually(func() (*v1.Workload, error) {
			workloads, err := v1.NewWorkloadClient(mgmtMgr.GetClient()).ListWorkload(rootCtx)
			if err != nil {
				return nil, err
			}
			if len(workloads.Items) == 0 {
				return nil, nil
			}
			return &workloads.Items[0], nil
		}, time.Minute).ShouldNot(BeNil())

		time.Sleep(time.Minute * 5)
	})
})
