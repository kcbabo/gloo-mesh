package main

import (
	"github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"istio.io/istio/pkg/test/echo/common/scheme"
	"istio.io/istio/pkg/test/framework/components/echo"
	"istio.io/istio/pkg/test/framework/resource"
	"net/http"
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
					Name: "virtual-gateway",
					Cases: []common.TestCase{
						{
							Name:        "prefix-using-virtualgateway",
							Description: "testing prefix based routing using virtualgateway only",
							Test:        prefixTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "prefix-virtualgateway.yaml",
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

// requests for /frontend apply to all ingressgateways and route to frontend in cluster0
// requests for /backend apply to all ingressgateways and route to backend in cluster0
func prefixTest(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {

	// should reach the frontend
	cluster := ctx.Clusters()[0]
	// frontend calling backend in mesh using virtual destination in same cluster
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("no-mesh").And(echo.InCluster(cluster)))
	apiHost := "api.solo.io"
	cluster0GatewayAddress, err := deploymentCtx.Meshes[0].GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
	if err != nil {
		t.Fatal(err)
		return
	}

	// happy requests from cluster-0
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 80,
		},
		Scheme:  scheme.HTTP,
		Address: cluster0GatewayAddress,
		Headers: map[string][]string{
			"Host": {apiHost},
		},
		Method:    http.MethodGet,
		Path:      "/frontend",
		Count:     5,
		Validator: echo.ExpectOK(),
	})
	// should reach the backend
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 80,
		},
		Scheme:  scheme.HTTP,
		Address: cluster0GatewayAddress,
		Headers: map[string][]string{
			"Host": {apiHost},
		},
		Method:    http.MethodGet,
		Path:      "/backend",
		Count:     5,
		Validator: echo.ExpectOK(),
	})

	cluster1GatewayAddress, err := deploymentCtx.Meshes[1].GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
	if err != nil {
		t.Fatal(err)
		return
	}

	// happy requests from cluster-0
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 80,
		},
		Scheme:  scheme.HTTP,
		Address: cluster1GatewayAddress,
		Headers: map[string][]string{
			"Host": {apiHost},
		},
		Method:    http.MethodGet,
		Path:      "/frontend",
		Count:     5,
		Validator: echo.ExpectOK(),
	})

	// should reach the backend
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 80,
		},
		Scheme:  scheme.HTTP,
		Address: cluster1GatewayAddress,
		Headers: map[string][]string{
			"Host": {apiHost},
		},
		Method:    http.MethodGet,
		Path:      "/backend",
		Count:     5,
		Validator: echo.ExpectOK(),
	})

}
