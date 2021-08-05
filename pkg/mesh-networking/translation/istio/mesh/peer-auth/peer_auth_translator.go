package peerAuth

import (
	"context"
	"fmt"

	"istio.io/api/meta/v1alpha1"
	"istio.io/api/security/v1beta1"

	security_istio_io_v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	corev1 "k8s.io/api/core/v1"
)

//go:generate mockgen -source ./peer_auth_translator.go -destination mocks/peer_auth_translator.go

const (
	defaultIstioOrg              = "Istio"
	defaultCitadelServiceAccount = "istio-citadel"
	defaultTrustDomain           = "cluster.local" // The default SPIFFE URL value for trust domain
	defaultIstioNamespace        = "istio-system"
	// name of the istio root CA secret
	// https://istio.io/latest/docs/tasks/security/cert-management/plugin-ca-cert/
	istioCaSecretName = "cacerts"
	// name of the istio root CA configmap distributed to all namespaces
	// copied from https://github.com/istio/istio/blob/88a2bfb/pilot/pkg/serviceregistry/kube/controller/namespacecontroller.go#L39
	// not imported due to issues with dependeny imports
	istioCaConfigMapName = "istio-ca-root-cert"

	defaultRootCertTTLDays                = 365
	defaultRootCertRsaKeySize             = 4096
	defaultOrgName                        = "gloo-mesh"
	defaultSecretRotationGracePeriodRatio = 0.10

	defaultPeerName = "default"
)

var (
	signingCertSecretType = corev1.SecretType(
		fmt.Sprintf("%s/generated_signing_cert", certificatesv1.SchemeGroupVersion.Group),
	)

	// used when the user provides a nil root cert
	defaultSelfSignedRootCa = &networkingv1.RootCertificateAuthority{
		CaSource: &networkingv1.RootCertificateAuthority_Generated{
			Generated: &certificatesv1.CommonCertOptions{
				TtlDays:         defaultRootCertTTLDays,
				RsaKeySizeBytes: defaultRootCertRsaKeySize,
				OrgName:         defaultOrgName,
			},
		},
	}
)

// used by networking reconciler to filter ignored secrets
func IsSigningCert(secret *corev1.Secret) bool {
	return secret.Type == signingCertSecretType
}

// the VirtualService translator translates a Mesh into a VirtualService.
type Translator interface {
	// Translate translates the appropriate VirtualService and DestinationRule for the given Mesh.
	// returns nil if no VirtualService or DestinationRule is required for the Mesh (i.e. if no VirtualService/DestinationRule features are required, such as subsets).
	// Output resources will be added to the istio.Builder
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		mesh *discoveryv1.Mesh,
		istioOutputs istio.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	ctx       context.Context
	workloads discoveryv1sets.WorkloadSet
}

func NewTranslator(
	ctx context.Context,
	workloads discoveryv1sets.WorkloadSet,
) Translator {
	return &translator{
		ctx:       ctx,
		workloads: workloads,
	}
}

// translate the appropriate resources for the given Mesh.
func (t *translator) Translate(
	mesh *discoveryv1.Mesh,
	istioOutputs istio.Builder,
	reporter reporting.Reporter,
) {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.Type)
		return
	}
	// Todo Double check that there aren't any other skip conditions

	if err := t.updatePeerAuthValues(mesh, istioOutputs); err != nil {
		reporter.ReportPeerAuthenticationToMesh(mesh, err)
	}
}

func (t *translator) updatePeerAuthValues(
	mesh *discoveryv1.Mesh,
	istioOutputs istio.Builder,
) error {

	//peerAuths []*security_istio_io_v1beta1.PeerAuthentication
	contextutils.LoggerFrom(t.ctx).Infof("temp testing: added peer auth")
	// TODO Figure out exactly how to allow user configs for this.
	// Idea 1: Everything is tied to a config on the mesh - mesh.Spec.PeerAuth
	// Idea 2: Selector (and possibly PortLevelMtls) is tied to a mesh config, Mtls is from existing config in v-mesh
	// Neither of these quite feel right, check with someone with a better understanding of config placement.
	istioOutputs.AddPeerAuthentications(&security_istio_io_v1beta1.PeerAuthentication{
		Spec: v1beta1.PeerAuthentication{
			Selector:      nil,
			Mtls:          nil,
			PortLevelMtls: nil,
		},
		Status: v1alpha1.IstioStatus{
			Conditions:         nil,
			ValidationMessages: nil,
			ObservedGeneration: 0,
		}})

	return nil
}

/*func (t *translator) constructPeerAuthentication(
	mesh *discoveryv1.Mesh,
) (*security_istio_io_v1beta1.PeerAuthentication, error) {

	istioTlsMode, err := tls.MapIstioTlsModeToPeerAuth(t.settings.Spec.Mtls.GetIstio().GetTlsMode())
	if err != nil {
		return nil, err
	}

	return &security_istio_io_v1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Name:        defaultPeerName,
			Namespace:   mesh.Spec.AgentInfo.AgentNamespace,
			ClusterName: mesh.Spec.GetIstio().GetInstallation().GetCluster(),
		},
		Spec: v1beta1.PeerAuthentication{
			Mtls: &v1beta1.PeerAuthentication_MutualTLS{
				Mode: istioTlsMode,
			},
		}}, nil
}*/
