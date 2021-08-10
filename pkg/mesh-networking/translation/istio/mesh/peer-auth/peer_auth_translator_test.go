package peerAuth_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	mock_istio "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio/mocks"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	peerAuth "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/peer-auth"
	"istio.io/api/security/v1beta1"
	security_istio_io_v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PeerAuthTranslator", func() {

	var (
		ctrl *gomock.Controller
		ctx  context.Context

		mockIstioBuilder *mock_istio.MockBuilder
		mockReporter     *mock_reporting.MockReporter

		istioMesh        *discoveryv1.Mesh
		settings         *settingsv1.Settings
		expectedPeerAuth *security_istio_io_v1beta1.PeerAuthentication
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		mockIstioBuilder = mock_istio.NewMockBuilder(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)

		settings = &settingsv1.Settings{
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaults.DefaultSettingsName,
				Namespace: defaults.DefaultPodNamespace,
			},
			Spec: settingsv1.SettingsSpec{
				PeerAuth: &settingsv1.PeerAuthenticationSettings{
					Enabled:         true,
					PeerAuthTlsMode: settingsv1.PeerAuthenticationSettings_PERMISSIVE,
				},
			},
		}

		istioMesh = &discoveryv1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-istio-mesh",
			},
			Spec: discoveryv1.MeshSpec{
				Type: &discoveryv1.MeshSpec_Istio_{
					Istio: &discoveryv1.MeshSpec_Istio{
						TrustDomain:          "cluster.not-local",
						IstiodServiceAccount: "istiod-not-standard",
						Installation: &discoveryv1.MeshInstallation{
							Namespace: "istio-system-2",
							Cluster:   "cluster-name",
						},
					},
				},
				AgentInfo: &discoveryv1.MeshSpec_AgentInfo{
					AgentNamespace: "gloo-mesh",
				},
			},
		}

		// modify values as needed in each test
		expectedPeerAuth = &security_istio_io_v1beta1.PeerAuthentication{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "default",
				Namespace:   "gloo-mesh",
				ClusterName: "cluster-name",
			},
			Spec: v1beta1.PeerAuthentication{
				Mtls: &v1beta1.PeerAuthentication_MutualTLS{
					Mode: v1beta1.PeerAuthentication_MutualTLS_PERMISSIVE,
				},
			},
		}

	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will skip if non-istio mesh", func() {
		settings.Spec.PeerAuth.Enabled = false
		translator := peerAuth.NewTranslator(ctx, settings, nil)
		mesh := &discoveryv1.Mesh{}

		// will fail if any functions off the mock inputs are called, which they shouldn't be
		translator.Translate(mesh, mockIstioBuilder, mockReporter)
	})

	It("will skip if settings are disabled", func() {
		settings.Spec.PeerAuth.Enabled = false
		translator := peerAuth.NewTranslator(ctx, settings, nil)

		// will fail if any functions off the mock inputs are called, which they shouldn't be
		translator.Translate(istioMesh, mockIstioBuilder, mockReporter)
	})

	It("will use the UNSET value if no tls mode is set", func() {
		settings.Spec.PeerAuth = &settingsv1.PeerAuthenticationSettings{
			Enabled: true,
		}
		expectedPeerAuth.Spec.Mtls.Mode = v1beta1.PeerAuthentication_MutualTLS_UNSET
		translator := peerAuth.NewTranslator(ctx, settings, nil)

		mockIstioBuilder.EXPECT().AddPeerAuthentications(expectedPeerAuth)
		translator.Translate(istioMesh, mockIstioBuilder, mockReporter)
	})

	It("will generate a peerAuth for an istio-mesh when settings allow it", func() {
		translator := peerAuth.NewTranslator(ctx, settings, nil)

		mockIstioBuilder.EXPECT().AddPeerAuthentications(expectedPeerAuth)
		translator.Translate(istioMesh, mockIstioBuilder, mockReporter)
	})
})
