// Code generated by skv2. DO NOT EDIT.

//go:generate mockgen -source ./snapshot.go -destination mocks/snapshot.go

// Definitions for Output Snapshots
package smi

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/rotisserie/eris"
	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/pkg/multicluster"
	"sigs.k8s.io/controller-runtime/pkg/client"

	split_smi_spec_io_v1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
	split_smi_spec_io_v1alpha2_sets "github.com/solo-io/external-apis/pkg/api/smi/split.smi-spec.io/v1alpha2/sets"

	access_smi_spec_io_v1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	access_smi_spec_io_v1alpha2_sets "github.com/solo-io/external-apis/pkg/api/smi/access.smi-spec.io/v1alpha2/sets"

	specs_smi_spec_io_v1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	specs_smi_spec_io_v1alpha3_sets "github.com/solo-io/external-apis/pkg/api/smi/specs.smi-spec.io/v1alpha3/sets"
)

// this error can occur if constructing a Partitioned Snapshot from a resource
// that is missing the partition label
var MissingRequiredLabelError = func(labelKey, resourceKind string, obj ezkube.ResourceId) error {
	return eris.Errorf("expected label %v not on labels of %v %v", labelKey, resourceKind, sets.Key(obj))
}

// the snapshot of output resources produced by a translation
type Snapshot interface {

	// return the set of TrafficSplits with a given set of labels
	TrafficSplits() []LabeledTrafficSplitSet
	// return the set of TrafficTargets with a given set of labels
	TrafficTargets() []LabeledTrafficTargetSet
	// return the set of HTTPRouteGroups with a given set of labels
	HTTPRouteGroups() []LabeledHTTPRouteGroupSet

	// apply the snapshot to the local cluster, garbage collecting stale resources
	ApplyLocalCluster(ctx context.Context, clusterClient client.Client, errHandler output.ErrorHandler)

	// apply resources from the snapshot across multiple clusters, garbage collecting stale resources
	ApplyMultiCluster(ctx context.Context, multiClusterClient multicluster.Client, errHandler output.ErrorHandler)

	// serialize the entire snapshot as JSON
	MarshalJSON() ([]byte, error)
}

type snapshot struct {
	name string

	trafficSplits   []LabeledTrafficSplitSet
	trafficTargets  []LabeledTrafficTargetSet
	hTTPRouteGroups []LabeledHTTPRouteGroupSet
	clusters        []string
}

func NewSnapshot(
	name string,

	trafficSplits []LabeledTrafficSplitSet,
	trafficTargets []LabeledTrafficTargetSet,
	hTTPRouteGroups []LabeledHTTPRouteGroupSet,
	clusters ...string, // the set of clusters to apply the snapshot to. only required for multicluster snapshots.
) Snapshot {
	return &snapshot{
		name: name,

		trafficSplits:   trafficSplits,
		trafficTargets:  trafficTargets,
		hTTPRouteGroups: hTTPRouteGroups,
		clusters:        clusters,
	}
}

// automatically partitions the input resources
// by the presence of the provided label.
func NewLabelPartitionedSnapshot(
	name,
	labelKey string, // the key by which to partition the resources

	trafficSplits split_smi_spec_io_v1alpha2_sets.TrafficSplitSet,

	trafficTargets access_smi_spec_io_v1alpha2_sets.TrafficTargetSet,

	hTTPRouteGroups specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet,
	clusters ...string, // the set of clusters to apply the snapshot to. only required for multicluster snapshots.
) (Snapshot, error) {

	partitionedTrafficSplits, err := partitionTrafficSplitsByLabel(labelKey, trafficSplits)
	if err != nil {
		return nil, err
	}
	partitionedTrafficTargets, err := partitionTrafficTargetsByLabel(labelKey, trafficTargets)
	if err != nil {
		return nil, err
	}
	partitionedHTTPRouteGroups, err := partitionHTTPRouteGroupsByLabel(labelKey, hTTPRouteGroups)
	if err != nil {
		return nil, err
	}

	return NewSnapshot(
		name,

		partitionedTrafficSplits,
		partitionedTrafficTargets,
		partitionedHTTPRouteGroups,
		clusters...,
	), nil
}

