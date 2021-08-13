package robustness_test

import (
	. "github.com/onsi/ginkgo"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/test/integration/robustness/resources/istio"
	skv1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var _ = Describe("Robustness", func() {
	It("deterministically translates inputs into expected outputs", func() {
		testCase{states: []testState{
			{
				description: "mesh detected when istiod present",
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
							// discovered east mesh
							&v1.Mesh{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "istiod-istio-namespace-remote-east",
									Namespace: "gloo-mesh",
									Labels: map[string]string{
										"cluster.discovery.mesh.gloo.solo.io": remoteEast,
										"cluster.multicluster.solo.io":        "",
										"owner.discovery.mesh.gloo.solo.io":   "gloo-mesh",
									},
								},
								Spec: v1.MeshSpec{
									Type: &v1.MeshSpec_Istio_{Istio: &v1.MeshSpec_Istio{
										Installation: &v1.MeshInstallation{
											Namespace: istio.IstioNamespace,
											Cluster:   remoteEast,
											PodLabels: istio.IstiodLabels,
											Version:   "latest",
											Region:    "",
										},
										TrustDomain:          istio.IstioTrustDomain,
										IstiodServiceAccount: istio.IstioServiceAccount,
										IngressGateways: []*v1.MeshSpec_Istio_IngressGatewayInfo{{
											Name:                "istio-ingress",
											Namespace:           istio.IstioNamespace,
											WorkloadLabels:      map[string]string{"istio": "ingressgateway"},
											ExternalAddress:     "12.34.56.78",
											ExternalAddressType: &v1.MeshSpec_Istio_IngressGatewayInfo_Ip{Ip: "12.34.56.78"},
											ExternalTlsPort:     1234,
											TlsContainerPort:    1234,
										}},
										SmartDnsProxyingEnabled: false,
									}},
								},
							},
							// discovered productpage workload
							&v1.Workload{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: "gloo-mesh",
									Name:      "productpage-bookinfo-remote-east-deployment",
									Labels: map[string]string{
										"owner.discovery.mesh.gloo.solo.io":   "gloo-mesh",
										"cluster.discovery.mesh.gloo.solo.io": "remote-east",
										"cluster.multicluster.solo.io":        "",
									},
								},
								Spec: v1.WorkloadSpec{
									Type: &v1.WorkloadSpec_Kubernetes{
										Kubernetes: &v1.WorkloadSpec_KubernetesWorkload{
											Controller: &skv1.ClusterObjectRef{
												Name:        istio.ProductpageName,
												Namespace:   istio.BookinfoNamespace,
												ClusterName: remoteEast,
											},
											PodLabels:          istio.ProductpageLabels,
											ServiceAccountName: istio.ProductpageName,
										},
									},
									Mesh: &skv1.ObjectRef{
										Name:      "istiod-istio-namespace-remote-east",
										Namespace: "gloo-mesh",
									},
								},
							},
							// discovered productpage destination
							&v1.Destination{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: "gloo-mesh",
									Name:      "productpage-bookinfo-remote-east",
									Labels: map[string]string{
										"owner.discovery.mesh.gloo.solo.io":   "gloo-mesh",
										"cluster.discovery.mesh.gloo.solo.io": "remote-east",
										"cluster.multicluster.solo.io":        "",
									},
								},
								Spec: v1.DestinationSpec{
									Type: &v1.DestinationSpec_KubeService_{
										KubeService: &v1.DestinationSpec_KubeService{
											Ref: &skv1.ClusterObjectRef{
												Name:        istio.ProductpageName,
												Namespace:   istio.BookinfoNamespace,
												ClusterName: remoteEast,
											},
											WorkloadSelectorLabels: istio.ProductpageLabels,
											Ports: []*v1.DestinationSpec_KubeService_KubeServicePort{{
												Port:       9080,
												Name:       "http",
												Protocol:   "TCP",
												TargetPort: &v1.DestinationSpec_KubeService_KubeServicePort_TargetPortNumber{TargetPortNumber: 9080},
											}},
											ServiceType: v1.DestinationSpec_KubeService_CLUSTER_IP,
										},
									},
									Mesh: &skv1.ObjectRef{
										Name:      "istiod-istio-namespace-remote-east",
										Namespace: "gloo-mesh",
									},
								},
							},

							// discovered west mesh
							&v1.Mesh{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "istiod-istio-namespace-remote-west",
									Namespace: "gloo-mesh",
									Labels: map[string]string{
										"cluster.discovery.mesh.gloo.solo.io": remoteWest,
										"cluster.multicluster.solo.io":        "",
										"owner.discovery.mesh.gloo.solo.io":   "gloo-mesh",
									},
								},
								Spec: v1.MeshSpec{
									Type: &v1.MeshSpec_Istio_{Istio: &v1.MeshSpec_Istio{
										Installation: &v1.MeshInstallation{
											Namespace: istio.IstioNamespace,
											Cluster:   remoteWest,
											PodLabels: istio.IstiodLabels,
											Version:   "latest",
											Region:    "",
										},
										TrustDomain:          istio.IstioTrustDomain,
										IstiodServiceAccount: istio.IstioServiceAccount,
										IngressGateways: []*v1.MeshSpec_Istio_IngressGatewayInfo{{
											Name:                "istio-ingress",
											Namespace:           istio.IstioNamespace,
											WorkloadLabels:      map[string]string{"istio": "ingressgateway"},
											ExternalAddress:     "12.34.56.78",
											ExternalAddressType: &v1.MeshSpec_Istio_IngressGatewayInfo_Ip{Ip: "12.34.56.78"},
											ExternalTlsPort:     1234,
											TlsContainerPort:    1234,
										}},
										SmartDnsProxyingEnabled: false,
									}},
								},
							},
							// discovered productpage workload
							&v1.Workload{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: "gloo-mesh",
									Name:      "productpage-bookinfo-remote-west-deployment",
									Labels: map[string]string{
										"owner.discovery.mesh.gloo.solo.io":   "gloo-mesh",
										"cluster.discovery.mesh.gloo.solo.io": "remote-west",
										"cluster.multicluster.solo.io":        "",
									},
								},
								Spec: v1.WorkloadSpec{
									Type: &v1.WorkloadSpec_Kubernetes{
										Kubernetes: &v1.WorkloadSpec_KubernetesWorkload{
											Controller: &skv1.ClusterObjectRef{
												Name:        istio.ProductpageName,
												Namespace:   istio.BookinfoNamespace,
												ClusterName: remoteWest,
											},
											PodLabels:          istio.ProductpageLabels,
											ServiceAccountName: istio.ProductpageName,
										},
									},
									Mesh: &skv1.ObjectRef{
										Name:      "istiod-istio-namespace-remote-west",
										Namespace: "gloo-mesh",
									},
								},
							},
							// discovered productpage destination
							&v1.Destination{
								ObjectMeta: metav1.ObjectMeta{
									Namespace: "gloo-mesh",
									Name:      "productpage-bookinfo-remote-west",
									Labels: map[string]string{
										"owner.discovery.mesh.gloo.solo.io":   "gloo-mesh",
										"cluster.discovery.mesh.gloo.solo.io": "remote-west",
										"cluster.multicluster.solo.io":        "",
									},
								},
								Spec: v1.DestinationSpec{
									Type: &v1.DestinationSpec_KubeService_{
										KubeService: &v1.DestinationSpec_KubeService{
											Ref: &skv1.ClusterObjectRef{
												Name:        istio.ProductpageName,
												Namespace:   istio.BookinfoNamespace,
												ClusterName: remoteWest,
											},
											WorkloadSelectorLabels: istio.ProductpageLabels,
											Ports: []*v1.DestinationSpec_KubeService_KubeServicePort{{
												Port:       9080,
												Name:       "http",
												Protocol:   "TCP",
												TargetPort: &v1.DestinationSpec_KubeService_KubeServicePort_TargetPortNumber{TargetPortNumber: 9080},
											}},
											ServiceType: v1.DestinationSpec_KubeService_CLUSTER_IP,
										},
									},
									Mesh: &skv1.ObjectRef{
										Name:      "istiod-istio-namespace-remote-west",
										Namespace: "gloo-mesh",
									},
								},
							},
						},
					},
					remoteEastMgr: {
						clusterInputs: []client.Object{
							// istio
							istio.IstioNamespaceObj,
							istio.IstiodDeploymentObj,
							istio.IstioMeshConfigConfigMapObj,
							istio.IstioIngressGatewayServiceObj,
							// bookinfo
							istio.BookinfoNamespaceObj,
							istio.ProductpageDeploymentObj,
							istio.ProductpageServiceObj,
						},
						clusterExpectedOutputs: nil,
					},
					remoteWestMgr: {
						clusterInputs: []client.Object{
							istio.IstioNamespaceObj,
							istio.IstiodDeploymentObj,
							istio.IstioMeshConfigConfigMapObj,
							istio.IstioIngressGatewayServiceObj,
							// bookinfo
							istio.BookinfoNamespaceObj,
							istio.ProductpageDeploymentObj,
							istio.ProductpageServiceObj,
						},
						clusterExpectedOutputs: nil,
					},
				},
			},
		}}.execute(rootCtx)
	})
})
