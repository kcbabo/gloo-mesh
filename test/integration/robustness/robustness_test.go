package robustness_test

import (
	. "github.com/onsi/ginkgo"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/test/integration/robustness/resources/istio"
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
										"cluster.discovery.mesh.gloo.solo.io": "remote-east",
										"cluster.multicluster.solo.io":        "",
										"owner.discovery.mesh.gloo.solo.io":   "gloo-mesh",
									},
								},
								Spec: v1.MeshSpec{
									Type: &v1.MeshSpec_Istio_{Istio: &v1.MeshSpec_Istio{
										Installation: &v1.MeshInstallation{
											Namespace: istio.IstioNamespace,
											Cluster:   "remote-east",
											PodLabels: map[string]string{
												"app": "istiod",
											},
											Version: "latest",
											Region:  "",
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
							// discovered west mesh
							&v1.Mesh{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "istiod-istio-namespace-remote-west",
									Namespace: "gloo-mesh",
									Labels: map[string]string{
										"cluster.discovery.mesh.gloo.solo.io": "remote-west",
										"cluster.multicluster.solo.io":        "",
										"owner.discovery.mesh.gloo.solo.io":   "gloo-mesh",
									},
								},
								Spec: v1.MeshSpec{
									Type: &v1.MeshSpec_Istio_{Istio: &v1.MeshSpec_Istio{
										Installation: &v1.MeshInstallation{
											Namespace: istio.IstioNamespace,
											Cluster:   "remote-west",
											PodLabels: map[string]string{
												"app": "istiod",
											},
											Version: "latest",
											Region:  "",
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
						},
					},
					remoteEastMgr: {
						clusterInputs: []client.Object{
							istio.IstioNamespaceObj,
							istio.IstiodDeploymentObj,
							istio.IstioMeshConfigConfigMapObj,
							istio.IstioIngressGatewayServiceObj,
						},
						clusterExpectedOutputs: nil,
					},
					remoteWestMgr: {
						clusterInputs: []client.Object{
							istio.IstioNamespaceObj,
							istio.IstiodDeploymentObj,
							istio.IstioMeshConfigConfigMapObj,
							istio.IstioIngressGatewayServiceObj,
						},
						clusterExpectedOutputs: nil,
					},
				},
			},
		}}.execute(rootCtx)
	})
})