// simplified constructor for a snapshot
// with a single label partition (i.e. all resources share a single set of labels).
func NewSinglePartitionedSnapshot(
	name string,
	snapshotLabels map[string]string, // a single set of labels shared by all resources

	trafficSplits split_smi_spec_io_v1alpha2_sets.TrafficSplitSet,

	trafficTargets access_smi_spec_io_v1alpha2_sets.TrafficTargetSet,

	hTTPRouteGroups specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet,
	clusters ...string, // the set of clusters to apply the snapshot to. only required for multicluster snapshots.
) (Snapshot, error) {

	labeledTrafficSplits, err := NewLabeledTrafficSplitSet(trafficSplits, snapshotLabels)
	if err != nil {
		return nil, err
	}
	labeledTrafficTargets, err := NewLabeledTrafficTargetSet(trafficTargets, snapshotLabels)
	if err != nil {
		return nil, err
	}
	labeledHTTPRouteGroups, err := NewLabeledHTTPRouteGroupSet(hTTPRouteGroups, snapshotLabels)
	if err != nil {
		return nil, err
	}

	return NewSnapshot(
		name,

		[]LabeledTrafficSplitSet{labeledTrafficSplits},
		[]LabeledTrafficTargetSet{labeledTrafficTargets},
		[]LabeledHTTPRouteGroupSet{labeledHTTPRouteGroups},
		clusters...,
	), nil
}

// apply the desired resources to the cluster state; remove stale resources where necessary
func (s *snapshot) ApplyLocalCluster(ctx context.Context, cli client.Client, errHandler output.ErrorHandler) {
	var genericLists []output.ResourceList

	for _, outputSet := range s.trafficSplits {
		genericLists = append(genericLists, outputSet.Generic())
	}
	for _, outputSet := range s.trafficTargets {
		genericLists = append(genericLists, outputSet.Generic())
	}
	for _, outputSet := range s.hTTPRouteGroups {
		genericLists = append(genericLists, outputSet.Generic())
	}

	output.Snapshot{
		Name:        s.name,
		ListsToSync: genericLists,
	}.SyncLocalCluster(ctx, cli, errHandler)
}

// apply the desired resources to multiple cluster states; remove stale resources where necessary
func (s *snapshot) ApplyMultiCluster(ctx context.Context, multiClusterClient multicluster.Client, errHandler output.ErrorHandler) {
	var genericLists []output.ResourceList

	for _, outputSet := range s.trafficSplits {
		genericLists = append(genericLists, outputSet.Generic())
	}
	for _, outputSet := range s.trafficTargets {
		genericLists = append(genericLists, outputSet.Generic())
	}
	for _, outputSet := range s.hTTPRouteGroups {
		genericLists = append(genericLists, outputSet.Generic())
	}

	output.Snapshot{
		Name:        s.name,
		Clusters:    s.clusters,
		ListsToSync: genericLists,
	}.SyncMultiCluster(ctx, multiClusterClient, errHandler)
}

func partitionTrafficSplitsByLabel(labelKey string, set split_smi_spec_io_v1alpha2_sets.TrafficSplitSet) ([]LabeledTrafficSplitSet, error) {
	setsByLabel := map[string]split_smi_spec_io_v1alpha2_sets.TrafficSplitSet{}

	for _, obj := range set.List() {
		if obj.Labels == nil {
			return nil, MissingRequiredLabelError(labelKey, "TrafficSplit", obj)
		}
		labelValue := obj.Labels[labelKey]
		if labelValue == "" {
			return nil, MissingRequiredLabelError(labelKey, "TrafficSplit", obj)
		}

		setForValue, ok := setsByLabel[labelValue]
		if !ok {
			setForValue = split_smi_spec_io_v1alpha2_sets.NewTrafficSplitSet()
			setsByLabel[labelValue] = setForValue
		}
		setForValue.Insert(obj)
	}

	// partition by label key
	var partitionedTrafficSplits []LabeledTrafficSplitSet

	for labelValue, setForValue := range setsByLabel {
		labels := map[string]string{labelKey: labelValue}

		partitionedSet, err := NewLabeledTrafficSplitSet(setForValue, labels)
		if err != nil {
			return nil, err
		}

		partitionedTrafficSplits = append(partitionedTrafficSplits, partitionedSet)
	}

	// sort for idempotency
	sort.SliceStable(partitionedTrafficSplits, func(i, j int) bool {
		leftLabelValue := partitionedTrafficSplits[i].Labels()[labelKey]
		rightLabelValue := partitionedTrafficSplits[j].Labels()[labelKey]
		return leftLabelValue < rightLabelValue
	})

	return partitionedTrafficSplits, nil
}

