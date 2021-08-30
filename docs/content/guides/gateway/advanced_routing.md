---
title: Advanced Routing
weight: 50
description: How to configure and use multiple route tables
---

This guide will walk you through setting up your `VirtualGateway` with multiple external `RouteTable`s in order to get
more fine-grained control over how routes are organized.

## Before You Begin

Before you begin, ensure that your setup for the management cluster (`cluster-1`) and a remote cluster (`cluster-2`) meets all of the following prerequisites:
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
  * Istio:
    * Istio is [installed in the `istio-system` namespace on both clusters]({{% versioned_link_path fromRoot="/guides/installing_istio" %}})
    * `istio-ingressgateway` is deployed in the `istio-system` namespace of `cluster-1`, which is the default installation namespace
    * The `bookinfo` app is [installed in the `bookinfo` namespace of both clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}})

To understand the custom resources that are used in this guide, see the [Gateway Concepts Overview]({{% versioned_link_path fromRoot="/guides/gateway/concepts" %}}).

## Setting up a Basic VirtualGateway

As a starting point, create the following `VirtualGateway` to route to the reviews and ratings service. We will be using
an inline `VirtualHost`.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml">}}
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
      - virtualHost:
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
                  clusterName: cluster-1
                  name: ratings
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews
            name: ratings
            routeAction:
              destinations:
              - kubeService:
                  clusterName: cluster-1
                  name: reviews
                  namespace: bookinfo
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
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
      - virtualHost:
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
                  clusterName: cluster-1
                  name: ratings
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews
            name: ratings
            routeAction:
              destinations:
              - kubeService:
                  clusterName: cluster-1
                  name: reviews
                  namespace: bookinfo
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

Alternatively you can forward the port to access locally, which is sufficient for the purposes of this guide:

```shell
kubectl --context $CONTEXT_1 -n istio-system port-forward deploy/istio-ingressgateway 8081
```

### Adding Routes

Next, you can add routes to split the reviews service into separate `v1` and `v2` services and route to
each version. Start by deleting the existing reviews service.

```shell
kubectl --context $CONTEXT_1 -n bookinfo delete svc reviews
```

And adding two services for the v1 and v2 deployments:

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml">}}
apiVersion: v1
kind: Service
metadata:
  annotations:
  labels:
    app: reviews
    service: reviews
    version: v1
  name: reviews
  namespace: bookinfo
spec:
  ports:
  - name: http
    port: 9080
    protocol: TCP
    targetPort: 9080
  selector:
    app: reviews
    version: v1
  sessionAffinity: None
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  annotations:
  labels:
    app: reviews
    service: reviews
    version: v2
  name: reviews-v2
  namespace: bookinfo
spec:
  ports:
  - name: http
    port: 9080
    protocol: TCP
    targetPort: 9080
  selector:
    app: reviews
    version: v2
  sessionAffinity: None
  type: ClusterIP
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: v1
kind: Service
metadata:
  annotations:
  labels:
    app: reviews
    service: reviews
    version: v1
  name: reviews
  namespace: bookinfo
spec:
  ports:
  - name: http
    port: 9080
    protocol: TCP
    targetPort: 9080
  selector:
    app: reviews
    version: v1
  sessionAffinity: None
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  annotations:
  labels:
    app: reviews-v2
    service: reviews
    version: v2
  name: reviews
  namespace: bookinfo
spec:
  ports:
  - name: http
    port: 9080
    protocol: TCP
    targetPort: 9080
  selector:
    app: reviews
    version: v2
  sessionAffinity: None
  type: ClusterIP
EOF
{{< /tab >}}
{{< /tabs >}}

