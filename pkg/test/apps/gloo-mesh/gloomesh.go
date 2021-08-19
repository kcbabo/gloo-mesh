package gloo_mesh

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"istio.io/istio/pkg/test/framework/resource"
	"istio.io/istio/pkg/test/util/retry"
)

type Config struct {
	ClusterKubeConfigs                  map[string]string
	DeployControlPlaneToManagementPlane bool
}

const glooMeshVersion = "1.1.0-rc1"

func Deploy(deploymentCtx *context.DeploymentContext, cfg *Config, licenseKey string) resource.SetupFn {
	return func(ctx resource.Context) error {
		if deploymentCtx == nil {
			*deploymentCtx = context.DeploymentContext{}
		}
		deploymentCtx.Meshes = []context.GlooMeshInstance{}
		var i context.GlooMeshInstance
		var err error

		version := os.Getenv("GLOO_MESH_ENTERPRISE_VERSION")
		if version == "" {
			version = glooMeshVersion
		}

		// sort the keys so its always cluster-0 for mp
		keys := make([]string, 0)
		for k, _ := range cfg.ClusterKubeConfigs {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// install management plane
		// we need the MP to always be installed and added to the instance list first because
		// istio uninstalls in reverse order meaning control planes unregister first before uninstalling MP
		mpcfg := InstanceConfig{
			managementPlane:               true,
			managementPlaneKubeConfigPath: cfg.ClusterKubeConfigs[keys[0]],
			version:                       version,
			clusterDomain:                 "",
			cluster:                       ctx.Clusters().GetByName(keys[0]),
		}
		mpInstance, err := newInstance(ctx, mpcfg, licenseKey)
		if err != nil {
			return err
		}
		deploymentCtx.Meshes = append(deploymentCtx.Meshes, mpInstance)

		var relayAddress string
		if licenseKey != "" {
			if err := retry.UntilSuccess(func() error {
				relayAddress, err = mpInstance.GetRelayServerAddress()
				if err != nil {
					return err
				}
				return nil
			}, retry.Timeout(3*time.Minute), retry.Delay(15*time.Second)); err != nil {
				return fmt.Errorf("failed to find relay server address %s", err.Error())
			}
		}

		// install control planes
		for index, key := range keys {
			value := cfg.ClusterKubeConfigs[key]
			if index == 0 && !cfg.DeployControlPlaneToManagementPlane {
				// skip deploying CP to MP
				index++
				continue
			}

			cpcfg := InstanceConfig{
				managementPlane:                   false,
				controlPlaneKubeConfigPath:        value,
				managementPlaneKubeConfigPath:     cfg.ClusterKubeConfigs[keys[0]],
				version:                           version,
				cluster:                           ctx.Clusters().GetByName(key),
				managementPlaneRelayServerAddress: relayAddress,
				clusterDomain:                     "",
			}
			i, err = newInstance(ctx, cpcfg, licenseKey)
			if err != nil {
				return err
			}
			deploymentCtx.Meshes = append(deploymentCtx.Meshes, i)
			index++
		}
		return nil
	}
}