func partitionTrafficTargetsByLabel(labelKey string, set access_smi_spec_io_v1alpha2_sets.TrafficTargetSet) ([]LabeledTrafficTargetSet, error) {
	setsByLabel := map[string]access_smi_spec_io_v1alpha2_sets.TrafficTargetSet{}

	for _, obj := range set.List() {
		if obj.Labels == nil {
			return nil, MissingRequiredLabelError(labelKey, "TrafficTarget", obj)
		}
		labelValue := obj.Labels[labelKey]
		if labelValue == "" {
			return nil, MissingRequiredLabelError(labelKey, "TrafficTarget", obj)
		}

		setForValue, ok := setsByLabel[labelValue]
		if !ok {
			setForValue = access_smi_spec_io_v1alpha2_sets.NewTrafficTargetSet()
			setsByLabel[labelValue] = setForValue
		}
		setForValue.Insert(obj)
	}

	// partition by label key
	var partitionedTrafficTargets []LabeledTrafficTargetSet

	for labelValue, setForValue := range setsByLabel {
		labels := map[string]string{labelKey: labelValue}

		partitionedSet, err := NewLabeledTrafficTargetSet(setForValue, labels)
		if err != nil {
			return nil, err
		}

		partitionedTrafficTargets = append(partitionedTrafficTargets, partitionedSet)
	}

	// sort for idempotency
	sort.SliceStable(partitionedTrafficTargets, func(i, j int) bool {
		leftLabelValue := partitionedTrafficTargets[i].Labels()[labelKey]
		rightLabelValue := partitionedTrafficTargets[j].Labels()[labelKey]
		return leftLabelValue < rightLabelValue
	})

	return partitionedTrafficTargets, nil
}

func partitionHTTPRouteGroupsByLabel(labelKey string, set specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet) ([]LabeledHTTPRouteGroupSet, error) {
	setsByLabel := map[string]specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet{}

	for _, obj := range set.List() {
		if obj.Labels == nil {
			return nil, MissingRequiredLabelError(labelKey, "HTTPRouteGroup", obj)
		}
		labelValue := obj.Labels[labelKey]
		if labelValue == "" {
			return nil, MissingRequiredLabelError(labelKey, "HTTPRouteGroup", obj)
		}

		setForValue, ok := setsByLabel[labelValue]
		if !ok {
			setForValue = specs_smi_spec_io_v1alpha3_sets.NewHTTPRouteGroupSet()
			setsByLabel[labelValue] = setForValue
		}
		setForValue.Insert(obj)
	}

	// partition by label key
	var partitionedHTTPRouteGroups []LabeledHTTPRouteGroupSet

	for labelValue, setForValue := range setsByLabel {
		labels := map[string]string{labelKey: labelValue}

		partitionedSet, err := NewLabeledHTTPRouteGroupSet(setForValue, labels)
		if err != nil {
			return nil, err
		}

		partitionedHTTPRouteGroups = append(partitionedHTTPRouteGroups, partitionedSet)
	}

	// sort for idempotency
	sort.SliceStable(partitionedHTTPRouteGroups, func(i, j int) bool {
		leftLabelValue := partitionedHTTPRouteGroups[i].Labels()[labelKey]
		rightLabelValue := partitionedHTTPRouteGroups[j].Labels()[labelKey]
		return leftLabelValue < rightLabelValue
	})

	return partitionedHTTPRouteGroups, nil
}

func (s snapshot) TrafficSplits() []LabeledTrafficSplitSet {
	return s.trafficSplits
}

func (s snapshot) TrafficTargets() []LabeledTrafficTargetSet {
	return s.trafficTargets
}

func (s snapshot) HTTPRouteGroups() []LabeledHTTPRouteGroupSet {
	return s.hTTPRouteGroups
}

