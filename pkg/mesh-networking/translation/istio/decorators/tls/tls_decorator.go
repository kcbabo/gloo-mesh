package tls

import (
	"github.com/rotisserie/eris"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	istiov1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/decorators"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	"istio.io/api/security/v1beta1"
)

const (
	decoratorName = "tls"
)

func init() {
	decorators.Register(decoratorConstructor)
}

func decoratorConstructor(_ decorators.Parameters) decorators.Decorator {
	return NewTlsDecorator()
}

// Handles setting TLS on a DestinationRule.
type tlsDecorator struct{}

var _ decorators.TrafficPolicyDestinationRuleDecorator = &tlsDecorator{}

func NewTlsDecorator() *tlsDecorator {
	return &tlsDecorator{}
}

func (d *tlsDecorator) DecoratorName() string {
	return decoratorName
}

func (d *tlsDecorator) ApplyTrafficPolicyToDestinationRule(
	appliedPolicy *istiov1.AppliedTrafficPolicy,
	_ *discoveryv1.Destination,
	output *networkingv1alpha3spec.DestinationRule,
	registerField decorators.RegisterField,
	peerAuthPolicy *settingsv1.PeerAuthenticationSettings,
) error {
	tlsSettings, err := d.translateTlsSettings(appliedPolicy.Spec, peerAuthPolicy)
	if err != nil {
		return err
	}
	if tlsSettings != nil {
		if err := registerField(&output.TrafficPolicy.Tls, tlsSettings); err != nil {
			return err
		}
		output.TrafficPolicy.Tls = tlsSettings
	}

	return nil
}

func (d *tlsDecorator) translateTlsSettings(
	trafficPolicy *istiov1.TrafficPolicySpec,
	peerAuthPolicy *settingsv1.PeerAuthenticationSettings,
) (*networkingv1alpha3spec.ClientTLSSettings, error) {
	// If TrafficPolicy doesn't specify mTLS configuration, use global default populated upstream during initialization.
	istioMtls := trafficPolicy.GetPolicy().GetMtls().GetIstio()
	if istioMtls == nil {
		return nil, nil
	}
	var istioTlsMode networkingv1alpha3spec.ClientTLSSettings_TLSmode
	var err error
	if peerAuthPolicy.GetEnabled() {
		istioTlsMode, err = MapIstioTlsModeCheckPeerAuth(istioMtls.GetTlsMode(), peerAuthPolicy.GetPeerAuthTlsMode())
	} else {
		istioTlsMode, err = MapIstioTlsMode(istioMtls.GetTlsMode())
	}
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &networkingv1alpha3spec.ClientTLSSettings{
		Mode: istioTlsMode,
	}, nil
}

// exported for use by destination rule translator
func MapIstioTlsMode(tlsMode istiov1.TrafficPolicySpec_Policy_MTLS_Istio_TLSmode) (networkingv1alpha3spec.ClientTLSSettings_TLSmode, error) {
	switch tlsMode {
	case istiov1.TrafficPolicySpec_Policy_MTLS_Istio_DISABLE:
		return networkingv1alpha3spec.ClientTLSSettings_DISABLE, nil
	case istiov1.TrafficPolicySpec_Policy_MTLS_Istio_SIMPLE:
		return networkingv1alpha3spec.ClientTLSSettings_SIMPLE, nil
	case istiov1.TrafficPolicySpec_Policy_MTLS_Istio_ISTIO_MUTUAL:
		return networkingv1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL, nil
	default:
		return 0, eris.Errorf("unrecognized Istio TLS mode %s", tlsMode)
	}
}

// exported for use by destination rule translator
// Requires the peerAuth TLS setting in to check that the two are not incompatible.
func MapIstioTlsModeCheckPeerAuth(tlsMode istiov1.TrafficPolicySpec_Policy_MTLS_Istio_TLSmode, peerAuthMode settingsv1.PeerAuthenticationSettings_MutualTLS) (networkingv1alpha3spec.ClientTLSSettings_TLSmode, error) {
	switch tlsMode {
	case istiov1.TrafficPolicySpec_Policy_MTLS_Istio_DISABLE:
		if peerAuthMode == settingsv1.PeerAuthenticationSettings_STRICT {
			return 0, eris.Errorf("Destination Rule TLS policy set to %s, but peerAuthentication policy set to incompatible value %s", tlsMode, peerAuthMode)
		}
	case istiov1.TrafficPolicySpec_Policy_MTLS_Istio_SIMPLE:
		if peerAuthMode == settingsv1.PeerAuthenticationSettings_DISABLE {
			return 0, eris.Errorf("Destination Rule TLS policy set to %s, but peerAuthentication policy set to incompatible value %s", tlsMode, peerAuthMode)
		}
	case istiov1.TrafficPolicySpec_Policy_MTLS_Istio_ISTIO_MUTUAL:
		if peerAuthMode == settingsv1.PeerAuthenticationSettings_DISABLE {
			return 0, eris.Errorf("Destination Rule TLS policy set to %s, but peerAuthentication policy set to incompatible value %s", tlsMode, peerAuthMode)
		}
	default:
		return 0, eris.Errorf("unrecognized Istio TLS mode %s", tlsMode)
	}
	return MapIstioTlsMode(tlsMode)
}

func MapIstioTlsModeToPeerAuth(tlsMode settingsv1.PeerAuthenticationSettings_MutualTLS) (v1beta1.PeerAuthentication_MutualTLS_Mode, error) {
	switch tlsMode {
	case settingsv1.PeerAuthenticationSettings_UNSET:
		return v1beta1.PeerAuthentication_MutualTLS_UNSET, nil
	case settingsv1.PeerAuthenticationSettings_DISABLE:
		return v1beta1.PeerAuthentication_MutualTLS_DISABLE, nil
	case settingsv1.PeerAuthenticationSettings_PERMISSIVE:
		return v1beta1.PeerAuthentication_MutualTLS_PERMISSIVE, nil
	case settingsv1.PeerAuthenticationSettings_STRICT:
		return v1beta1.PeerAuthentication_MutualTLS_STRICT, nil
	default:
		return 0, eris.Errorf("unrecognized Istio MutualTLS mode for peerAuth: %s", tlsMode)
	}
}
