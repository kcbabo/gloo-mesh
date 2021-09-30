package main

import (
	"fmt"
	echo2 "github.com/solo-io/gloo-mesh/pkg/test/apps/echo"
	"net/http"
	"testing"

	"github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"istio.io/istio/pkg/test/echo/client"
	"istio.io/istio/pkg/test/echo/common/scheme"
	"istio.io/istio/pkg/test/framework/components/echo"
	"istio.io/istio/pkg/test/framework/resource"

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
							Name:        "external-service-virtualgateway",
							Description: "testing external service based routing using virtual gateway",
							Test:        externalServiceTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtualgateway.yaml",
							Folder:      "gloo-mesh/virtual-gateway/external-service",
						},
						{
							Name:        "external-service-virtualhost",
							Description: "testing external service based routing using virtual host",
							Test:        externalServiceTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtualhost.yaml",
							Folder:      "gloo-mesh/virtual-gateway/external-service",
						},
						{
							Name:        "external-service-routetable",
							Description: "testing external service based routing using routetable",
							Test:        externalServiceTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "routetable.yaml",
							Folder:      "gloo-mesh/virtual-gateway/external-service",
						},
						{
							Name:        "secure-gateways",
							Description: "testing https based routing using virtualgateway",
							Test:        httpsTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "https-redirect.yaml",
							Folder:      "gloo-mesh/virtual-gateway",
						},
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
							FileName:    "multi-cluster-app.yaml",
							Folder:      "gloo-mesh/virtual-gateway",
						},
						{
							Name:        "global-service",
							Description: "routing internally and externally on the same dns name",
							Test:        globalServiceTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "global-service.yaml",
							Folder:      "gloo-mesh/virtual-gateway",
						},
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
						{
							Name:        "rewrite-using-virtualgateway",
							Description: "testing rewrite based routing using virtualgateway only",
							Test:        rewriteTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtualgateway.yaml",
							Folder:      "gloo-mesh/virtual-gateway/rewrite",
						},
						{
							Name:        "rewrite-using-virtualhost",
							Description: "testing rewrite based routing using virtualhost",
							Test:        rewriteTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "virtualhost.yaml",
							Folder:      "gloo-mesh/virtual-gateway/rewrite",
						},
						{
							Name:        "rewrite-using-routetable",
							Description: "testing rewrite based routing using routetable",
							Test:        rewriteTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "routetable.yaml",
							Folder:      "gloo-mesh/virtual-gateway/rewrite",
						},
						{
							Name:        "rewrite-using-multi-virtualhost",
							Description: "testing rewrite based routing using multiple virtual hosts",
							Test:        rewriteTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "multi-virtualhost.yaml",
							Folder:      "gloo-mesh/virtual-gateway/rewrite",
						},
						{
							Name:        "rewrite-using-multi-routetable",
							Description: "testing rewrite based routing using multiple route tables",
							Test:        rewriteTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "multi-routetable.yaml",
							Folder:      "gloo-mesh/virtual-gateway/rewrite",
						},
						{
							Name:        "rewrite-using-multi-virtualhost-multi-routetable",
							Description: "testing rewrite based routing using route tables per virtualhost",
							Test:        rewriteTest,
							Namespace:   deploymentCtx.EchoContext.AppNamespace.Name(),
							FileName:    "multi-virtualhost-multi-routetable.yaml",
							Folder:      "gloo-mesh/virtual-gateway/rewrite",
						},
					},
				},
			}
			for _, tg := range tgs {
				tg.Run(ctx, t, &deploymentCtx)
			}
		})

}

func getMeshForCluster(clusterName string, deploymentCtx *context.DeploymentContext) context.GlooMeshInstance {
	for _, m := range deploymentCtx.Meshes {
		if m.GetCluster().Name() == clusterName {
			return m
		}
	}
	return nil
}
func externalServiceTest(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	cluster := ctx.Clusters()[cluster0Index]
	// calling the gateway from outside the mesh
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("no-mesh").And(echo.InCluster(cluster)))

	apiHost := "api.solo.io"
	cluster0GatewayAddress, err := getMeshForCluster(cluster.Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
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
		Path:      "/external",
		Count:     5,
		Validator: echo.ExpectOK(),
	})

	cluster1GatewayAddress, err := getMeshForCluster(ctx.Clusters()[cluster1Index].Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway",
		"istio-system", "")
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
		Path:      "/external",
		Count:     5,
		CaCert:    echo2.GetEchoCACert(),
		Validator: echo.ExpectOK(),
	})
}