func (s snapshot) MarshalJSON() ([]byte, error) {
	snapshotMap := map[string]interface{}{"name": s.name}

	trafficSplitSet := split_smi_spec_io_v1alpha2_sets.NewTrafficSplitSet()
	for _, set := range s.trafficSplits {
		trafficSplitSet = trafficSplitSet.Union(set.Set())
	}
	snapshotMap["trafficSplits"] = trafficSplitSet.List()

	trafficTargetSet := access_smi_spec_io_v1alpha2_sets.NewTrafficTargetSet()
	for _, set := range s.trafficTargets {
		trafficTargetSet = trafficTargetSet.Union(set.Set())
	}
	snapshotMap["trafficTargets"] = trafficTargetSet.List()

	hTTPRouteGroupSet := specs_smi_spec_io_v1alpha3_sets.NewHTTPRouteGroupSet()
	for _, set := range s.hTTPRouteGroups {
		hTTPRouteGroupSet = hTTPRouteGroupSet.Union(set.Set())
	}
	snapshotMap["hTTPRouteGroups"] = hTTPRouteGroupSet.List()

	return json.Marshal(snapshotMap)
}

// LabeledTrafficSplitSet represents a set of trafficSplits
// which share a common set of labels.
// These labels are used to find diffs between TrafficSplitSets.
type LabeledTrafficSplitSet interface {
	// returns the set of Labels shared by this TrafficSplitSet
	Labels() map[string]string

	// returns the set of TrafficSplites with the given labels
	Set() split_smi_spec_io_v1alpha2_sets.TrafficSplitSet

	// converts the set to a generic format which can be applied by the Snapshot.Apply functions
	Generic() output.ResourceList
}

type labeledTrafficSplitSet struct {
	set    split_smi_spec_io_v1alpha2_sets.TrafficSplitSet
	labels map[string]string
}

func NewLabeledTrafficSplitSet(set split_smi_spec_io_v1alpha2_sets.TrafficSplitSet, labels map[string]string) (LabeledTrafficSplitSet, error) {
	// validate that each TrafficSplit contains the labels, else this is not a valid LabeledTrafficSplitSet
	for _, item := range set.List() {
		for k, v := range labels {
			// k=v must be present in the item
			if item.Labels[k] != v {
				return nil, eris.Errorf("internal error: %v=%v missing on TrafficSplit %v", k, v, item.Name)
			}
		}
	}

	return &labeledTrafficSplitSet{set: set, labels: labels}, nil
}

func (l *labeledTrafficSplitSet) Labels() map[string]string {
	return l.labels
}

func (l *labeledTrafficSplitSet) Set() split_smi_spec_io_v1alpha2_sets.TrafficSplitSet {
	return l.set
}

func (l labeledTrafficSplitSet) Generic() output.ResourceList {
	var desiredResources []ezkube.Object
	for _, desired := range l.set.List() {
		desiredResources = append(desiredResources, desired)
	}

	// enable list func for garbage collection
	listFunc := func(ctx context.Context, cli client.Client) ([]ezkube.Object, error) {
		var list split_smi_spec_io_v1alpha2.TrafficSplitList
		if err := cli.List(ctx, &list, client.MatchingLabels(l.labels)); err != nil {
			return nil, err
		}
		var items []ezkube.Object
		for _, item := range list.Items {
			item := item // pike
			items = append(items, &item)
		}
		return items, nil
	}

	return output.ResourceList{
		Resources:    desiredResources,
		ListFunc:     listFunc,
		ResourceKind: "TrafficSplit",
	}
}

// LabeledTrafficTargetSet represents a set of trafficTargets
// which share a common set of labels.
// These labels are used to find diffs between TrafficTargetSets.
type LabeledTrafficTargetSet interface {
	// returns the set of Labels shared by this TrafficTargetSet
	Labels() map[string]string

	// returns the set of TrafficTargetes with the given labels
	Set() access_smi_spec_io_v1alpha2_sets.TrafficTargetSet

	// converts the set to a generic format which can be applied by the Snapshot.Apply functions
	Generic() output.ResourceList
}

type labeledTrafficTargetSet struct {
	set    access_smi_spec_io_v1alpha2_sets.TrafficTargetSet
	labels map[string]string
}

