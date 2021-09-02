---
title: Advanced Routing
weight: 50
description: How to configure and use multiple route tables
---

This guide will walk you through setting up your `VirtualGateway` with multiple external `RouteTable`s in order to get
more fine-grained control over how routes are organized.

## Before You Begin

Before you begin, ensure that your setup for the management cluster (`cluster-1`) and a remote cluster (`cluster-2`) meets all of the following prerequisites:
  * Istio:
    * Istio is [installed in the `istio-system` namespace on both clusters]({{% versioned_link_path fromRoot="/guides/installing_istio" %}})
    * `istio-ingressgateway` is deployed in the `istio-system` namespace of `cluster-1`, which is the default installation namespace
    * The `bookinfo` app is [installed in the `bookinfo` namespace of both clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})
  * Gloo Mesh:
    * Gloo Mesh Enterprise is [installed in relay mode in the `gloo-mesh` namespace of `cluster-1`]({{% versioned_link_path fromRoot="/setup/installation/enterprise_installation/" %}})
    * `enterprise-networking` is deployed in the `gloo-mesh` namespace of `cluster-1` and exposes its gRPC server on port 9900
    * `enterprise-agent` is deployed on both clusters and exposes its gRPC server on port 9977
    * Both `cluster-1` and `cluster-2` are [registered with Gloo Mesh]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
    * The following environment variables are set:
      ```shell
      CONTEXT_1=cluster_1_context
      CONTEXT_2=cluster_2_context
      ```

To understand the custom resources that are used in this guide, see the [Gateway Concepts Overview]({{% versioned_link_path fromRoot="/guides/gateway/concepts" %}}).

## Setting up a Basic VirtualGateway and VirtualHosts

As a starting point, create the following `VirtualGateway` and `VirtualHost`s to route to the reviews and ratings
services.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml">}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
  labels:
    app: bookinfo
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews
    name: reviews
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: reviews
          namespace: bookinfo
---
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  name: demo-gateway
  namespace: gloo-mesh
spec:
  ingressGatewaySelectors:
  - portName: http2
    destinationSelectors:
    - kubeServiceMatcher:
        clusters:
        - mgmt-cluster
        labels:
          istio: ingressgateway
        namespaces:
        - istio-system
  connectionHandlers:
  - http:
      routeConfig:
      - virtualHostSelector:
          namespaces:
          - gloo-mesh
          labels:
            app: bookinfo
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
  labels:
    app: bookinfo
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews
    name: reviews
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: reviews
          namespace: bookinfo
---
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  name: demo-gateway
  namespace: gloo-mesh
spec:
  ingressGatewaySelectors:
  - portName: http2
    destinationSelectors:
    - kubeServiceMatcher:
        clusters:
        - cluster-1
        labels:
          istio: ingressgateway
        namespaces:
        - istio-system
  connectionHandlers:
  - http:
      routeConfig:
      - virtualHostSelector:
          namespaces:
          - gloo-mesh
EOF
{{< /tab >}}
{{< /tabs >}}

Verify that the `istio-ingressgateway` is exposed as a `LoadBalancer` service with an external IP address.

```shell
kubectl get service -n istio-system
```

In this example output, the **EXTERNAL_IP** address is `32.12.34.555`:

{{< highlight shell "hl_lines=2" >}}
NAME                   TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)                                                                      AGE
istio-ingressgateway   LoadBalancer   10.96.229.177   32.12.34.555   15021:31911/TCP,80:30166/TCP,443:32302/TCP,15012:30471/TCP,15443:31931/TCP   10m
istiod                 ClusterIP      10.96.180.254   <none>         15010/TCP,15012/TCP,443/TCP,15014/TCP                                        10m
{{< /highlight >}}

If using an alternative environment such as `KinD` that does not support `LoadBalancer` services, you can forward the
`http2` port of the `istio-ingressgateway` deployment.

```shell
kubectl --context $CONTEXT_1 -n istio-system port-forward deploy/istio-ingressgateway 8080
```

Now the ratings service should be available, which can be confirmed via an HTTP call:

```shell
curl -H "Host: www.example.com" localhost:8081/ratings/1
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}
```

### Adding Routes

