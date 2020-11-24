package traffictarget

import (
	"context"
	"fmt"

	"github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	"github.com/solo-io/go-utils/contextutils"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/appmesh"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	v1alpha2types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/types"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/workloadutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/errors"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -source ./appmesh_traffic_target_translator.go -destination mocks/appmesh_traffic_target_translator.go

// Translator translator translates a TrafficTarget into a equivalent appmesh config.
type Translator interface {
	// Output resources will configure the underlying appmesh mesh.
	// Errors caused by invalid user config will be reported using the Reporter.
	Translate(
		ctx context.Context,
		in input.Snapshot,
		trafficTarget *discoveryv1alpha2.TrafficTarget,
		outputs appmesh.Builder,
		reporter reporting.Reporter,
	)
}

type translator struct{}

func NewTranslator() Translator {
	return &translator{}
}

// translate the appropriate resources for the given TrafficTarget.
func (t *translator) Translate(
	ctx context.Context,
	in input.Snapshot,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	outputs appmesh.Builder,
	reporter reporting.Reporter,
) {
	// Only translate appmesh TrafficTargets.
	if !isAppmeshTrafficTarget(ctx, trafficTarget, in.Meshes()) {
		return
	}

	// AppMesh doesn't support all policies; report those which aren't implemented.
	for _, tp := range trafficTarget.Status.GetAppliedTrafficPolicies() {
		report(tp, trafficTarget, reporter)
	}

	virtualService, virtualRouter := translate(trafficTarget, in.Workloads())
	outputs.AddVirtualServices(virtualService)
	outputs.AddVirtualRouters(virtualRouter)
}

func translate(trafficTarget *discoveryv1alpha2.TrafficTarget, workloads v1alpha2sets.WorkloadSet) (*v1beta2.VirtualService, *v1beta2.VirtualRouter) {
	meta := metautils.TranslatedObjectMeta(
		trafficTarget.Spec.GetKubeService().Ref,
		trafficTarget.Annotations,
	)

	vr := getVirtualRouter(meta, trafficTarget)

	// TODO joekelley error-handling
	backingWorkloads := workloadutils.FindBackingWorkloads(trafficTarget.Spec.GetKubeService(), workloads)
	workload := backingWorkloads[0]
	arn := workload.Spec.AppMesh.VirtualNodeArn

	vs := getVirtualService(meta, vr, arn)

	return vs, vr
}

func getVirtualService(
	meta metav1.ObjectMeta,
	virtualRouter *v1beta2.VirtualRouter,
	arn string,
) *v1beta2.VirtualService {
	var provider *v1beta2.VirtualServiceProvider
	if virtualRouter != nil {
		provider = &v1beta2.VirtualServiceProvider{
			VirtualRouter: &v1beta2.VirtualRouterServiceProvider{
				VirtualRouterRef: &v1beta2.VirtualRouterReference{
					Namespace: &meta.Namespace,
					Name:      meta.Name,
				},
			},
		}
	} else {
		provider = &v1beta2.VirtualServiceProvider{
			VirtualNode: &v1beta2.VirtualNodeServiceProvider{
				VirtualNodeARN: &arn,
			},
		}
	}

	// This is the default name from the AWS controller.
	// We must provide it explicitly, else the App Mesh controller's
	// validating admission webhook will reject our changes on update.
	awsName := fmt.Sprintf("%s.%s", meta.Name, meta.Namespace)
	return &v1beta2.VirtualService{
		ObjectMeta: meta,
		Spec: v1beta2.VirtualServiceSpec{
			AWSName:  &awsName,
			Provider: provider,
		},
	}
}

func getVirtualRouter(meta metav1.ObjectMeta, trafficTarget *discoveryv1alpha2.TrafficTarget) *v1beta2.VirtualRouter {
	routes := getRoutes(trafficTarget)
	if len(routes) == 0 {
		// There are no routes, so we don't need to create a virtual router
		return nil
	}

	return &v1beta2.VirtualRouter{
		ObjectMeta: meta,
		Spec: v1beta2.VirtualRouterSpec{
			Listeners: getVirtualRouterListeners(trafficTarget),
			Routes:    routes,
		},
	}
}

func getVirtualRouterListeners(trafficTarget *discoveryv1alpha2.TrafficTarget) []v1beta2.VirtualRouterListener {
	var listeners []v1beta2.VirtualRouterListener
	for _, port := range trafficTarget.Spec.GetKubeService().Ports {
		listener := v1beta2.VirtualRouterListener{
			PortMapping: v1beta2.PortMapping{
				Port:     v1beta2.PortNumber(port.Port),
				Protocol: v1beta2.PortProtocol(port.Protocol),
			},
		}
		listeners = append(listeners, listener)
	}
	return listeners
}

func getRoutes(trafficTarget *discoveryv1alpha2.TrafficTarget) []v1beta2.Route {
	var routes []v1beta2.Route
	for _, tp := range trafficTarget.Status.GetAppliedTrafficPolicies() {
		routes = append(routes, getTrafficPolicyRoutes(tp.Ref, tp.Spec)...)
	}
	return routes
}

func convertHeaders(in []*v1alpha2.TrafficPolicySpec_HeaderMatcher) []v1beta2.HTTPRouteHeader {
	var output []v1beta2.HTTPRouteHeader
	for _, headerMatcher := range in {
		headerValue := headerMatcher.Value
		invert := headerMatcher.InvertMatch

		matchMethod := &v1beta2.HeaderMatchMethod{}
		if headerMatcher.Regex {
			matchMethod.Regex = &headerValue
		} else {
			matchMethod.Exact = &headerValue
		}

		output = append(output, v1beta2.HTTPRouteHeader{
			Name:   headerMatcher.Name,
			Match:  matchMethod,
			Invert: &invert,
		})
	}
	return output
}

func convertMethod(in *v1alpha2.TrafficPolicySpec_HttpMethod) *string {
	var str string
	switch in.Method {
	case v1alpha2types.HttpMethodValue_GET:
		str = "GET"
	case v1alpha2types.HttpMethodValue_POST:
		str = "POST"
	case v1alpha2types.HttpMethodValue_PUT:
		str = "PUT"
	case v1alpha2types.HttpMethodValue_DELETE:
		str = "DELETE"
	case v1alpha2types.HttpMethodValue_HEAD:
		str = "HEAD"
	case v1alpha2types.HttpMethodValue_CONNECT:
		str = "CONNECT"
	case v1alpha2types.HttpMethodValue_OPTIONS:
		str = "OPTIONS"
	case v1alpha2types.HttpMethodValue_TRACE:
		str = "TRACE"
	case v1alpha2types.HttpMethodValue_PATCH:
		str = "PATCH"
	default:
		return nil
	}
	return &str
}

func getTrafficPolicyRoutes(trafficPolicyRef *v1.ObjectRef, trafficPolicy *v1alpha2.TrafficPolicySpec) []v1beta2.Route {
	getMatches := func(networkingMatchers []*v1alpha2.TrafficPolicySpec_HttpMatcher) []v1beta2.HTTPRouteMatch {
		var httpRouteMatches []v1beta2.HTTPRouteMatch

		for _, nm := range networkingMatchers {
			if nm.GetPrefix() == "" {
				// TODO report any non-prefix matchers as they're not supported by app mesh
				continue
			}

			httpRouteMatches = append(httpRouteMatches, v1beta2.HTTPRouteMatch{
				Headers: convertHeaders(nm.Headers),
				Method:  convertMethod(nm.Method),
				Prefix:  nm.GetPrefix(),
				Scheme:  nil,
			})
		}

		return httpRouteMatches
	}

	getRouteAction := func() v1beta2.HTTPRouteAction {
		// If there is no traffic shift, route all traffic to the virtual node backing this traffic target.
		if trafficPolicy.GetTrafficShift() == nil {
			var virtualNodeArn string
			return v1beta2.HTTPRouteAction{
				WeightedTargets: []v1beta2.WeightedTarget{{
					VirtualNodeARN: &virtualNodeArn,
					Weight:         1,
				}},
			}
		}

		var weightedTargets []v1beta2.WeightedTarget
		for _, destination := range trafficPolicy.GetTrafficShift().GetDestinations() {
			// TODO joekelley report on anything but kube service

			destination.GetKubeService()

			var workloads v1alpha2sets.WorkloadSet
			var arn string

			// TODO
			// - get traffic target for kube service destination
			// - get backing workloads for traffic target
			// - get ARN from workload

			// TODO joekelley get virtual node info for each backing service
			// kubeservice -> workloads -> aws info

			weightedTargets = append(weightedTargets, v1beta2.WeightedTarget{
				VirtualNodeARN: &arn,
				Weight:         int64(destination.Weight),
			})
		}

		return v1beta2.HTTPRouteAction{
			WeightedTargets: weightedTargets,
		}
	}

	getRetryPolicy := func() *v1beta2.HTTPRetryPolicy {
		if trafficPolicy.Retries == nil {
			return nil
		}

		var perRetryTimeout v1beta2.Duration
		if trafficPolicy.Retries.PerTryTimeout != nil {
			perRetryTimeout.Value = trafficPolicy.Retries.PerTryTimeout.Seconds
			perRetryTimeout.Unit = v1beta2.DurationUnitS
		}

		// Use all supported HTTP and TCP retry events.
		return &v1beta2.HTTPRetryPolicy{
			HTTPRetryEvents: []v1beta2.HTTPRetryPolicyEvent{"server-error", "gateway-error", "client-error", "stream-error"},
			TCPRetryEvents:  []v1beta2.TCPRetryPolicyEvent{"connection-error"},
			MaxRetries:      int64(trafficPolicy.Retries.Attempts),
			PerRetryTimeout: perRetryTimeout,
		}
	}

	getTimeoutPolicy := func() *v1beta2.HTTPTimeout {
		if trafficPolicy.RequestTimeout == nil {
			return nil
		}

		return &v1beta2.HTTPTimeout{
			PerRequest: &v1beta2.Duration{
				Unit:  v1beta2.DurationUnitS,
				Value: trafficPolicy.RequestTimeout.Seconds,
			},
		}
	}

	var routes []v1beta2.Route
	for i, routeMatch := range getMatches(trafficPolicy.HttpRequestMatchers) {
		routes = append(routes, v1beta2.Route{
			Name: fmt.Sprintf("%s-%s-%d", trafficPolicyRef.Namespace, trafficPolicyRef.Name, i),
			// TODO implement the other route types
			HTTPRoute: &v1beta2.HTTPRoute{
				Match:       routeMatch,
				Action:      getRouteAction(),
				RetryPolicy: getRetryPolicy(),
				Timeout:     getTimeoutPolicy(),
			},
		})
	}

	return routes
}

func isAppmeshTrafficTarget(
	ctx context.Context,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	allMeshes v1alpha2sets.MeshSet,
) bool {
	meshRef := trafficTarget.Spec.Mesh
	if meshRef == nil {
		contextutils.LoggerFrom(ctx).Errorf("internal error: trafficTarget %v missing mesh ref", sets.Key(trafficTarget))
		return false
	}
	mesh, err := allMeshes.Find(meshRef)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorf("internal error: could not find mesh %v for trafficTarget %v", sets.Key(meshRef), sets.Key(trafficTarget))
		return false
	}

	return mesh.Spec.GetAwsAppMesh() != nil
}

