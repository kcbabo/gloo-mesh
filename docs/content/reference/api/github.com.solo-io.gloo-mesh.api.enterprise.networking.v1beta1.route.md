
---

title: "route.proto"

---

## Package : `networking.enterprise.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for route.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## route.proto


## Table of Contents
  - [DelegateAction](#networking.enterprise.mesh.gloo.solo.io.DelegateAction)
  - [DirectResponseAction](#networking.enterprise.mesh.gloo.solo.io.DirectResponseAction)
  - [RedirectAction](#networking.enterprise.mesh.gloo.solo.io.RedirectAction)
  - [Route](#networking.enterprise.mesh.gloo.solo.io.Route)
  - [Route.LabelsEntry](#networking.enterprise.mesh.gloo.solo.io.Route.LabelsEntry)
  - [Route.RouteAction](#networking.enterprise.mesh.gloo.solo.io.Route.RouteAction)

  - [DelegateAction.SortMethod](#networking.enterprise.mesh.gloo.solo.io.DelegateAction.SortMethod)
  - [RedirectAction.RedirectResponseCode](#networking.enterprise.mesh.gloo.solo.io.RedirectAction.RedirectResponseCode)






<a name="networking.enterprise.mesh.gloo.solo.io.DelegateAction"></a>

### DelegateAction
Note: This message needs to be at this level (rather than nested) due to cue restrictions. DelegateActions are used to delegate routing decisions to other resources, for example RouteTables.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| refs | [][core.skv2.solo.io.ObjectRef]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectRef" >}}) | repeated | Delegate to the RouteTable resources with matching `name` and `namespace`. |
  | selector | [core.skv2.solo.io.ObjectSelector]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.skv2.api.core.v1.core#core.skv2.solo.io.ObjectSelector" >}}) |  | Delegate to the RouteTables that match the given selector. Selected route tables are ordered by creation time stamp in ascending order to guarantee consistent ordering. |
  | sortMethod | [networking.enterprise.mesh.gloo.solo.io.DelegateAction.SortMethod]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.DelegateAction.SortMethod" >}}) |  | How routes should be sorted |
  





<a name="networking.enterprise.mesh.gloo.solo.io.DirectResponseAction"></a>

### DirectResponseAction
Note: This message needs to be at this level (rather than nested) due to cue restrictions. DirectResponseAction is copied directly from https://github.com/envoyproxy/envoy/blob/master/api/envoy/api/v2/route/route.proto


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | uint32 |  | Specifies the HTTP response status to be returned. |
  | body | string |  | Specifies the content of the response body. If this setting is omitted, no body is included in the generated response.<br>Note: Headers can be specified using the Header Modification feature in the enclosing Route, ConnectionHandler, or Gateway options. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.RedirectAction"></a>

### RedirectAction
Note: This message needs to be at this level (rather than nested) due to cue restrictions. Notice: RedirectAction is copied directly from https://github.com/envoyproxy/envoy/blob/master/api/envoy/api/v2/route/route.proto


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hostRedirect | string |  | The host portion of the URL will be swapped with this value. |
  | pathRedirect | string |  | The path portion of the URL will be swapped with this value. |
  | prefixRewrite | string |  | Indicates that during redirection, the matched prefix (or path) should be swapped with this value. This option allows redirect URLs be dynamically created based on the request.<br>  Pay attention to the use of trailing slashes as mentioned in   `RouteAction`'s `prefix_rewrite`. |
  | responseCode | [networking.enterprise.mesh.gloo.solo.io.RedirectAction.RedirectResponseCode]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.RedirectAction.RedirectResponseCode" >}}) |  | The HTTP status code to use in the redirect response. The default response code is MOVED_PERMANENTLY (301). |
  | httpsRedirect | bool |  | The scheme portion of the URL will be swapped with "https". |
  | stripQuery | bool |  | Indicates that during redirection, the query portion of the URL will be removed. Default value is false. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.Route"></a>

### Route
A route specifies how to match a request and what action to take when the request is matched.<br>When a request matches on a route, the route can perform one of the following actions: - *Route* the request to a destination - Reply with a *Direct Response* - Send a *Redirect* response to the client - *Delegate* the action for the request to one or more [`RouteTable`]({{< ref "/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route_table.md" >}}) resources DelegateActions can be used to delegate the behavior for a set out routes to `RouteTable` resources.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The name provides a convenience for users to be able to refer to a route by name. It includes names of VS, Route, and RouteTable ancestors of the Route. |
  | matchers | [][networking.mesh.gloo.solo.io.HttpMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.request_matchers#networking.mesh.gloo.solo.io.HttpMatcher" >}}) | repeated | Matchers contain parameters for matching requests (i.e., based on HTTP path, headers, etc.). If empty, the route will match all requests (i.e, a single "/" path prefix matcher). For delegated routes, any parent matcher must have a `prefix` path matcher. |
  | routeAction | [networking.enterprise.mesh.gloo.solo.io.Route.RouteAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.Route.RouteAction" >}}) |  | This action is the primary action to be selected for most routes. The RouteAction tells the proxy to route requests to an upstream. |
  | redirectAction | [networking.enterprise.mesh.gloo.solo.io.RedirectAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.RedirectAction" >}}) |  | Redirect actions tell the proxy to return a redirect response to the downstream client. |
  | directResponseAction | [networking.enterprise.mesh.gloo.solo.io.DirectResponseAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.DirectResponseAction" >}}) |  | Return an arbitrary HTTP response directly, without proxying. |
  | delegateAction | [networking.enterprise.mesh.gloo.solo.io.DelegateAction]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.DelegateAction" >}}) |  | Delegate routing actions for the given matcher to one or more RouteTables. |
  | options | [networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy#networking.mesh.gloo.solo.io.TrafficPolicySpec.Policy" >}}) |  | Route Options extend the behavior of routes. Route options include configuration such as retries, rate limiting, and request/response transformation. RouteOption behavior will be inherited by delegated routes which do not specify their own `options` |
  | labels | [][networking.enterprise.mesh.gloo.solo.io.Route.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route#networking.enterprise.mesh.gloo.solo.io.Route.LabelsEntry" >}}) | repeated | Specify labels for this route, which are used by other resources (e.g. TrafficPolicy) to select specific routes within a given gateway object. |
  





<a name="networking.enterprise.mesh.gloo.solo.io.Route.LabelsEntry"></a>

### Route.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="networking.enterprise.mesh.gloo.solo.io.Route.RouteAction"></a>

### Route.RouteAction
RouteActions are used to route matched requests to upstreams.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| destinations | [][networking.mesh.gloo.solo.io.WeightedDestination]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.weighed_destination#networking.mesh.gloo.solo.io.WeightedDestination" >}}) | repeated | Defines the destination upstream for routing Some destinations require additional configuration for the route (e.g. AWS upstreams require a function name to be specified). |
  | pathRewrite | string |  | Replace the path specified in the matcher with this value before passing upstream. When a prefix matcher is used, only the prefix portion of the path is rewritten. When an exact matcher is used, the whole path is replaced. Rewriting the path when a regex matcher is used is currently unsupported. |
  




 <!-- end messages -->


<a name="networking.enterprise.mesh.gloo.solo.io.DelegateAction.SortMethod"></a>

### DelegateAction.SortMethod


| Name | Number | Description |
| ---- | ------ | ----------- |
| TABLE_WEIGHT | 0 | Routes are kept in the order that they appear relative to their tables, but tables are sorted by weight. Tables that have the same weight will stay in the same order that they are listed in, which is the list order when given as a reference and by creation timestamp when selected. |
| ROUTE_SPECIFICITY | 1 | After processing all routes, including additional route tables delegated to, the resulting routes are sorted by specificity to reduce the chance that a more specific route will be short-circuited by a general route. Matchers with exact path matchers are considered more specific than regex path patchers, which are more specific than prefix path matchers. Matchers of the same type are sorted by length of the path in descending order. Only the most specific matcher on each route is used. |



<a name="networking.enterprise.mesh.gloo.solo.io.RedirectAction.RedirectResponseCode"></a>

### RedirectAction.RedirectResponseCode


| Name | Number | Description |
| ---- | ------ | ----------- |
| MOVED_PERMANENTLY | 0 | Moved Permanently HTTP Status Code - 301. |
| FOUND | 1 | Found HTTP Status Code - 302. |
| SEE_OTHER | 2 | See Other HTTP Status Code - 303. |
| TEMPORARY_REDIRECT | 3 | Temporary Redirect HTTP Status Code - 307. |
| PERMANENT_REDIRECT | 4 | Permanent Redirect HTTP Status Code - 308. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