Next, we will be updating the `VirtualHost` to route to both versions of the reviews service via subset routing. This
will let you request a specific version of the reviews service via the route. 

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml">}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews/v1
    name: reviews-v1
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: reviews
          namespace: bookinfo
          subset:
            version: v1
      pathRewrite: /reviews
  - matchers:
    - uri:
        prefix: /reviews/v2
    name: reviews-v2
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: reviews
          namespace: bookinfo
          subset:
            version: v2
      pathRewrite: /reviews
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews/v1
    name: reviews-v1
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: reviews
          namespace: bookinfo
          subset:
            version: v1
      pathRewrite: /reviews
  - matchers:
    - uri:
        prefix: /reviews/v2
    name: reviews-v2
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: reviews
          namespace: bookinfo
          subset:
            version: v2
      pathRewrite: /reviews
EOF
{{< /tab >}}
{{< /tabs >}}

Now, you can test that both versions are accessible by sending requests to the new versioned endpoints:

```shell
curl -H "Host: www.example.com" localhost:8081/reviews/v1/1
curl -H "Host: www.example.com" localhost:8081/reviews/v2/1
```

Both endpoints should return the same review JSON, but the v2 endpoint will also have a rating.


## Splitting Out the Routes

Now that multiple routes are set up for the `/reviews` prefix, you can create one `RouteTable` resource to split the
routes into respective route tables for better organization. Route tables also provide more fine-grained access control
because authorized users can have permission to edit specific route tables without having permission to edit the
`VirtualHost` or other route tables. Start by creating a `RouteTable` resource with the `v1` and `v2` routes. Note that
the `/reviews` prefix is omitted.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml">}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: demo-routetable
  namespace: gloo-mesh
spec:
  routes:
  - matchers:
    - uri:
        prefix: /v1
    name: v1
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews
          namespace: bookinfo
          subset:
            version: v1
      pathRewrite: /reviews
  - matchers:
    - uri:
        prefix: /v2
    name: v2
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews
          namespace: bookinfo
          subset:
            version: v2
      pathRewrite: /reviews
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: demo-routetable
  namespace: gloo-mesh
spec:
  routes:
  - matchers:
    - uri:
        prefix: /v1
    name: v1
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews
          namespace: bookinfo
          subset:
            version: v1
      pathRewrite: /reviews
  - matchers:
    - uri:
        prefix: /v2
    name: v2
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews
          namespace: bookinfo
          subset:
            version: v2
      pathRewrite: /reviews
EOF
{{< /tab >}}
{{< /tabs >}}

Now, specify a `delegateAction` field in the `VirtualGateway` to send all `/reviews` requests to the reviews route table.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews
    name: reviews
    delegateAction:
      refs:
      - name: demo-routetable
        namespace: gloo-mesh
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews
    name: reviews
    delegateAction:
      refs:
      - name: demo-routetable
        namespace: gloo-mesh
EOF
{{< /tab >}}
{{< /tabs >}}

Route behavior is unchanged, but the subroutes are now defined in a separate Kubernetes resource.

Next, say you want to create a catch-all `404` route for reviews on a different `RouteTable`:

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml">}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: demo-routetable2
  namespace: gloo-mesh
spec:
  routes:
  - matchers:
    - uri:
        prefix: /
    name: not-found
    directResponseAction:
      status: 404
      body: "'custom not found'"
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: reviews-misc-rt
  namespace: gloo-mesh
  labels:
    service: reviews
spec:
  routes:
  - matchers:
    - uri:
        prefix: /
    name: not-found
    directResponseAction:
      status: 404
      body: "'custom not found'"
EOF
{{< /tab >}}
{{< /tabs >}}

Next, add the table to the `VirtualHost`.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews
    name: reviews
    delegateAction:
      refs:
      - name: demo-routetable
        namespace: gloo-mesh
      - name: demo-routetable2
        namespace: gloo-mesh
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews
    name: reviews
    delegateAction:
      refs:
      - name: demo-routetable
        namespace: gloo-mesh
      - name: demo-routetable2
        namespace: gloo-mesh
EOF
{{< /tab >}}
{{< /tabs >}}

It is important that the second route table comes second in the list of references, otherwise it will short circuit all
the other routes. You should now be able to request the reviews services endpoints to get the reviews and all other
random routes should return our custom `404` message.

### Improving the Sorting Logic

In the previous section, you added two `RouteTables` to handle routing logic for the reviews service and used a delegate
action to route to the route tables. This process included adding every `RouteTable` to the `VirtualHost` delegate 
action. However, depending on your requirements, manually adding individual route table references might not 
be viable. Instead, you can specify a selector to the delegate action.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews
    name: reviews
    delegateAction:
      selector:
        namespaces:
        - gloo-mesh
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews
    name: reviews
    delegateAction:
      selector:
        namespaces:
        - gloo-mesh
EOF
{{< /tab >}}
{{< /tabs >}}