Update the `VirtualHost` to route `/v1` and `/v2` to the versioned services.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml">}}
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
      - virtualHost:
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
                  clusterName: cluster-1
                  name: ratings
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews/v1
            name: reviews-v1
            routeAction:
              destinations:
              - kubeService:
                  clusterName: cluster-1
                  name: reviews-v1
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews/v2
            name: reviews-v2
            routeAction:
              destinations:
              - kubeService:
                  clusterName: cluster-1
                  name: reviews-v2
                  namespace: bookinfo
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
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
      - virtualHost:
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
                  clusterName: cluster-1
                  name: ratings
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews/v1
            name: reviews-v1
            routeAction:
              destinations:
              - kubeService:
                  clusterName: cluster-1
                  name: reviews-v1
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews/v2
            name: reviews-v2
            routeAction:
              destinations:
              - kubeService:
                  clusterName: cluster-1
                  name: reviews-v2
                  namespace: bookinfo
EOF
{{< /tab >}}
{{< /tabs >}}

## Splitting Out the Routes

Now that multiple routes are set up for the `/reviews` prefix, you can create one `RouteTable` resource to split the routes into respective route tables for better 
organization. Route tables also provide more fine-grained access control because authorized users can 
have permission to edit specific route tables without having permission to edit the `VirtualHost` or other route tables. Start by creating a
`RouteTable` resource with the `v1` and `v2` routes. Note that the `/reviews` prefix is omitted.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml">}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: reviews-rt
  namespace: gloo-mesh
  labels:
    service: reviews
spec:
  routes:
  - matchers:
    - uri:
        prefix: /v1
    name: reviews-v1
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews-v1
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /v2
    name: reviews-v2
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews-v2
          namespace: bookinfo
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: reviews-rt
  namespace: gloo-mesh
  labels:
    service: reviews
spec:
  routes:
  - matchers:
    - uri:
        prefix: /v1
    name: reviews-v1
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews-v1
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /v2
    name: reviews-v2
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews-v2
          namespace: bookinfo
EOF
{{< /tab >}}
{{< /tabs >}}

Now, specify a `delegateAction` field in the `VirtualGateway` to send all `/reviews` requests to the reviews route table.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml" >}}
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
      - virtualHost:
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
                  clusterName: cluster-1
                  name: ratings
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews
            name: reviews
            delegateAction:
              refs:
              - name: reviews-rt
                namespace: gloo-mesh
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
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
      - virtualHost:
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
                  clusterName: cluster-1
                  name: ratings
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews
            name: reviews
            routeAction:
              destinations:
              - kubeService:
                  clusterName: cluster-1
                  name: reviews-v1
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews
            name: reviews
            delegateAction:
              refs:
              - name: reviews-rt
                namespace: gloo-mesh
EOF
{{< /tab >}}
{{< /tabs >}}

Route behavior is unchanged, but the subroutes are now defined in a separate Kubernetes resource.

Next, say you want to create a few routes that are directly on the `/reviews` path. First, create another route table.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml">}}
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
        exact: /versions
    name: versions
    directResponseAction:
      status: 200
      body: "['v1', 'v2']"
  - matchers:
    - uri:
        prefix: /
    name: not-found
    directResponseAction:
      status: 404
      body: "'not found'"
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
        exact: /versions
    name: versions
    directResponseAction:
      status: 200
      body: "['v1', 'v2']"
  - matchers:
    - uri:
        prefix: /
    name: not-found
    directResponseAction:
      status: 404
      body: "'not found'"
EOF
{{< /tab >}}
{{< /tabs >}}

Next, add the table to the `VirtualGateway`.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml" >}}
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
      - virtualHost:
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
                  clusterName: cluster-1
                  name: ratings
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews
            name: reviews
            delegateAction:
              refs:
              - name: reviews-rt
                namespace: gloo-mesh
              - name: reviews-misc-rt
                namespace: gloo-mesh
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
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
      - virtualHost:
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
                  clusterName: cluster-1
                  name: ratings
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews
            name: reviews
            routeAction:
              destinations:
              - kubeService:
                  clusterName: cluster-1
                  name: reviews-v1
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews
            name: reviews
            delegateAction:
              refs:
              - name: reviews-rt
                namespace: gloo-mesh
              - name: reviews-misc-rt
                namespace: gloo-mesh
