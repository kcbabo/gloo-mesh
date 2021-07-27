package peerAuth

import (
	"context"
	"fmt"
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/mtls"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	"istio.io/istio/pkg/spiffe"
	"istio.io/istio/security/pkg/pki/util"

	settingsv1 "github.com/solo-io/gloo-mesh/pkg/api/settings.mesh.gloo.solo.io/v1"

	"istio.io/api/security/v1beta1"

	security_istio_io_v1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"

	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/version"
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

	defaultAuthName = "default"
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
		virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
		istioOutputs istio.Builder,
		localOutputs local.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct {
	settings  *settingsv1.Settings
	ctx       context.Context
	secrets   corev1sets.SecretSet
	workloads discoveryv1sets.WorkloadSet
}

// General TODO: NewTranslator bunch of this is copied from mtls_translator - would it be worth abstracting out
// struct references and making those original copied functions static, then referencing them here?
func NewTranslator(
	settings *settingsv1.Settings,
	ctx context.Context,
	secrets corev1sets.SecretSet,
	workloads discoveryv1sets.WorkloadSet,
) Translator {
	return &translator{
		settings:  settings,
		ctx:       ctx,
		secrets:   secrets,
		workloads: workloads,
	}
}

// translate the appropriate resources for the given Mesh.
func (t *translator) Translate(
	mesh *discoveryv1.Mesh,
	virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
	istioOutputs istio.Builder,
	localOutputs local.Builder,
	reporter reporting.Reporter,
) {
	istioMesh := mesh.Spec.GetIstio()
	if istioMesh == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("ignoring non istio mesh %v %T", sets.Key(mesh), mesh.Spec.Type)
		return
	}

	// Todo Double check that there aren't any other skip conditions
	if err := t.updatePeerAuthValues(mesh, virtualMesh, istioOutputs, localOutputs); err != nil {
		reporter.ReportPeerAuthenticationToMesh(mesh, virtualMesh.Ref, err)
	}
}

func (t *translator) updatePeerAuthValues(
	mesh *discoveryv1.Mesh,
	virtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
	istioOutputs istio.Builder,
	localOutputs local.Builder,
) error {

	//peerAuths []*security_istio_io_v1beta1.PeerAuthentication
	// TODO Figure out exactly how to allow user configs for this.
	// Idea 1: Everything is tied to a config on the mesh - mesh.Spec.PeerAuth
	// Idea 2: Selector (and possibly PortLevelMtls) is tied to a mesh config, Mtls is from existing config in v-mesh
	// Neither of these quite feel right, check with someone with a better understanding of config placement.

	mtlsConfig := virtualMesh.Spec.MtlsConfig
	if mtlsConfig == nil {
		// nothing to do
		contextutils.LoggerFrom(t.ctx).Debugf("no translation for VirtualMesh %v which has no mTLS configuration", sets.Key(mesh))
		return nil
	}
	if mtlsConfig.TrustModel == nil {
		return eris.Errorf("must specify trust model to use for issuing certificates")
	}

	switch trustModel := mtlsConfig.TrustModel.(type) {
	case *networkingv1.VirtualMeshSpec_MTLSConfig_Shared:
		return t.configureSharedTrust(
			mesh,
			trustModel.Shared,
			virtualMesh.Ref,
			istioOutputs,
			localOutputs,
			mtlsConfig.AutoRestartPods,
		)
	case *networkingv1.VirtualMeshSpec_MTLSConfig_Limited:
		return eris.Errorf("limited trust not supported in version %v of Gloo Mesh", version.Version)
	}

	istioOutputs.AddPeerAuthentications(&security_istio_io_v1beta1.PeerAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Name:        defaultAuthName,
			Namespace:   mesh.Spec.AgentInfo.AgentNamespace,
			ClusterName: mesh.Spec.GetIstio().GetInstallation().GetCluster(),
		},
		Spec: v1beta1.PeerAuthentication{
			Mtls: &v1beta1.PeerAuthentication_MutualTLS{
				Mode: v1beta1.PeerAuthentication_MutualTLS_PERMISSIVE,
			},
		}})

	return nil
}