func NewLabeledTrafficTargetSet(set access_smi_spec_io_v1alpha2_sets.TrafficTargetSet, labels map[string]string) (LabeledTrafficTargetSet, error) {
	// validate that each TrafficTarget contains the labels, else this is not a valid LabeledTrafficTargetSet
	for _, item := range set.List() {
		for k, v := range labels {
			// k=v must be present in the item
			if item.Labels[k] != v {
				return nil, eris.Errorf("internal error: %v=%v missing on TrafficTarget %v", k, v, item.Name)
			}
		}
	}

	return &labeledTrafficTargetSet{set: set, labels: labels}, nil
}

func (l *labeledTrafficTargetSet) Labels() map[string]string {
	return l.labels
}

func (l *labeledTrafficTargetSet) Set() access_smi_spec_io_v1alpha2_sets.TrafficTargetSet {
	return l.set
}

func (l labeledTrafficTargetSet) Generic() output.ResourceList {
	var desiredResources []ezkube.Object
	for _, desired := range l.set.List() {
		desiredResources = append(desiredResources, desired)
	}

	// enable list func for garbage collection
	listFunc := func(ctx context.Context, cli client.Client) ([]ezkube.Object, error) {
		var list access_smi_spec_io_v1alpha2.TrafficTargetList
		if err := cli.List(ctx, &list, client.MatchingLabels(l.labels)); err != nil {
			return nil, err
		}
		var items []ezkube.Object
		for _, item := range list.Items {
			item := item // pike
			items = append(items, &item)
		}
		return items, nil
	}

	return output.ResourceList{
		Resources:    desiredResources,
		ListFunc:     listFunc,
		ResourceKind: "TrafficTarget",
	}
}

// LabeledHTTPRouteGroupSet represents a set of hTTPRouteGroups
// which share a common set of labels.
// These labels are used to find diffs between HTTPRouteGroupSets.
type LabeledHTTPRouteGroupSet interface {
	// returns the set of Labels shared by this HTTPRouteGroupSet
	Labels() map[string]string

	// returns the set of HTTPRouteGroupes with the given labels
	Set() specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet

	// converts the set to a generic format which can be applied by the Snapshot.Apply functions
	Generic() output.ResourceList
}

type labeledHTTPRouteGroupSet struct {
	set    specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet
	labels map[string]string
}

func NewLabeledHTTPRouteGroupSet(set specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet, labels map[string]string) (LabeledHTTPRouteGroupSet, error) {
	// validate that each HTTPRouteGroup contains the labels, else this is not a valid LabeledHTTPRouteGroupSet
	for _, item := range set.List() {
		for k, v := range labels {
			// k=v must be present in the item
			if item.Labels[k] != v {
				return nil, eris.Errorf("internal error: %v=%v missing on HTTPRouteGroup %v", k, v, item.Name)
			}
		}
	}

	return &labeledHTTPRouteGroupSet{set: set, labels: labels}, nil
}

func (l *labeledHTTPRouteGroupSet) Labels() map[string]string {
	return l.labels
}

func (l *labeledHTTPRouteGroupSet) Set() specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet {
	return l.set
}

func (l labeledHTTPRouteGroupSet) Generic() output.ResourceList {
	var desiredResources []ezkube.Object
	for _, desired := range l.set.List() {
		desiredResources = append(desiredResources, desired)
	}

	// enable list func for garbage collection
	listFunc := func(ctx context.Context, cli client.Client) ([]ezkube.Object, error) {
		var list specs_smi_spec_io_v1alpha3.HTTPRouteGroupList
		if err := cli.List(ctx, &list, client.MatchingLabels(l.labels)); err != nil {
			return nil, err
		}
		var items []ezkube.Object
		for _, item := range list.Items {
			item := item // pike
			items = append(items, &item)
		}
		return items, nil
	}

	return output.ResourceList{
		Resources:    desiredResources,
		ListFunc:     listFunc,
		ResourceKind: "HTTPRouteGroup",
	}
}

type builder struct {
	ctx      context.Context
	name     string
	clusters []string

	trafficSplits split_smi_spec_io_v1alpha2_sets.TrafficSplitSet

	trafficTargets access_smi_spec_io_v1alpha2_sets.TrafficTargetSet

	hTTPRouteGroups specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet
}

