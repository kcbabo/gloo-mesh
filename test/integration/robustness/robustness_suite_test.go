package robustness_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/output/discovery"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/local"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/resource"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/ghodss/yaml"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/controllerutils"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

const (
	yamlSeparator = "\n---\n"
	yamlDir       = "test_yamls"
)

func TestRobustness(t *testing.T) {
	RegisterFailHandler(func(message string, callerSkip ...int) {
		failurePrintMessage()
		Fail(message, callerSkip...)
	})
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

	// cluster names
	mgmt       = "mgmt"
	remoteEast = "remote-east"
	remoteWest = "remote-west"

	mgmtMgr       manager.Manager
	remoteEastMgr manager.Manager
	remoteWestMgr manager.Manager

	mgrNames map[manager.Manager]string

	// parameters for networking + discovery
	params bootstrap.StartParameters

	// used to print debug info on failure
	currentTest     testState
	currentTestLock *sync.RWMutex
)

var _ = BeforeSuite(func() {
	envtestAssets := os.Getenv("KUBEBUILDER_ASSETS")
	if envtestAssets == "" {
		Fail("KUBEBUILDER_ASSETS not set. Run `make install-test-tools` to install integration test assets")
	}

	currentTest = testState{}
	currentTestLock = &sync.RWMutex{}

	mgmtMgr, remoteEastMgr, remoteWestMgr = runManager(), runManager(), runManager()
	mgrNames = map[manager.Manager]string{
		mgmtMgr:       mgmt,
		remoteEastMgr: remoteEast,
		remoteWestMgr: remoteWest,
	}

	remoteManagers := map[string]manager.Manager{
		remoteEast: remoteEastMgr,
		remoteWest: remoteWestMgr,
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
	netOpts.VerboseMode = true

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
	discOpts.VerboseMode = true

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
	name        string
	description string
}

// applies the input resources of the state to each cluster and verifies eventually they are consistent
func (s testState) execute(ctx context.Context) {
	currentTestLock.Lock()
	currentTest = s
	currentTestLock.Unlock()
	By(s.description)
	eg, ctx := errgroup.WithContext(ctx)
	for mgr, cluster := range mgrNames {
		var inputYamls []string

		clusterInputs, err := readTestManifest(
			s.name,
			cluster,
			manifestType_inputs,
			mgr.GetScheme(),
		)
		Expect(err).NotTo(HaveOccurred())
		for _, obj := range clusterInputs {
			gvk, err := apiutil.GVKForObject(obj, mgr.GetScheme())
			Expect(err).NotTo(HaveOccurred())
			obj.(resource.TypedObject).SetGroupVersionKind(gvk)

			// upsert all objects
			upsert(ctx, mgr, obj)

			inputYamls = append(inputYamls, sprintObj(obj))
		}

		clusterExpectedOutputs, err := readTestManifest(
			s.name,
			cluster,
			manifestType_outputs,
			mgr.GetScheme(),
		)
		Expect(err).NotTo(HaveOccurred())
		// start watching for expected outputs
		var outputYamls []string
		for _, expectedObj := range clusterExpectedOutputs {

			gvk, err := apiutil.GVKForObject(expectedObj, mgr.GetScheme())
			Expect(err).NotTo(HaveOccurred())
			expectedObj.(resource.TypedObject).SetGroupVersionKind(gvk)

			mgr := mgr                 // pike
			expectedObj := expectedObj // pike
			// begin checking the object eventually exists and matches the expected state
			eg.Go(func() error {
				defer GinkgoRecover()
				Eventually(func() (client.Object, error) {
					return getLatest(ctx, mgr, expectedObj)
				}, testCaseTimeout).Should(matchKubeObject{objToMatch: expectedObj}, s.description+" on test cluster "+cluster)
				fmt.Fprintf(GinkgoWriter, "expected object %v matched", sets.TypedKey(expectedObj))
				return nil
			})
			outputYamls = append(outputYamls, sprintObj(expectedObj))
		}

		err = writeTestYamls(s.name, mgrNames[mgr], inputYamls, outputYamls)
		Expect(err).NotTo(HaveOccurred())
	}
	err := eg.Wait()
	Expect(err).NotTo(HaveOccurred())
}

type manifestType string

const (
	manifestType_inputs  manifestType = "inputs"
	manifestType_outputs manifestType = "outputs"
	manifestType_actual  manifestType = "actual"
)

func testStateDir(testState string) string {
	return filepath.Join(util.MustGetThisDir(), yamlDir, testState)
}

func getTestManifestPath(testCase, cluster string, t manifestType) string {
	return filepath.Join(testStateDir(testCase), cluster+"_"+string(t)+".yaml")
}

func readTestManifest(testCase, cluster string, t manifestType, s *runtime.Scheme) ([]client.Object, error) {
	raw, err := ioutil.ReadFile(getTestManifestPath(testCase, cluster, t))
	if err != nil {
		return nil, err
	}
	objYamls := bytes.Split(raw, []byte(yamlSeparator))
	var objs []client.Object
	for _, objYam := range objYamls {
		typeOnly := &metav1.TypeMeta{}
		err := yaml.Unmarshal(objYam, typeOnly)
		if err != nil {
			return nil, err
		}
		if typeOnly.Kind == "" {
			// empty yaml
			continue
		}
		gvk := typeOnly.GetObjectKind().GroupVersionKind()
		obj, err := s.New(gvk)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(objYam, obj)
		if err != nil {
			return nil, err
		}
		objs = append(objs, obj.(client.Object))
	}
	return objs, nil
}

func writeTestYamls(testCase, cluster string, inputs, outputs []string) error {
	if err := os.MkdirAll(testStateDir(testCase), 0777); err != nil {
		return err
	}
	if err := ioutil.WriteFile(
		getTestManifestPath(testCase, cluster, manifestType_inputs),
		[]byte(strings.Join(inputs, yamlSeparator)),
		0644,
	); err != nil {
		return err
	}
	if err := ioutil.WriteFile(
		getTestManifestPath(testCase, cluster, manifestType_outputs),
		[]byte(strings.Join(outputs, yamlSeparator)),
		0644,
	); err != nil {
		return err
	}
	return nil
}

func writeActualYamls(testCase, cluster string, actualObjs map[schema.GroupVersionKind][]client.Object) error {
	if err := ioutil.WriteFile(
		getTestManifestPath(testCase, cluster, manifestType_actual), []byte(sprintAllObjs(actualObjs)), 0644); err != nil {
		return err
	}
	return nil
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
	ExpectWithOffset(2, err).NotTo(HaveOccurred())
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
	return fmt.Sprintf("# %v :\n%s\n", sets.TypedKey(obj), sanitizeObjString(y))
}

// removes server-set metadata fields
func sanitizeObjString(yamObj []byte) []byte {
	m := map[string]interface{}{}
	err := yaml.Unmarshal(yamObj, &m)
	if err != nil {
		panic(err)
	}
	meta, ok := m["metadata"]
	if !ok {
		panic("metadata does not exist in obj")
	}

	metaMap, ok := meta.(map[string]interface{})
	if !ok {
		panic("metadata not map[string]interface{}")
	}

	sanitizedMetaMap := map[string]interface{}{
		"name": metaMap["name"],
	}

	if namespace, ok := metaMap["namespace"]; ok {
		sanitizedMetaMap["namespace"] = namespace
	}
	if labels, ok := metaMap["labels"]; ok {
		sanitizedMetaMap["labels"] = labels
	}
	if annotations, ok := metaMap["annotations"]; ok {
		sanitizedMetaMap["annotations"] = annotations
	}

	if status, ok := m["status"].(map[string]interface{}); ok && len(status) == 0 {
		delete(m, "status")
	}

	m["metadata"] = sanitizedMetaMap

	y, err := yaml.Marshal(m)
	if err != nil {
		panic(err)
	}

	return y
}

func failurePrintMessage() {
	if mgrNames != nil {
		for _, mgr := range []manager.Manager{
			mgmtMgr,
			remoteEastMgr,
			remoteWestMgr,
		} {
			name := mgrNames[mgr]
			objsForMgr, err := getAllObjs(rootCtx, mgr)
			if err != nil {
				fmt.Fprintf(GinkgoWriter, "failed getting snapshot in failure state for %v", name)
			} else {
				currentTestLock.RLock()
				t := currentTest
				currentTestLock.RUnlock()
				err := writeActualYamls(t.name, name, objsForMgr)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

var gvksToPrint = func() []schema.GroupVersionKind {
	var gvks []schema.GroupVersionKind
	gvks = append(gvks, istio.SnapshotGVKs...)
	gvks = append(gvks, local.SnapshotGVKs...)
	gvks = append(gvks, discovery.SnapshotGVKs...)
	return gvks
}()

func sprintAllObjs(allObjs map[schema.GroupVersionKind][]client.Object) string {
	var allObsStr []string
	for _, gvk := range gvksToPrint {
		var gvkStr []string
		for _, obj := range allObjs[gvk] {
			gvkStr = append(gvkStr, sprintObj(obj))
		}
		if len(gvkStr) == 0 {
			continue
		}
		allObsStr = append(allObsStr, fmt.Sprintf("### %v:\n\n%v\n\n", gvk.String(), strings.Join(gvkStr, yamlSeparator)))
	}
	return strings.Join(allObsStr, yamlSeparator)
}

func getAllObjs(ctx context.Context, mgr manager.Manager) (map[schema.GroupVersionKind][]client.Object, error) {
	objsByGvk := map[schema.GroupVersionKind][]client.Object{}
	for _, gvk := range gvksToPrint {
		metaList := &metav1.PartialObjectMetadataList{}
		metaList.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   gvk.Group,
			Version: gvk.Version,
			Kind:    gvk.Kind + "List",
		})
		err := mgr.GetClient().List(ctx, metaList)
		if err != nil && !meta.IsNoMatchError(err) {
			return nil, err
		}
		var items []client.Object
		for _, item := range metaList.Items {
			item := item // pike
			obj, err := mgr.GetScheme().New(gvk)
			if err != nil {
				return nil, err
			}
			typedObj, ok := obj.(resource.TypedObject)
			if !ok {
				return nil, eris.Errorf("not a typed obj %T", obj)
			}
			err = mgr.GetClient().Get(ctx, ezkube.MakeClientObjectKey(&item), typedObj)
			if err != nil {
				return nil, err
			}
			items = append(items, typedObj)
		}
		if len(items) == 0 {
			continue
		}
		objsByGvk[gvk] = items
	}
	return objsByGvk, nil
}