// will create the secret if it is self-signed,
// otherwise will return the user-provided secret ref in the mtls config
func (t *translator) configureSharedTrust(
	mesh *discoveryv1.Mesh,
	sharedTrust *networkingv1.SharedTrust,
	virtualMeshRef *skv2corev1.ObjectRef,
	istioOutputs istio.Builder,
	localOutputs local.Builder,
	autoRestartPods bool,
) error {
	agentInfo := mesh.Spec.AgentInfo
	if agentInfo == nil {
		contextutils.LoggerFrom(t.ctx).Debugf("cannot configure root certificates for mesh %v which has no cert-agent", sets.Key(mesh))
		return nil
	}
	// Construct the skeleton of the issuedCertificate
	issuedCertificate, podBounceDirective := t.constructIssuedCertificate(
		mesh,
		sharedTrust,
		agentInfo.AgentNamespace,
		autoRestartPods,
	)

	switch typedCa := sharedTrust.GetCertificateAuthority().(type) {
	case *networkingv1.SharedTrust_IntermediateCertificateAuthority:
		// Copy intermediate CA data to IssuedCertificate
		issuedCertificate.Spec.CertificateAuthority = &certificatesv1.IssuedCertificateSpec_AgentCa{
			AgentCa: typedCa.IntermediateCertificateAuthority,
		}
	case *networkingv1.SharedTrust_RootCertificateAuthority:
		switch typedCaSource := typedCa.RootCertificateAuthority.GetCaSource().(type) {
		case *networkingv1.RootCertificateAuthority_Generated:
			// Generated CA cert secret.
			// Check if it exists
			rootCaSecret, err := t.getOrCreateGeneratedCaSecret(
				typedCaSource.Generated,
				virtualMeshRef,
				localOutputs,
			)
			if err != nil {
				return err
			}
			issuedCertificate.Spec.CertificateAuthority = &certificatesv1.IssuedCertificateSpec_GlooMeshCa{
				GlooMeshCa: &certificatesv1.RootCertificateAuthority{
					CertificateAuthority: &certificatesv1.RootCertificateAuthority_SigningCertificateSecret{
						SigningCertificateSecret: rootCaSecret,
					},
				},
			}
			// Set deprecated field for backwards compatibility
			issuedCertificate.Spec.SigningCertificateSecret = rootCaSecret
		case *networkingv1.RootCertificateAuthority_Secret:
			issuedCertificate.Spec.CertificateAuthority = &certificatesv1.IssuedCertificateSpec_GlooMeshCa{
				GlooMeshCa: &certificatesv1.RootCertificateAuthority{
					CertificateAuthority: &certificatesv1.RootCertificateAuthority_SigningCertificateSecret{
						SigningCertificateSecret: typedCaSource.Secret,
					},
				},
			}
			issuedCertificate.Spec.SigningCertificateSecret = typedCaSource.Secret
			// Set deprecated field for backwards compatibility
		default:
			return eris.Errorf("No root ca source specified for Virtual Mesh (%s)", sets.Key(virtualMeshRef))
		}
	default:
		return eris.Errorf("No ca source specified for Virtual Mesh (%s)", sets.Key(virtualMeshRef))
	}

}



