package trafficpolicyutils

import (
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
)

// conversion function to  make it easy to work with the deprecated request matchers
func ConvertDeprecatedRequestMatchers(deprecated []*commonv1.DeprecatedHttpMatcher) []*commonv1.HttpMatcher {
	var converted []*commonv1.HttpMatcher
	for _, match := range deprecated {
		if match == nil {
			continue
		}
		converted = append(converted, convertDeprecatedRequestMatcher(match))
	}
	return converted
}

func convertDeprecatedRequestMatcher(deprecated *commonv1.DeprecatedHttpMatcher) *commonv1.HttpMatcher {
	return &commonv1.HttpMatcher{
		Name:            deprecated.Name,
		Uri:             convertUri(deprecated),
		Headers:         deprecated.Headers,
		QueryParameters: convertQueryParams(deprecated.QueryParameters),
		Method:          deprecated.Method,
	}
}

func convertUri(deprecated *commonv1.DeprecatedHttpMatcher) *commonv1.StringMatch {
	if deprecated.Uri != nil {
		// use new uri if provided
		return deprecated.Uri
	}
	if deprecated.PathSpecifier == nil {
		// no uri provided
		return nil
	}
	m := &commonv1.StringMatch{}
	switch path := deprecated.PathSpecifier.(type) {
	case *commonv1.DeprecatedHttpMatcher_Prefix:
		m.MatchType = &commonv1.StringMatch_Prefix{
			Prefix: path.Prefix,
		}
	case *commonv1.DeprecatedHttpMatcher_Exact:
		m.MatchType = &commonv1.StringMatch_Exact{
			Exact: path.Exact,
		}
	case *commonv1.DeprecatedHttpMatcher_Regex:
		m.MatchType = &commonv1.StringMatch_Regex{
			Regex: path.Regex,
		}
	}
	return m
}

func convertQueryParams(deprecated []*commonv1.DeprecatedHttpMatcher_QueryParameterMatcher) []*commonv1.HttpMatcher_QueryParameterMatcher {
	var converted []*commonv1.HttpMatcher_QueryParameterMatcher
	for _, match := range deprecated {
		if match == nil {
			continue
		}
		converted = append(converted, &commonv1.HttpMatcher_QueryParameterMatcher{
			Name:  match.Name,
			Value: match.Value,
			Regex: match.Regex,
		})
	}
	return converted
}
