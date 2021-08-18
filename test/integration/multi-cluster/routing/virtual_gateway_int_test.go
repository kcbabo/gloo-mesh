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

const (
	cluster0Index = 0
	cluster1Index = 1
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
							Name:        "single-cluster-gateway",
							Description: "testing prefix based routing using virtualgateway on a single cluster",
							Test:        singleClusterGatewayTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "single-cluster-gateway.yaml",
							Folder:      "gloo-mesh/virtual-gateway",
						},
						{
							Name:        "multi-cluster-application",
							Description: "testing routing when the application exists in multiple clusters",
							Test:        multiClusterApplicationTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "single-cluster-gateway.yaml",
							Folder:      "gloo-mesh/virtual-gateway",
						},
						{
							Name:        "global-service",
							Description: "routing internally and externally on the same dns name",
							Test:        globalServiceTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "multi-cluster-app.yaml",
							Folder:      "gloo-mesh/virtual-gateway",
						},
						{
							Name:        "prefix-using-virtualgateway",
							Description: "testing prefix based routing using virtualgateway only",
							Test:        prefixTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtualgateway.yaml",
							Folder:      "gloo-mesh/virtual-gateway/prefix",
						},
						{
							Name:        "prefix-using-virtualhost",
							Description: "testing prefix based routing using virtualhost",
							Test:        prefixTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtualhost.yaml",
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

// only the gateway in cluster-0 is configured to accept traffic
func singleClusterGatewayTest(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	cluster := ctx.Clusters()[cluster0Index]
	// calling the gateway from outside the mesh
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("no-mesh").And(echo.InCluster(cluster)))

	frontend := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))

	frontendIDValidator := echo.ValidatorFunc(func(responses client.ParsedResponses, _ error) error {
		for _, r := range responses {
			if r.Hostname != frontend.WorkloadsOrFail(t)[0].PodName() {
				return fmt.Errorf("response from pod %s does not match expected %s", r.Hostname, frontend.WorkloadsOrFail(t)[0].PodName())
			}
		}
		return nil
	})

	apiHost := "api.solo.io"
	cluster0GatewayAddress, err := deploymentCtx.Meshes[0].GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
	if err != nil {
		t.Fatal(err)
		return
	}
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
		Validator: echo.ExpectCode("503"),
	})
}

// the app exists in both clusters, so the gateway in each cluster should only route traffic to its local applications
func multiClusterApplicationTest(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {

	cluster := ctx.Clusters()[cluster0Index]
	// calling the gateway from outside the mesh
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("no-mesh").And(echo.InCluster(cluster)))

	frontend := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))

	frontend0IDValidator := echo.ValidatorFunc(func(responses client.ParsedResponses, _ error) error {
		for _, r := range responses {
			if r.Hostname != frontend.WorkloadsOrFail(t)[0].PodName() {
				return fmt.Errorf("response from pod %s does not match expected %s", r.Hostname, frontend.WorkloadsOrFail(t)[0].PodName())
			}
		}
		return nil
	})

	apiHost := "api.solo.io"
	cluster0GatewayAddress, err := deploymentCtx.Meshes[0].GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
	if err != nil {
		t.Fatal(err)
		return
	}

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
		Count:     15,
		Validator: echo.And(echo.ExpectOK(), frontend0IDValidator),
	})

	cluster = ctx.Clusters()[cluster1Index]

	frontend = deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))

	frontend1IDValidator := echo.ValidatorFunc(func(responses client.ParsedResponses, _ error) error {
		for _, r := range responses {
			if r.Hostname != frontend.WorkloadsOrFail(t)[0].PodName() {
				return fmt.Errorf("response from pod %s does not match expected %s", r.Hostname, frontend.WorkloadsOrFail(t)[0].PodName())
			}
		}
		return nil
	})

	cluster1GatewayAddress, err := deploymentCtx.Meshes[1].GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
	if err != nil {
		t.Fatal(err)
		return
	}

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
		Count:     15,
		Validator: echo.And(echo.ExpectOK(), frontend1IDValidator),
	})
}

func globalServiceTest(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	cluster0 := ctx.Clusters()[cluster0Index]
	// calling the gateway from outside the mesh
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("no-mesh").And(echo.InCluster(cluster0)))

	frontendHost := "http-frontend.solo.io"
	cluster0GatewayAddress, err := deploymentCtx.Meshes[0].GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
	if err != nil {
		t.Fatal(err)
		return
	}

	frontend := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster0)))
	frontendIDValidator := echo.ValidatorFunc(func(responses client.ParsedResponses, _ error) error {
		for _, r := range responses {
			if r.Hostname != frontend.WorkloadsOrFail(t)[0].PodName() {
				return fmt.Errorf("response from pod %s does not match expected %s", r.Hostname, frontend.WorkloadsOrFail(t)[0].PodName())
			}
		}
		return nil
	})

	// calling the gateway in cluster0-0 for http-frontend.solo.io
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 80,
		},
		Scheme:  scheme.HTTP,
		Address: cluster0GatewayAddress,
		Headers: map[string][]string{
			"Host": {frontendHost},
		},
		Method:    http.MethodGet,
		Path:      "/info",
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), frontendIDValidator),
	})

	// calling cluster0 1 with the same host
	cluster1GatewayAddress, err := deploymentCtx.Meshes[1].GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
	if err != nil {
		t.Fatal(err)
		return
	}

	cluster1 := ctx.Clusters()[cluster1Index]
	frontend1 := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster1)))
	frontend1IDValidator := echo.ValidatorFunc(func(responses client.ParsedResponses, _ error) error {
		for _, r := range responses {
			if r.Hostname != frontend1.WorkloadsOrFail(t)[0].PodName() {
				return fmt.Errorf("response from pod %s does not match expected %s", r.Hostname, frontend.WorkloadsOrFail(t)[0].PodName())
			}
		}
		return nil
	})

	// calling the gateway in cluster0-0 for http-frontend.solo.io
	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 80,
		},
		Scheme:  scheme.HTTP,
		Address: cluster1GatewayAddress,
		Headers: map[string][]string{
			"Host": {frontendHost},
		},
		Method:    http.MethodGet,
		Path:      "/info",
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), frontend1IDValidator),
	})

	// calling from inside the mesh as well

	backendSrc := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("backend").And(echo.InCluster(cluster0)))

	backendSrc.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   frontendHost,
		Method:    http.MethodGet,
		Path:      "/info",
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(cluster0.Name())),
	})
	// cluster1 2 test
	cluster1 = ctx.Clusters()[cluster1Index]

	backendSrc = deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("backend").And(echo.InCluster(cluster1)))
	backendSrc.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "http",
			ServicePort: 8090,
		},
		Scheme:    scheme.HTTP,
		Address:   frontendHost,
		Method:    http.MethodGet,
		Path:      "/info",
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), echo.ExpectCluster(cluster1.Name())),
	})
}

// requests for /frontend apply to all ingressgateways and route to frontend in cluster0
// requests for /backend apply to all ingressgateways and route to backend in cluster0
func prefixTest(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {

	// should reach the frontend
	cluster := ctx.Clusters()[cluster0Index]
	// calling the gateway from outside the mesh
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

	// wrong host
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
