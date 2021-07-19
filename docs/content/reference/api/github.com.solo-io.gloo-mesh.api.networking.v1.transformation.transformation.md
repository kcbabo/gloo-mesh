
---

title: "transformation.proto"

---

## Package : `transformation.networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for transformation.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## transformation.proto


## Table of Contents
  - [GatewayTransformations](#transformation.networking.mesh.gloo.solo.io.GatewayTransformations)
  - [InjaTemplateTransformation](#transformation.networking.mesh.gloo.solo.io.InjaTemplateTransformation)
  - [RequestTransformation](#transformation.networking.mesh.gloo.solo.io.RequestTransformation)
  - [ResponseTransformation](#transformation.networking.mesh.gloo.solo.io.ResponseTransformation)
  - [ResponseTransformation.ResponseMatcher](#transformation.networking.mesh.gloo.solo.io.ResponseTransformation.ResponseMatcher)
  - [RouteTransformations](#transformation.networking.mesh.gloo.solo.io.RouteTransformations)
  - [TextTransformation](#transformation.networking.mesh.gloo.solo.io.TextTransformation)
  - [XsltTransformation](#transformation.networking.mesh.gloo.solo.io.XsltTransformation)







<a name="transformation.networking.mesh.gloo.solo.io.GatewayTransformations"></a>

### GatewayTransformations
GatewayTransformation enables use of the Transformation feature on Gateway.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enabled | bool |  | Explicitly enable Transformations on a gateway. Only required if strict filter management is set on the gateway. |
  





<a name="transformation.networking.mesh.gloo.solo.io.InjaTemplateTransformation"></a>

### InjaTemplateTransformation
transform HTTP body and headers using Inja templates.<br>TODO: implement






<a name="transformation.networking.mesh.gloo.solo.io.RequestTransformation"></a>

### RequestTransformation
match and transform the contents of an HTTP Request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| matchers | [][common.mesh.gloo.solo.io.HttpMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.request_matchers#common.mesh.gloo.solo.io.HttpMatcher" >}}) | repeated | Specify criteria that HTTP requests must satisfy for the RequestTransformation to apply Omit to apply to any HTTP request. |
  | transformation | [transformation.networking.mesh.gloo.solo.io.TextTransformation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.transformation.transformation#transformation.networking.mesh.gloo.solo.io.TextTransformation" >}}) |  | the text transformation to apply to to the matched request |
  | recalculateRoutingDestination | bool |  | If the request was transformed such that it would match a different route within the same Gateway, recalculate the routing destination (select a new route) based on the transformed content of the request. |
  | applyBeforeAuth | bool |  | Apply this transformation before Auth and Rate Limit checks are performed on the request. This can be used to modify the request headers before they are captured by the ExtAuth & Rate Limiter services. |
  





<a name="transformation.networking.mesh.gloo.solo.io.ResponseTransformation"></a>

### ResponseTransformation
match and transform the contents of an HTTP Response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| matchers | [][transformation.networking.mesh.gloo.solo.io.ResponseTransformation.ResponseMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.transformation.transformation#transformation.networking.mesh.gloo.solo.io.ResponseTransformation.ResponseMatcher" >}}) | repeated | Match elements of the Response in order to apply a response transformation. If no response matchers are specified, the transformation will always be applied. |
  | transformation | [transformation.networking.mesh.gloo.solo.io.TextTransformation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.transformation.transformation#transformation.networking.mesh.gloo.solo.io.TextTransformation" >}}) |  | the text transformation to apply to to the matched response |
  





<a name="transformation.networking.mesh.gloo.solo.io.ResponseTransformation.ResponseMatcher"></a>

### ResponseTransformation.ResponseMatcher
specifies a set of criteria for matching an HTTP response


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| headers | [][common.mesh.gloo.solo.io.HeaderMatcher]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.common.v1.request_matchers#common.mesh.gloo.solo.io.HeaderMatcher" >}}) | repeated | Specify a set of response headers which must match in entirety (all headers must match). |
  | responseCodeDetails | string |  | Response code detail to match on. To see the response code details for your usecase, you can use the envoy access log %RESPONSE_CODE_DETAILS% formatter to log it. |
  





<a name="transformation.networking.mesh.gloo.solo.io.RouteTransformations"></a>

### RouteTransformations
A RouteTransformation defines a text transformations for the content of an HTTP Request and/or Response on a matched route. Transformation takes the existing HTTP Headers (and, optionally, Body) and transforms them into a new set of headers (and body). Transformations can be used on outbound Request data as well as inbound Response data. Various types of transformations can be performed depending on the format of the input and output data types. Currently, Inja (for JSON-to-JSON transformations) and XSLT (for JSON-to-XML and XML-to-XML transformations) are currently supported. Transformations can optionally define a set of request/response HTTP match criteria. The first matched transformation in a list will be applied to the HTTP request/response.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request | [][transformation.networking.mesh.gloo.solo.io.RequestTransformation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.transformation.transformation#transformation.networking.mesh.gloo.solo.io.RequestTransformation" >}}) | repeated | Transformations to apply on the outbound HTTP request before it arrives at the routing Destination. Only the first matched transformation will be applied to the request. |
  | response | [][transformation.networking.mesh.gloo.solo.io.ResponseTransformation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.transformation.transformation#transformation.networking.mesh.gloo.solo.io.ResponseTransformation" >}}) | repeated | Transformations to apply on the inbound HTTP request before it returns to the HTTP client (traffic source). Only the first matched transformation will be applied to the response. |
  





<a name="transformation.networking.mesh.gloo.solo.io.TextTransformation"></a>

### TextTransformation
Transform the HTTP Headers/Body of a request / response using one of the supported mechanisms


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| xsltTransformation | [transformation.networking.mesh.gloo.solo.io.XsltTransformation]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.transformation.transformation#transformation.networking.mesh.gloo.solo.io.XsltTransformation" >}}) |  | transform HTTP body using XSLT styling language. |
  





<a name="transformation.networking.mesh.gloo.solo.io.XsltTransformation"></a>

### XsltTransformation
transform HTTP body using XSLT styling language.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| xslt | string |  | XSLT transformation template which you want to transform requests/responses with. Invalid XSLT transformation templates will result will result in an invalid route. |
  | setContentType | string |  | Changes the content-type header of the HTTP request/response to what is set here. This is useful in situations where an XSLT transformation is used to transform XML to JSON and the content-type should be changed from `application/xml` to `application/json`. If left empty, the content-type header remains unmodified by default. |
  | nonXmlTransform | bool |  | This should be set to true if the content being transformed is not XML. For example, if the content being transformed is from JSON to XML, this should be set to true. XSLT transformations can only take valid XML as input to be transformed. If the body is not a valid XML (e.g. using JSON as input in a JSON-to-XML transformation), setting `non_xml_transform` to true will allow the XSLT to accept the non-XML input without throwing an error by passing the input as XML CDATA. defaults to false. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

