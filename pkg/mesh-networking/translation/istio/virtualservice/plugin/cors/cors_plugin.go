package cors

import (
	"github.com/rotisserie/eris"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/equalityutils"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/fieldutils"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
)

const (
	pluginName = "cors"
)

// handles setting Cors on a VirtualService
type corsPlugin struct {
}

func NewCorsPlugin() *corsPlugin {
	return &corsPlugin{
	}
}

func (p *corsPlugin) PluginName() string {
	return pluginName
}

func (p *corsPlugin) ProcessTrafficPolicy(
	appliedPolicy *discoveryv1alpha1.MeshServiceStatus_AppliedTrafficPolicy,
	_ *discoveryv1alpha1.MeshService,
	output *istiov1alpha3spec.HTTPRoute,
	fieldRegistry fieldutils.FieldOwnershipRegistry,
) error {
	cors, err := p.translateCors(appliedPolicy.Spec)
	if err != nil {
		return err
	}
	if cors != nil && !equalityutils.Equals(output.CorsPolicy, cors) {
		if err := fieldRegistry.RegisterFieldOwner(
			&output.CorsPolicy,
			appliedPolicy.Ref,
			0,
		); err != nil {
			return err
		}
		output.CorsPolicy = cors
	}
	return nil
}

func (p *corsPlugin) translateCors(
	trafficPolicy *v1alpha1.TrafficPolicySpec,
) (*istiov1alpha3spec.CorsPolicy, error) {
	corsPolicy := trafficPolicy.CorsPolicy
	if corsPolicy == nil {
		return nil, nil
	}
	var allowOrigins []*istiov1alpha3spec.StringMatch
	for i, allowOrigin := range corsPolicy.GetAllowOrigins() {
		var stringMatch *istiov1alpha3spec.StringMatch
		switch matchType := allowOrigin.GetMatchType().(type) {
		case *v1alpha1.TrafficPolicySpec_StringMatch_Exact:
			stringMatch = &istiov1alpha3spec.StringMatch{MatchType: &istiov1alpha3spec.StringMatch_Exact{Exact: allowOrigin.GetExact()}}
		case *v1alpha1.TrafficPolicySpec_StringMatch_Prefix:
			stringMatch = &istiov1alpha3spec.StringMatch{MatchType: &istiov1alpha3spec.StringMatch_Prefix{Prefix: allowOrigin.GetPrefix()}}
		case *v1alpha1.TrafficPolicySpec_StringMatch_Regex:
			stringMatch = &istiov1alpha3spec.StringMatch{MatchType: &istiov1alpha3spec.StringMatch_Regex{Regex: allowOrigin.GetRegex()}}
		default:
			return nil, eris.Errorf("AllowOrigins[%d].MatchType has unexpected type %T", i, matchType)
		}
		allowOrigins = append(allowOrigins, stringMatch)
	}
	translatedCorsPolicy := &istiov1alpha3spec.CorsPolicy{
		AllowOrigins:     allowOrigins,
		AllowMethods:     corsPolicy.GetAllowMethods(),
		AllowHeaders:     corsPolicy.GetAllowHeaders(),
		ExposeHeaders:    corsPolicy.GetExposeHeaders(),
		MaxAge:           corsPolicy.GetMaxAge(),
		AllowCredentials: corsPolicy.GetAllowCredentials(),
	}

	return translatedCorsPolicy, nil
}