func NewBuilder(ctx context.Context, name string) *builder {
	return &builder{
		ctx:  ctx,
		name: name,

		trafficSplits: split_smi_spec_io_v1alpha2_sets.NewTrafficSplitSet(),

		trafficTargets: access_smi_spec_io_v1alpha2_sets.NewTrafficTargetSet(),

		hTTPRouteGroups: specs_smi_spec_io_v1alpha3_sets.NewHTTPRouteGroupSet(),
	}
}

// the output Builder uses a builder pattern to allow
// iteratively collecting outputs before producing a final snapshot
type Builder interface {

	// add TrafficSplits to the collected outputs
	AddTrafficSplits(trafficSplits ...*split_smi_spec_io_v1alpha2.TrafficSplit)

	// get the collected TrafficSplits
	GetTrafficSplits() split_smi_spec_io_v1alpha2_sets.TrafficSplitSet

	// add TrafficTargets to the collected outputs
	AddTrafficTargets(trafficTargets ...*access_smi_spec_io_v1alpha2.TrafficTarget)

	// get the collected TrafficTargets
	GetTrafficTargets() access_smi_spec_io_v1alpha2_sets.TrafficTargetSet

	// add HTTPRouteGroups to the collected outputs
	AddHTTPRouteGroups(hTTPRouteGroups ...*specs_smi_spec_io_v1alpha3.HTTPRouteGroup)

	// get the collected HTTPRouteGroups
	GetHTTPRouteGroups() specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet

	// build the collected outputs into a label-partitioned snapshot
	BuildLabelPartitionedSnapshot(labelKey string) (Snapshot, error)

	// build the collected outputs into a snapshot with a single partition
	BuildSinglePartitionedSnapshot(snapshotLabels map[string]string) (Snapshot, error)

	// add a cluster to the collected clusters.
	// this can be used to collect clusters for use with MultiCluster snapshots.
	AddCluster(cluster string)
}

func (b *builder) AddTrafficSplits(trafficSplits ...*split_smi_spec_io_v1alpha2.TrafficSplit) {
	for _, obj := range trafficSplits {
		if obj == nil {
			continue
		}
		contextutils.LoggerFrom(b.ctx).Debugf("added output TrafficSplit %v", sets.Key(obj))
		b.trafficSplits.Insert(obj)
	}
}
func (b *builder) AddTrafficTargets(trafficTargets ...*access_smi_spec_io_v1alpha2.TrafficTarget) {
	for _, obj := range trafficTargets {
		if obj == nil {
			continue
		}
		contextutils.LoggerFrom(b.ctx).Debugf("added output TrafficTarget %v", sets.Key(obj))
		b.trafficTargets.Insert(obj)
	}
}
func (b *builder) AddHTTPRouteGroups(hTTPRouteGroups ...*specs_smi_spec_io_v1alpha3.HTTPRouteGroup) {
	for _, obj := range hTTPRouteGroups {
		if obj == nil {
			continue
		}
		contextutils.LoggerFrom(b.ctx).Debugf("added output HTTPRouteGroup %v", sets.Key(obj))
		b.hTTPRouteGroups.Insert(obj)
	}
}

func (b *builder) GetTrafficSplits() split_smi_spec_io_v1alpha2_sets.TrafficSplitSet {
	return b.trafficSplits
}

func (b *builder) GetTrafficTargets() access_smi_spec_io_v1alpha2_sets.TrafficTargetSet {
	return b.trafficTargets
}

func (b *builder) GetHTTPRouteGroups() specs_smi_spec_io_v1alpha3_sets.HTTPRouteGroupSet {
	return b.hTTPRouteGroups
}

func (b *builder) BuildLabelPartitionedSnapshot(labelKey string) (Snapshot, error) {
	return NewLabelPartitionedSnapshot(
		b.name,
		labelKey,

		b.trafficSplits,

		b.trafficTargets,

		b.hTTPRouteGroups,
		b.clusters...,
	)
}

func (b *builder) BuildSinglePartitionedSnapshot(snapshotLabels map[string]string) (Snapshot, error) {
	return NewSinglePartitionedSnapshot(
		b.name,
		snapshotLabels,

		b.trafficSplits,

		b.trafficTargets,

		b.hTTPRouteGroups,
		b.clusters...,
	)
}

func (b *builder) AddCluster(cluster string) {
	b.clusters = append(b.clusters, cluster)
}