func report(
	tp *discoveryv1alpha2.TrafficTargetStatus_AppliedTrafficPolicy,
	trafficTarget *discoveryv1alpha2.TrafficTarget,
	reporter reporting.Reporter,
) {
	getMessage := func(feature string) string {
		return fmt.Sprintf("Service Mesh Hub does not support %s for AppMesh", feature)
	}

	// TODO joekelley add mTLS here

	if tp.GetSpec().GetCorsPolicy() != nil {
		reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, tp.GetRef(), errors.NewUnsupportedFeatureError(
			tp.GetRef(),
			"CorsPolicy",
			getMessage("CorsPolicy"),
		))
	}
	if tp.GetSpec().GetFaultInjection() != nil {
		reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, tp.GetRef(), errors.NewUnsupportedFeatureError(
			tp.GetRef(),
			"FaultInjection",
			getMessage("FaultInjection"),
		))
	}
	if tp.GetSpec().GetHeaderManipulation() != nil {
		reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, tp.GetRef(), errors.NewUnsupportedFeatureError(
			tp.GetRef(),
			"HeaderManipulation",
			getMessage("HeaderManipulation"),
		))
	}
	if tp.GetSpec().GetMirror() != nil {
		reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, tp.GetRef(), errors.NewUnsupportedFeatureError(
			tp.GetRef(),
			"Mirror",
			getMessage("Mirror"),
		))
	}
	if tp.GetSpec().GetRequestTimeout() != nil {
		reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, tp.GetRef(), errors.NewUnsupportedFeatureError(
			tp.GetRef(),
			"RequestTimeout",
			getMessage("RequestTimeout"),
		))
	}
	if tp.GetSpec().GetSourceSelector() != nil {
		reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, tp.GetRef(), errors.NewUnsupportedFeatureError(
			tp.GetRef(),
			"SourceSelector",
			getMessage("SourceSelector"),
		))
	}
}