Requests are now automatically delegated to all `RouteTable`s in the `gloo-mesh` namespace label without any changes to
the `VirtualHost`. You can use a combination of specific `RouteTable` references and label selectors. However,
`RouteTable` references take precedence over `RouteTable` selectors, which are sorted alphabetically by namespace and
then name because Kubernetes does not guarantee a deterministic order when selecting multiple objects. Because the `404`
route matches all requests to `/reviews`, the `404` route must come _after_ the service routes to prevent it from short-circuiting them.
In situations like this, you can add weights to the `RouteTable`s in order to guarantee a sort order.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml">}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: demo-routetable
  namespace: gloo-mesh
spec:
  routes:
  - matchers:
    - uri:
        prefix: /v1
    name: v1
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews
          namespace: bookinfo
          subset:
            version: v1
      pathRewrite: /reviews
  - matchers:
    - uri:
        prefix: /v2
    name: v2
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews
          namespace: bookinfo
          subset:
            version: v2
      pathRewrite: /reviews
  weight: 10
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: demo-routetable
  namespace: gloo-mesh
spec:
  routes:
  - matchers:
    - uri:
        prefix: /v1
    name: v1
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews
          namespace: bookinfo
          subset:
            version: v1
      pathRewrite: /reviews
  - matchers:
    - uri:
        prefix: /v2
    name: v2
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews
          namespace: bookinfo
          subset:
            version: v2
      pathRewrite: /reviews
  weight: 10
EOF
{{< /tab >}}
{{< /tabs >}}

The routes of that `RouteTable` will now always be first.

Now, say you want to add a version endpoint to each review service. You can add them to the new `RouteTable`.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml">}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: demo-routetable2
  namespace: gloo-mesh
spec:
  routes:
  - matchers:
    - uri:
        exact: /v1/version
    name: v1-version
    directResponseAction:
      status: 200
      body: "'v1'"
  - matchers:
    - uri:
        exact: /v2/version
    name: v2-version
    directResponseAction:
      status: 200
      body: "'v1'"
  - matchers:
    - uri:
        prefix: /
    name: not-found
    directResponseAction:
      status: 404
      body: "'custom not found'"
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: reviews-misc-rt
  namespace: gloo-mesh
  labels:
    service: reviews
spec:
  routes:
  - matchers:
    - uri:
        exact: /v1/version
    name: v1-version
    directResponseAction:
      status: 200
      body: "'v1'"
  - matchers:
    - uri:
        exact: /v2/version
    name: v2-version
    directResponseAction:
      status: 200
      body: "'v2'"
  - matchers:
    - uri:
        prefix: /
    name: not-found
    directResponseAction:
      status: 404
      body: "'custom not found'"
EOF
{{< /tab >}}
{{< /tabs >}}

Now if you try to request the new version endpoints, you will see that you get a `404` from the reviews service as a
reply. The reason is because the first route table that we created takes precedence due to its higher weight, which
means the prefix route at `/reviews/v1` is getting matched before it reaches the new exact route `/reviews/v1/version`.
You could adjust the weight on the second route table to be higher than that of the first, but that will cause the `404`
route to then short circuit the review service routes. One solution is using a separate route table for the version
endpoints and `404` endpoint, but there is a better way. Gloo Mesh Gateway routing supports two types of sorting. The
first and default method is via table weight where routes are kept in the same order that they appear on their table and
relative to table order. The alternative method is sorting by route specificity where routes are sorted via a heuristic
that estimates how "specific" a route is and then puts more specific routes before more general ones in order to
minimize short circuits. Exact match routes are considered more specific than regex match routes, which are considered
more specific than prefix match routes. Routes of the same match type are than compared based on the length where
longer matches are more specific. The sorting action is set in the `delegateAction` level and will overwrite any sorting
done in `delegateAction`s of child `RouteTable`s. You can enable it on the `VirtualHost` to allow the routes of both
tables to be intermingled.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews
    name: reviews
    delegateAction:
      selector:
        namespaces:
        - gloo-mesh
      sortMethod: ROUTE_SPECIFICITY
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: mgmt-cluster
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews
    name: reviews
    delegateAction:
      selector:
        namespaces:
        - gloo-mesh
      sortMethod: ROUTE_SPECIFICITY
EOF
{{< /tab >}}
{{< /tabs >}}

The resulting routes will be sorted such that the version endpoints can be hit as can the service endpoints, and all
other requests will be caught by the `404` route.

You can learn more about the routing options by looking at the Helm values for
[routes]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.enterprise.networking.v1beta1.route/" %}}).
