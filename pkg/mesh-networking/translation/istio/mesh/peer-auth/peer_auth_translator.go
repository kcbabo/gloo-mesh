package peerAuth

import (
	"context"

	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators/tls"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/api/security/v1beta1"

	security_istio_io_v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

//go:generate mockgen -source ./peer_auth_translator.go -destination mocks/peer_auth_translator.go

const (
	defaultPeerName = "default"
)

// TODO - or would it be more apt to say that it produces a peerAUth for each mesh?
// the PeerAuthentication translator translates a Mesh into a PeerAuthentication.
type Translator interface {
	// Translate translates the appropriate PeerAuthentication for the given Mesh.
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
	settings  *settingsv1.Settings
	workloads discoveryv1sets.WorkloadSet
}

func NewTranslator(
	ctx context.Context,
	settings *settingsv1.Settings,
	workloads discoveryv1sets.WorkloadSet,
) Translator {
	return &translator{
		ctx:       ctx,
		settings:  settings,
		workloads: workloads,
	}
}

// translate the appropriate resources for the given Mesh.
func (t *translator) Translate(
	mesh *discoveryv1.Mesh,
	istioOutputs istio.Builder,
	reporter reporting.Reporter,
) {
	if !t.settings.Spec.PeerAuth.GetEnabled() {
		return
	}

	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.Type)
		return
	}
	if auth, err := t.updatePeerAuthValues(mesh, istioOutputs); err != nil {
		reporter.ReportPeerAuthenticationToMesh(mesh, auth, err)
	}
}

func (t *translator) updatePeerAuthValues(
	mesh *discoveryv1.Mesh,
	istioOutputs istio.Builder,
) (*security_istio_io_v1beta1.PeerAuthentication, error) {
	istioTlsMode, err := tls.MapIstioTlsModeToPeerAuth(t.settings.Spec.GetPeerAuth().GetPeerAuthTlsMode())

	if err != nil {
		return nil, err
	}
	// note: if 'mode' is set to the UNSET enum, then the resulting yaml will simply not set this value.
	// Which I found unintuitive, but it does make literal sense.
	newPeerAuth := &security_istio_io_v1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Name:        defaultPeerName,
			Namespace:   mesh.Spec.AgentInfo.AgentNamespace,
			ClusterName: mesh.Spec.GetIstio().GetInstallation().GetCluster(),
		},
		Spec: v1beta1.PeerAuthentication{
			Mtls: &v1beta1.PeerAuthentication_MutualTLS{
				Mode: istioTlsMode,
			},
		},
	}

	istioOutputs.AddPeerAuthentications(newPeerAuth)
	return newPeerAuth, nil
}