// will create the secret if it is self-signed,
// otherwise will return the user-provided secret ref in the mtls config
func (t *translator) getOrCreateGeneratedCaSecret(
	generatedRootCa *certificatesv1.CommonCertOptions,
	virtualMeshRef *skv2corev1.ObjectRef,
	localOutputs local.Builder,
) (*skv2corev1.ObjectRef, error) {

	if generatedRootCa == nil {
		generatedRootCa = defaultSelfSignedRootCa.GetGenerated()
	}

	generatedSecretName := virtualMeshRef.Name + "." + virtualMeshRef.Namespace
	// write the signing secret to the gloomesh namespace
	generatedSecretNamespace := defaults.GetPodNamespace()
	// use the existing secret if it exists
	rootCaSecret := &skv2corev1.ObjectRef{
		Name:      generatedSecretName,
		Namespace: generatedSecretNamespace,
	}
	selfSignedCertSecret, err := t.secrets.Find(rootCaSecret)
	if err != nil {
		selfSignedCert, err := mtls.GenerateSelfSignedCert(generatedRootCa)
		if err != nil {
			// should never happen
			return nil, err
		}
		// the self signed cert goes to the master/local cluster
		selfSignedCertSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: generatedSecretName,
				// write to the agent namespace
				Namespace: generatedSecretNamespace,
				// ensure the secret is written to the maser/local cluster
				ClusterName: "",
				Labels:      metautils.TranslatedObjectLabels(),
			},
			Data: selfSignedCert.ToSecretData(),
			Type: signingCertSecretType,
		}
	}

	// Append the VirtualMesh as a parent to the output secret
	metautils.AppendParent(t.ctx, selfSignedCertSecret, virtualMeshRef, networkingv1.VirtualMesh{}.GVK())

	localOutputs.AddSecrets(selfSignedCertSecret)

	return rootCaSecret, nil
}

func (t *translator) constructIssuedCertificate(
	mesh *discoveryv1.Mesh,
	sharedTrust *networkingv1.SharedTrust,
	agentNamespace string,
	autoRestartPods bool,
) (*certificatesv1.IssuedCertificate, *certificatesv1.PodBounceDirective) {
	istioMesh := mesh.Spec.GetIstio()

	trustDomain := istioMesh.GetTrustDomain()
	if trustDomain == "" {
		trustDomain = defaultTrustDomain
	}
	istiodServiceAccount := istioMesh.GetIstiodServiceAccount()
	if istiodServiceAccount == "" {
		istiodServiceAccount = defaultCitadelServiceAccount
	}
	istioNamespace := istioMesh.GetInstallation().GetNamespace()
	if istioNamespace == "" {
		istioNamespace = defaultIstioNamespace
	}

	clusterName := istioMesh.GetInstallation().GetCluster()
	issuedCertificateMeta := metav1.ObjectMeta{
		Name: mesh.Name,
		// write to the agent namespace
		Namespace: agentNamespace,
		// write to the mesh cluster
		ClusterName: clusterName,
		Labels:      metautils.TranslatedObjectLabels(),
	}

	// get the pods that need to be bounced for this mesh
	podsToBounce := mtls.GetPodsToBounce(mesh, sharedTrust, t.workloads, autoRestartPods)
	var (
		podBounceDirective *certificatesv1.PodBounceDirective
		podBounceRef       *skv2corev1.ObjectRef
	)
	if len(podsToBounce) > 0 {
		podBounceDirective = &certificatesv1.PodBounceDirective{
			ObjectMeta: issuedCertificateMeta,
			Spec: certificatesv1.PodBounceDirectiveSpec{
				PodsToBounce: podsToBounce,
			},
		}
		podBounceRef = ezkube.MakeObjectRef(podBounceDirective)
	}

	issuedCert := &certificatesv1.IssuedCertificate{
		ObjectMeta: issuedCertificateMeta,
		Spec: certificatesv1.IssuedCertificateSpec{
			Hosts: []string{mtls.BuildSpiffeURI(trustDomain, istioNamespace, istiodServiceAccount)},
			CertOptions: mtls.BuildDefaultCertOptions(
				sharedTrust.GetIntermediateCertOptions(),
				defaultIstioOrg,
			),
			// Set deprecated field for backwards compatibility
			Org:                defaultIstioOrg,
			PodBounceDirective: podBounceRef,
		},
	}

	// Only set issuedCert when not using vault CA
	if sharedTrust.GetIntermediateCertificateAuthority().GetVault() == nil {
		// the default location of the istio CA Certs secret
		// the certificate workflow will produce a cert with this ref
		issuedCert.Spec.IssuedCertificateSecret = &skv2corev1.ObjectRef{
			Name:      istioCaSecretName,
			Namespace: istioNamespace,
		}
	}

	// issue a certificate to the mesh agent
	return issuedCert, podBounceDirective
}
