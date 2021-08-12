package agent

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	corev1clients "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/agent/input"
	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/reconciliation"
	podbouncer "github.com/solo-io/gloo-mesh/pkg/certificates/agent/reconciliation/pod-bouncer"
	"github.com/solo-io/gloo-mesh/pkg/certificates/agent/translation"
	"github.com/solo-io/gloo-mesh/pkg/common/schemes"
	"github.com/solo-io/skv2/pkg/bootstrap"
)

// the mesh-networking controller is the Kubernetes Controller/Operator
// which processes k8s storage events to produce
// discovered resources.
func Start(ctx context.Context, opts bootstrap.Options) error {
	return bootstrap.Start(
		ctx,
		StartFuncExt(func(ctx context.Context, parameters bootstrap.StartParameters) CertAgentReconcilerExtensionOpts {
			return CertAgentReconcilerExtensionOpts{}
		}),
		opts,
		schemes.SchemeBuilder,
		true,
	)
}

// Extended start function
func StartFuncExt(makeExtensionOpts MakeExtensionOpts) bootstrap.StartFunc {
	// start the main reconcile loop
	return func(ctx context.Context, parameters bootstrap.StartParameters) error {

		extOpts := makeExtensionOpts(ctx, parameters)

		return StartWithManager(ctx, parameters.MasterManager, extOpts)
	}
}

// exported for test
func StartWithManager(ctx context.Context, mgr manager.Manager, extOpts CertAgentReconcilerExtensionOpts) error {
	extOpts.initDefaults()

	snapshotBuilder := input.NewSingleClusterBuilder(mgr)

	translator := translation.NewCertAgentTranslator()

	podBounder := podbouncer.NewPodBouncer(
		corev1clients.NewPodClient(mgr.GetClient()),
		extOpts.RootCertMatcher,
	)

	return reconciliation.Start(
		ctx,
		snapshotBuilder,
		mgr,
		podBounder,
		extOpts.MakeTranslator(translator),
	)
}

// Options for extending the functionality of the Networking controller

type MakeExtensionOpts func(ctx context.Context, parameters bootstrap.StartParameters) CertAgentReconcilerExtensionOpts

// Options for overriding functionality of the Networking Reconciler
type CertAgentReconcilerExtensionOpts struct {

	// Hook to override Translator used by Networking Reconciler
	MakeTranslator func(translator translation.Translator) translation.Translator
	// Pod Bouncer to be used by translator, allows overriding the dependency
	RootCertMatcher podbouncer.RootCertMatcher
}

func (opts *CertAgentReconcilerExtensionOpts) initDefaults() {

	if opts.MakeTranslator == nil {
		// use default translator
		opts.MakeTranslator = func(translator translation.Translator) translation.Translator {
			return translator
		}
	}

	if opts.RootCertMatcher == nil {
		opts.RootCertMatcher = podbouncer.NewSecretRootCertMatcher()
	}
}
