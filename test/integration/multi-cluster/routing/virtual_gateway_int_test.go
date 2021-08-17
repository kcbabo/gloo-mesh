package main

import (
	"fmt"
	"github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"istio.io/istio/pkg/test/echo/client"
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
							Folder:      "gloo-mesh/virtual-gateway/prefix",
						},
						{
							Name:        "prefix-using-virtualhost",
							Description: "testing prefix based routing using virtualhost",
							Test:        prefixTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "prefix-virtualhost.yaml",
							Folder:      "gloo-mesh/virtual-gateway/prefix",
						},
						{
							Name:        "prefix-using-routetable",
							Description: "testing prefix based routing using routetable",
							Test:        prefixTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "prefix-routetable.yaml",
							Folder:      "gloo-mesh/virtual-gateway/prefix",
						},
						{
							Name:        "prefix-using-multi-virtualhost",
							Description: "testing prefix based routing using multiple virtual hosts",
							Test:        prefixTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "prefix-multi-virtualhost.yaml",
							Folder:      "gloo-mesh/virtual-gateway/prefix",
						},
						{
							Name:        "prefix-using-multi-routetable",
							Description: "testing prefix based routing using multiple route tables",
							Test:        prefixTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "prefix-multi-routetable.yaml",
							Folder:      "gloo-mesh/virtual-gateway/prefix",
						},
						{
							Name:        "prefix-using-multi-virtualhost-multi-routetable",
							Description: "testing prefix based routing using route tables per virtualhost",
							Test:        prefixTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "prefix-multi-virtualhost-multi-routetable.yaml",
							Folder:      "gloo-mesh/virtual-gateway/prefix",
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

	frontend := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	backend := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("backend").And(echo.InCluster(cluster)))

	apiHost := "api.solo.io"
	cluster0GatewayAddress, err := deploymentCtx.Meshes[0].GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
	if err != nil {
		t.Fatal(err)
		return
	}
	frontendIDValidator := echo.ValidatorFunc(func(responses client.ParsedResponses, _ error) error {
		for _, r := range responses {
			if r.Hostname != frontend.WorkloadsOrFail(t)[0].PodName() {
				return fmt.Errorf("response from pod %s does not match expected %s", r.Hostname, frontend.WorkloadsOrFail(t)[0].PodName())
			}
		}
		return nil
	})
	backendIDValidator := echo.ValidatorFunc(func(responses client.ParsedResponses, _ error) error {
		for _, r := range responses {
			if r.Hostname != backend.WorkloadsOrFail(t)[0].PodName() {
				return fmt.Errorf("response from pod %s does not match expected %s", r.Hostname, backend.WorkloadsOrFail(t)[0].PodName())
			}
		}
		return nil
	})

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
		Validator: echo.And(echo.ExpectOK(), frontendIDValidator),
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
		Validator: echo.And(echo.ExpectOK(), backendIDValidator),
	})
	// invalid host
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 80,
		},
		Scheme:  scheme.HTTP,
		Address: cluster0GatewayAddress,
		Headers: map[string][]string{
			"Host": {"wrong.solo.io"},
		},
		Method:    http.MethodGet,
		Path:      "/frontend",
		Count:     5,
		Validator: echo.ExpectCode("404"),
	})
	// invalid path
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
		Path:      "/wrong",
		Count:     5,
		Validator: echo.ExpectCode("404"),
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
		Validator: echo.And(echo.ExpectOK(), frontendIDValidator),
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
		Validator: echo.And(echo.ExpectOK(), backendIDValidator),
	})
	// invalid host
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 80,
		},
		Scheme:  scheme.HTTP,
		Address: cluster1GatewayAddress,
		Headers: map[string][]string{
			"Host": {"wrong.solo.io"},
		},
		Method:    http.MethodGet,
		Path:      "/frontend",
		Count:     5,
		Validator: echo.ExpectCode("404"),
	})
	// invalid path
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
		Path:      "/wrong",
		Count:     5,
		Validator: echo.ExpectCode("404"),
	})
}
