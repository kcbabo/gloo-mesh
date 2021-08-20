package trafficpolicyutils

import (
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
)

// conversion function to  make it easy to work with the deprecated request matchers
func ConvertDeprecatedRequestMatchers(deprecated []*v1.DeprecatedHttpMatcher) []*v1.HttpMatcher {
	var converted []*v1.HttpMatcher
	for _, match := range deprecated {
		if match == nil {
			continue
		}
		converted = append(converted, convertDeprecatedRequestMatcher(match))
	}
	return converted
}

func convertDeprecatedRequestMatcher(deprecated *v1.DeprecatedHttpMatcher) *v1.HttpMatcher {
	return &v1.HttpMatcher{
		Name:            deprecated.GetName(),
		Uri:             convertUri(deprecated),
		Headers:         deprecated.GetHeaders(),
		QueryParameters: convertQueryParams(deprecated.GetQueryParameters()),
		Method:          deprecated.GetMethod(),
	}
}

func convertUri(deprecated *v1.DeprecatedHttpMatcher) *commonv1.StringMatch {
	if deprecated.GetUri() != nil {
		// use new uri if provided
		return deprecated.GetUri()
	}
	if deprecated.GetPathSpecifier() == nil {
		// no uri provided
		return nil
	}
	m := &commonv1.StringMatch{}
	switch path := deprecated.GetPathSpecifier().(type) {
	case *v1.DeprecatedHttpMatcher_Prefix:
		m.MatchType = &commonv1.StringMatch_Prefix{
			Prefix: path.Prefix,
		}
	case *v1.DeprecatedHttpMatcher_Exact:
		m.MatchType = &commonv1.StringMatch_Exact{
			Exact: path.Exact,
		}
	case *v1.DeprecatedHttpMatcher_Regex:
		m.MatchType = &commonv1.StringMatch_Regex{
			Regex: path.Regex,
		}
	}
	return m
}

func convertQueryParams(deprecated []*v1.DeprecatedHttpMatcher_QueryParameterMatcher) []*v1.HttpMatcher_QueryParameterMatcher {
	var converted []*v1.HttpMatcher_QueryParameterMatcher
	for _, match := range deprecated {
		if match == nil {
			continue
		}
		converted = append(converted, &v1.HttpMatcher_QueryParameterMatcher{
			Name:  match.GetName(),
			Value: match.GetValue(),
			Regex: match.GetRegex(),
		})
	}
	return converted
}
