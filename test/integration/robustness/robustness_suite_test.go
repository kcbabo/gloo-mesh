package robustness_test

import (
	"context"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/controllerutils"
	"k8s.io/apimachinery/pkg/api/errors"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func TestRobustness(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Robustness Suite")
}

var (
	// how long the test will wait for state to become consistent with expectation
	testCaseTimeout = time.Second * 5

	rootCtx = context.TODO()

	// paths to the directories containing networking and agent crds
	crdDirs = []string{
		utils.GetModulePath("istio.io/istio") + "/manifests/charts/base/crds",
		util.MustGetThisDir() + "/../../../install/helm/agent-crds/crds",
		util.MustGetThisDir() + "/../../../install/helm/gloo-mesh-crds/crds",
	}

	mgmtMgr       manager.Manager
	remoteEastMgr manager.Manager
	remoteWestMgr manager.Manager

	params bootstrap.StartParameters
)

var _ = BeforeSuite(func() {
	envtestAssets := os.Getenv("KUBEBUILDER_ASSETS")
	if envtestAssets == "" {
		Fail("KUBEBUILDER_ASSETS not set. Run `make install-test-tools` to install integration test assets")
	}

	mgmtMgr, remoteEastMgr, remoteWestMgr = runManager(), runManager(), runManager()

	remoteManagers := map[string]manager.Manager{
		"remote-east": remoteEastMgr,
		"remote-west": remoteWestMgr,
	}
	mcWatcher := utils.FakeClusterWatcher{
		RootCtx:            rootCtx,
		PerClusterManagers: remoteManagers,
	}
	mcClient := utils.FakeMulticlusterClient{
		PerClusterManagers: remoteManagers,
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

	startGlooMeshComponents(rootCtx)
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
		err = mgr.Start(rootCtx)
		Expect(err).NotTo(HaveOccurred())
	}()

	return mgr
}

func startGlooMeshComponents(ctx context.Context) {

	netOpts := &mesh_networking.NetworkingOpts{
		Options: &bootstrap.Options{},
	}
	netOpts.AddToFlags(&pflag.FlagSet{}) // just to set defaults

	err := mesh_networking.NetworkingStarter{
		NetworkingOpts: netOpts,
		MakeExtensions: func(ctx context.Context, parameters bootstrap.StartParameters) mesh_networking.ExtensionOpts {
			// no extensions
			return mesh_networking.ExtensionOpts{}
		},
	}.StartReconciler(ctx, params)
	Expect(err).NotTo(HaveOccurred())

	// start discovery
	discOpts := &mesh_discovery.DiscoveryOpts{
		Options: &bootstrap.Options{},
	}
	discOpts.AddToFlags(&pflag.FlagSet{}) // just to set defaults

	err = reconciliation.Start(
		ctx,
		"",
		params.MasterManager,
		params.Clusters,
		params.McClient,
		params.SnapshotHistory,
		params.VerboseMode,
		&params.SettingsRef,
	)
	Expect(err).NotTo(HaveOccurred())

	// start cert agent on remote-east cluster
	err = agent.StartWithManager(ctx, remoteEastMgr, agent.CertAgentReconcilerExtensionOpts{})
	Expect(err).NotTo(HaveOccurred())

	// start cert agent on remote-west cluster
	err = agent.StartWithManager(ctx, remoteWestMgr, agent.CertAgentReconcilerExtensionOpts{})
	Expect(err).NotTo(HaveOccurred())

	// give components 3 sec to start
	time.Sleep(time.Second * 3)
}

// Test case definition
type testCase struct {
	// the sequence of expected states to test. each state will define a set of inputs on each cluster
	// and the expected outputs to see on that cluster after Gloo Mesh components eventually reconcile.
	// We may also want to verify that some things *consistently* are true, for example, never incorrectly garbage collecting (even temporarily) a resource.
	states []testState
}

// runs through each of the test states, applies them to the cluster
// should be run inside an It() block.
func (c testCase) execute(ctx context.Context) {
	for _, state := range c.states {
		state.execute(ctx)
	}
}

// state of the test at a given point in time
type testState struct {
	description string
	// state expected in each cluster
	clusterStates map[manager.Manager]configState
}

// applies the input resources of the state to each cluster and verifies eventually they are consistent
func (s testState) execute(ctx context.Context) {
	eg, ctx := errgroup.WithContext(ctx)
	for mgr, state := range s.clusterStates {
		for _, obj := range state.clusterInputs {
			// upsert all objects
			upsert(ctx, mgr, obj)
		}

		// start watching for expected outputs
		for _, expectedObj := range state.clusterExpectedOutputs {
			mgr := mgr // pike
			expectedObj := expectedObj // pike
			// begin checking the object eventually exists and matches the expected state
			eg.Go(func() error {
				defer GinkgoRecover()
				Eventually(func() (client.Object, error) {
					return getLatest(ctx, mgr, expectedObj)
				}, testCaseTimeout).Should(matchKubeObject{objToMatch: expectedObj}, s.description)
				contextutils.LoggerFrom(ctx).Infof("expected object %v matched", sets.TypedKey(expectedObj))
				return nil
			})
		}
	}
	err := eg.Wait()
	Expect(err).NotTo(HaveOccurred())
}

// the expected state of the cluster given the provided inputs.
// it is expected that when the inputs are removed the outputs will also be removed.
type configState struct {
	clusterInputs          []client.Object
	clusterExpectedOutputs []client.Object
}

// shared test funcs

// upsert an obj
func upsert(ctx context.Context, mgr manager.Manager, obj client.Object) {
	_, err := controllerutils.Upsert(ctx, mgr.GetClient(), obj.DeepCopyObject().(client.Object))
	ExpectWithOffset(3, err).NotTo(HaveOccurred())
}

// fetch latest version of an obj from the manager
func getLatest(ctx context.Context, mgr manager.Manager, obj client.Object) (client.Object, error) {

	key := client.ObjectKeyFromObject(obj)

	// Always valid because obj is client.Object
	existing := obj.DeepCopyObject().(client.Object)

	if err := mgr.GetClient().Get(ctx, key, existing); err != nil {
		if !errors.IsNotFound(err) {
			ExpectWithOffset(3, err).NotTo(HaveOccurred())
		}
		return nil, err
	}
	return existing, nil
}

type matchKubeObject struct {
	objToMatch client.Object
}

func (m matchKubeObject) Match(actual interface{}) (success bool, err error) {
	obj, ok := actual.(client.Object)
	if !ok {
		return false, nil
	}
	return controllerutils.ObjectsEqual(m.objToMatch, obj), nil
}

func (m matchKubeObject) FailureMessage(actual interface{}) (message string) {
	obj, ok := actual.(client.Object)
	if !ok {
		return fmt.Sprintf("expected object %T, got %T", m.objToMatch, actual)
	}
	return fmt.Sprintf("expected obj %v, got %v", sprintObj(m.objToMatch), sprintObj(obj))
}

func (m matchKubeObject) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("expected obj not to match %T/%v", m.objToMatch, m.objToMatch)
}

// print an object as a human readable string
func sprintObj(obj client.Object) string {
	y, err := yaml.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%v<yaml:\n%s\n>", sets.TypedKey(obj), y)
}