EOF
{{< /tab >}}
{{< /tabs >}}

### Improving the Sorting Logic

In the previous section, you added two `RouteTables` to handle routing logic for the reviews service and used a delegate
action to route to the the route tables. This process included adding every `RouteTable` to the `VirtualGateway` delegate 
action. However, depending on your requirements, manually adding individual route table references might not 
be viable. Instead, you can specify a selector to the delegate action.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml" >}}
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
      - virtualHost:
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
                  clusterName: cluster-1
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
                labels
                  service: reviews
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
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
      - virtualHost:
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
                  clusterName: cluster-1
                  name: ratings
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews
            name: reviews
            routeAction:
              destinations:
              - kubeService:
                  clusterName: cluster-1
                  name: reviews-v1
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews
            name: reviews
            delegateAction:
              selector:
                namespaces:
                - gloo-mesh
                labels
                  service: reviews
EOF
{{< /tab >}}
{{< /tabs >}}

Now all `RouteTables` in the gloo-mesh namespace with the `service: reviews` will be automatically delegated to without
any changes to the `VirtualGateway`. Both manually listing `RouteTable`s and selecting them can be desirable based on
the requirements of your system. You can even use a combination, but note that `RouteTable`s listed by reference take
higher precedence than selected ones, which are sorted alphabetically by namespace then name since Kubernetes does not
guarantee a deterministic order when selecting multiple objects.

In the `reviews-misc-rt` route table resource, `not-found` is defined as a catch-all 404 route. This route must always be 
last in the list of routes that requests are matched against to ensure that it does not bypass the other routes for services. 
To designate the order of routes that requests are matched against, you can add weights to each `RouteTable`.
For example, add `weight: 10` to the `reviews-rt` route table so that all its routes are listed before the routes in the `reviews-misc-rt` 
route table, which has a default weight of 0.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml">}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: reviews-rt
  namespace: gloo-mesh
  labels:
    service: reviews
spec:
  weight: 10
  routes:
  - matchers:
    - uri:
        prefix: /v1
    name: reviews-v1
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews-v1
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /v2
    name: reviews-v2
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews-v2
          namespace: bookinfo
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: reviews-rt
  namespace: gloo-mesh
  labels:
    service: reviews
spec:
  weight: 10
  routes:
  - matchers:
    - uri:
        prefix: /v1
    name: reviews-v1
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews-v1
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /v2
    name: reviews-v2
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews-v2
          namespace: bookinfo
EOF
{{< /tab >}}
{{< /tabs >}}

Routes for services in `reviews-rt` are now first in the route list. Alternatively, when routes are segmented into
separate route tables, you can change the delegate action's sort method to `ROUTE_SPECIFICITY`. This sort method uses a
heuristic to estimate the specificity of each route, and orders routes from most to least specific. To enable 
this sort method, add `sortMethod: ROUTE_SPECIFICITY` to the delegate action.

{{< tabs >}}
{{< tab name="YAML File" codelang="yaml" >}}
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
      - virtualHost:
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
                  clusterName: cluster-1
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
                labels
                  service: reviews
              sortMethod: ROUTE_SPECIFICITY
{{< /tab >}}
{{< tab name="CLI Inline" codelang="shell" >}}
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
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
      - virtualHost:
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
                  clusterName: cluster-1
                  name: ratings
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews
            name: reviews
            routeAction:
              destinations:
              - kubeService:
                  clusterName: cluster-1
                  name: reviews-v1
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews
            name: reviews
            delegateAction:
              selector:
                namespaces:
                - gloo-mesh
                labels
                  service: reviews
              sortMethod: ROUTE_SPECIFICITY
EOF
{{< /tab >}}
{{< /tabs >}}
