package utils

import (
	"context"
	"sort"

	"github.com/rotisserie/eris"
	"github.com/solo-io/skv2/pkg/multicluster"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// FakeMulticlusterClient wraps a set of rest configs for multiple clusters
type FakeMulticlusterClient struct {
	// pre load your own managers here
	PerClusterManagers map[string]manager.Manager
}

var _ multicluster.Client = FakeMulticlusterClient{}

func (f FakeMulticlusterClient) ListClusters() []string {
	var clusters []string
	for cluster := range f.PerClusterManagers {
		clusters = append(clusters, cluster)
	}
	sort.Strings(clusters)
	return clusters
}

func (f FakeMulticlusterClient) Cluster(name string) (client.Client, error) {
	m, ok := f.PerClusterManagers[name]
	if !ok {
		return nil, eris.Errorf("no manager found for cluster %v", name)
	}
	return m.GetClient(), nil
}

// FakeMulticlusterClient wraps a set of rest configs for multiple clusters
type FakeClusterWatcher struct {
	// shared root ctx
	RootCtx context.Context
	// pre load your own managers here
	PerClusterManagers map[string]manager.Manager
}

var _ multicluster.Interface = FakeClusterWatcher{}

func (f FakeClusterWatcher) Run(_ manager.Manager) error {
	return nil
}

func (f FakeClusterWatcher) RegisterClusterHandler(handler multicluster.ClusterHandler) {
	// only works if the clusters have been added first
	for cluster, mgr := range f.PerClusterManagers {
		cluster := cluster // pike
		mgr := mgr         // pike
		go func() {
			// run in a goroutine
			// as to not block
			handler.AddCluster(
				f.RootCtx,
				cluster,
				mgr,
			)
		}()
	}
}

func (f FakeClusterWatcher) Cluster(cluster string) (manager.Manager, error) {
	m, ok := f.PerClusterManagers[cluster]
	if !ok {
		return nil, eris.Errorf("no manager found for cluster %v", cluster)
	}
	return m, nil
}

func (f FakeClusterWatcher) ListClusters() []string {
	var clusters []string
	for cluster := range f.PerClusterManagers {
		clusters = append(clusters, cluster)
	}
	sort.Strings(clusters)
	return clusters
}