func httpsTest(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	cluster := ctx.Clusters()[cluster0Index]
	// calling the gateway from outside the mesh
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("no-mesh").And(echo.InCluster(cluster)))

	frontend := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	backend := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("backend").And(echo.InCluster(cluster)))

	apiHost := "api.solo.io"
	cluster0GatewayAddress, err := getMeshForCluster(cluster.Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
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

	// Check for redirect
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
		Validator: echo.ExpectCode("301"),
	})
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
		Validator: echo.ExpectCode("301"),
	})

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
		Path:      "/no-route",
		Count:     5,
		Validator: echo.ExpectCode("301"),
	})

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "https",
			ServicePort: 443,
		},
		Scheme:  scheme.HTTPS,
		Address: cluster0GatewayAddress,
		Headers: map[string][]string{
			"Host": {apiHost},
		},
		CaCert:    echo2.GetEchoCACert(),
		Method:    http.MethodGet,
		Path:      "/frontend",
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), frontendIDValidator),
	})

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "https",
			ServicePort: 443,
		},
		Scheme:  scheme.HTTPS,
		Address: cluster0GatewayAddress,
		Headers: map[string][]string{
			"Host": {apiHost},
		},
		Method:    http.MethodGet,
		Path:      "/backend",
		Count:     5,
		CaCert:    echo2.GetEchoCACert(),
		Validator: echo.And(echo.ExpectOK(), backendIDValidator),
	})

	cluster1GatewayAddress, err := getMeshForCluster(ctx.Clusters()[cluster1Index].Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway",
		"istio-system", "")
	if err != nil {
		t.Fatal(err)
		return
	}

	// Check for redirect
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
		Validator: echo.ExpectCode("301"),
	})
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
		Validator: echo.ExpectCode("301"),
	})

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "https",
			ServicePort: 443,
		},
		Scheme:  scheme.HTTPS,
		Address: cluster1GatewayAddress,
		Headers: map[string][]string{
			"Host": {apiHost},
		},
		Method:    http.MethodGet,
		Path:      "/frontend",
		Count:     5,
		CaCert:    echo2.GetEchoCACert(),
		Validator: echo.And(echo.ExpectOK(), frontendIDValidator),
	})

	src.CallOrFail(t, echo.CallOptions{
		Port: &echo.Port{
			Protocol:    "https",
			ServicePort: 443,
		},
		Scheme:  scheme.HTTPS,
		Address: cluster1GatewayAddress,
		Headers: map[string][]string{
			"Host": {apiHost},
		},
		Method:    http.MethodGet,
		Path:      "/backend",
		Count:     5,
		CaCert:    echo2.GetEchoCACert(),
		Validator: echo.And(echo.ExpectOK(), backendIDValidator),
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
	cluster0GatewayAddress, err := getMeshForCluster(cluster.Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
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

	cluster1GatewayAddress, err := getMeshForCluster(ctx.Clusters()[cluster1Index].Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway",
		"istio-system", "")
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
		Validator: echo.ExpectError(),
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
	cluster0GatewayAddress, err := getMeshForCluster(ctx.Clusters()[cluster0Index].Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway",
		"istio-system", "")
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

	cluster1GatewayAddress, err := getMeshForCluster(ctx.Clusters()[cluster1Index].Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
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
		Validator: echo.And(echo.ExpectOK(), frontend1IDValidator),
	})
}

func globalServiceTest(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {
	cluster0 := ctx.Clusters()[cluster0Index]
	// calling the gateway from outside the mesh
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("no-mesh").And(echo.InCluster(cluster0)))

	frontendHost := "http-frontend.solo.io"
	cluster0GatewayAddress, err := getMeshForCluster(ctx.Clusters()[cluster0Index].Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
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
	cluster1GatewayAddress, err := getMeshForCluster(ctx.Clusters()[cluster1Index].Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
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
			ServicePort: 80,
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
			ServicePort: 80,
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
	cluster0GatewayAddress, err := getMeshForCluster(cluster.Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
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

	cluster1GatewayAddress, err := getMeshForCluster(ctx.Clusters()[cluster1Index].Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
	if err != nil {
		t.Fatal(err)
		return
	}

	// happy requests from cluster-1
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

func rewriteTest(ctx resource.Context, t *testing.T, deploymentCtx *context.DeploymentContext) {

	// should reach the frontend
	cluster := ctx.Clusters()[cluster0Index]
	// calling the gateway from outside the mesh
	src := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("no-mesh").And(echo.InCluster(cluster)))

	frontend := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("frontend").And(echo.InCluster(cluster)))
	backend := deploymentCtx.EchoContext.Deployments.GetOrFail(t, echo.Service("backend").And(echo.InCluster(cluster)))

	apiHost := "api.solo.io"
	cluster0GatewayAddress, err := getMeshForCluster(cluster.Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
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

	pathValidator := func(path string) echo.ValidatorFunc {
		return func(responses client.ParsedResponses, _ error) error {
			for _, r := range responses {
				if r.URL != path {
					return fmt.Errorf("expected path %s but got %s", path, r.URL)
				}
			}
			return nil
		}
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
		Path:      "/api/frontend",
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), frontendIDValidator, pathValidator("/frontend")),
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
		Path:      "/api/frontend",
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
		Path:      "/api/backend",
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), backendIDValidator, pathValidator("/backend")),
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
		Path:      "/api/frontend",
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
		Path:      "/api/wrong",
		Count:     5,
		Validator: echo.ExpectCode("404"),
	})

	cluster1GatewayAddress, err := getMeshForCluster(ctx.Clusters()[cluster1Index].Name(), deploymentCtx).GetIngressGatewayAddress("istio-ingressgateway", "istio-system", "")
	if err != nil {
		t.Fatal(err)
		return
	}

	// happy requests from cluster-1
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
		Path:      "/api/frontend",
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), frontendIDValidator, pathValidator("/frontend")),
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
		Path:      "/api/backend",
		Count:     5,
		Validator: echo.And(echo.ExpectOK(), backendIDValidator, pathValidator("/backend")),
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
		Path:      "/api/frontend",
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
		Path:      "/api/wrong",
		Count:     5,
		Validator: echo.ExpectCode("404"),
	})
}
