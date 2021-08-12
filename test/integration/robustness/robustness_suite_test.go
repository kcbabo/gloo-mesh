package robustness_test

import (
	"context"

	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo-mesh/pkg/certificates/agent"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/common/schemes"
	mesh_discovery "github.com/solo-io/gloo-mesh/pkg/mesh-discovery"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/reconciliation"
	mesh_networking "github.com/solo-io/gloo-mesh/pkg/mesh-networking"
	"github.com/solo-io/gloo-mesh/test/utils"
	"github.com/solo-io/go-utils/errgroup"
	"github.com/solo-io/skv2/codegen/util"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/bootstrap"
	"github.com/solo-io/skv2/pkg/stats"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

/*
todo:
- reproduce the issue by deleting workloads in the cluster and watching the issuedcert get recreated. can also watch the vmesh status.
- write integration test (maybe e2e?) to repro it. run both discovery and networking components in the test in memory with fake clients
- solutions to implement:
   - workload discovery decouple from pods
   - destionation discovery decouple from workload
   - persist last known good mesh config

*/

//NOTE: set USE_EXISTING_CLUSTER=1 to use a real k8s cluster. Otherwise envtest will use controller-runtime's envtest to spin up a local apiserver.

func TestRobustness(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Robustness Suite")
}

// paths to the directories containing networking and agent crds
var (
	ctx     = context.TODO()
	crdDirs = []string{
		util.MustGetThisDir() + "/../../../install/helm/agent-crds/crds",
		util.MustGetThisDir() + "/../../../install/helm/gloo-mesh-crds/crds",
	}

	mgmtMgr   manager.Manager
	remoteMgr manager.Manager

	params bootstrap.StartParameters
)

var _ = BeforeSuite(func() {
	envtestAssets := os.Getenv("KUBEBUILDER_ASSETS")
	if envtestAssets == "" {
		Fail("KUBEBUILDER_ASSETS not set. Run `make install-test-tools` to install integration test assets")
	}

	mgmtMgr, remoteMgr = runManager(), runManager()

	managers := map[string]manager.Manager{
		"mgmt-cluster":   mgmtMgr,
		"remote-cluster": remoteMgr,
	}
	mcWatcher := utils.FakeClusterWatcher{
		RootCtx:            ctx,
		PerClusterManagers: managers,
	}
	mcClient := utils.FakeMulticlusterClient{
		PerClusterManagers: managers,
	}

	params = bootstrap.StartParameters{
		MasterManager:   mgmtMgr,
		McClient:        mcClient,
		Clusters:        mcWatcher,
		SnapshotHistory: stats.NewSnapshotHistory(), // just to prevent panic
		SettingsRef: v1.ObjectRef{
			Name:      defaults.DefaultSettingsName,
			Namespace: defaults.GetPodNamespace(),
		},
		VerboseMode: true,
	}

	startGlooMeshComponents()
})

func runManager() manager.Manager {
	env := envtest.Environment{
		CRDDirectoryPaths: crdDirs,
	}
	cfg, err := env.Start()
	Expect(err).NotTo(HaveOccurred())
	s := scheme.Scheme
	err = schemes.AddToScheme(s)
	Expect(err).NotTo(HaveOccurred())
	mgr, err := manager.New(cfg, manager.Options{
		Scheme:             s,
		MetricsBindAddress: "0",
	})
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
	}()

	return mgr
}

func startGlooMeshComponents() {

	ctx := context.TODO()

	eg, ctx := errgroup.WithContext(ctx)

	// start networking
	eg.Go(func() error {
		netOpts := &mesh_networking.NetworkingOpts{
			Options: &bootstrap.Options{},
		}
		netOpts.AddToFlags(&pflag.FlagSet{}) // just to set defaults

		return mesh_networking.NetworkingStarter{
			NetworkingOpts: netOpts,
			MakeExtensions: func(ctx context.Context, parameters bootstrap.StartParameters) mesh_networking.ExtensionOpts {
				// no extensions
				return mesh_networking.ExtensionOpts{}
			},
		}.StartReconciler(ctx, params)
	})

	// start discovery
	eg.Go(func() error {
		discOpts := &mesh_discovery.DiscoveryOpts{
			Options: &bootstrap.Options{},
		}
		discOpts.AddToFlags(&pflag.FlagSet{}) // just to set defaults

		return reconciliation.Start(
			ctx,
			"",
			params.MasterManager,
			params.Clusters,
			params.McClient,
			params.SnapshotHistory,
			params.VerboseMode,
			&params.SettingsRef,
		)
	})

	// start cert agent on mgmt cluster
	eg.Go(func() error {
		return agent.StartWithManager(ctx, mgmtMgr, agent.CertAgentReconcilerExtensionOpts{})
	})

	// start cert agent on remote cluster
	eg.Go(func() error {
		return agent.StartWithManager(ctx, remoteMgr, agent.CertAgentReconcilerExtensionOpts{})
	})

	go func() {
		defer GinkgoRecover()
		err := eg.Wait()
		Expect(err).NotTo(HaveOccurred())
	}()

	// give components 3 sec to start
	time.Sleep(time.Second * 3)
}
