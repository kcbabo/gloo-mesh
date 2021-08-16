package main

import (
	"github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"istio.io/istio/pkg/test/framework/resource"
	"testing"

	"github.com/solo-io/gloo-mesh/pkg/test/common"
	"istio.io/istio/pkg/test/framework"
)

func TestVirtualGateways(t *testing.T) {
	framework.
		NewTest(t).
		Run(func(ctx framework.TestContext) {

			tgs := []common.TestGroup{
				{
					Name: "traffic-policies",
					Cases: []common.TestCase{
						{
							Name:        "happy-path",
							Description: "",
							Test:        testHappyPath,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "gateway.yaml",
							Folder:      "gloo-mesh/virtual-gateway",
						},
					},
				},
			}
			for _, tg := range tgs {
				tg.Run(ctx, t, &deploymentCtx)
			}
		})

}

func testHappyPath(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {

}