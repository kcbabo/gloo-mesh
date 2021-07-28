# Istio Gateway Deployment

The gateway architecture that we recommend requires you to create your own LoadBalanced service not managed by Istio. This way if you choose to, you can run multiple versions of the same gateway behind the same loadbalancers in a blue/green type architecture.

This image illustrates the multiple IstioOperator configurations that might exist for your Istio deployment if running multiple versions of Istio. 

![Production Istio Gateways](../../img/production-istio_gateways.png)

It is recommended that each Istio Gateway have its own IstioOperator configuration file. This allows the admin to upgrade each independently when changes are made.

## NOTE

* https://github.com/istio/istio/issues/33075
* Due to a significant bug in which the gateways rely on the same configmap istiod uses `istio-1-10-3` in `istio-system`. the operator has to copy this configmap to the `istio-gateways` namespace

```sh
REVISION=1-10-3
CM_DATA=$(kubectl get configmap istio-$REVISION -n istio-system -o jsonpath={.data})

cat <<EOF > ./istio-$REVISION.json
{
    "apiVersion": "v1",
    "data": $CM_DATA,
    "kind": "ConfigMap",
    "metadata": {
        "labels": {
            "istio.io/rev": "1-10-3"
        },
        "name": "istio-1-10-3",
        "namespace": "istio-gateways"
    }
}
EOF

kubectl apply -f istio-$REVISION.json
```


## Examples

```yaml
# single point of entry
apiVersion: v1
kind: Service
metadata:
  name: istio-ingressgateway
  namespace: istio-gateways
spec:
  type: LoadBalancer
  selector:
    istio: ingressgateway
    # select the 1-10-3 revision
    version: 1-10-3
    ports:
      - name: status-port
        port: 15021
        targetPort: 15021
      - name: http2
        port: 80
        targetPort: 8080
      - name: https
        port: 443
        targetPort: 8443
      - name: tcp
        port: 31400
        targetPort: 31400
      - name: tls
        port: 15443
        targetPort: 15443
---
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: ingress-gateway-1-10-3
  namespace: istio-gateways
spec:
  profile: empty
  # This value is required for Gloo Mesh Istio
  hub: gcr.io/istio-enterprise
  # This value can be any Gloo Mesh Istio tag
  tag: 1.10.3
  revision: 1-10-3
  components:
    ingressGateways:
      - name: istio-ingressgateway-1-10-3
        namespace: istio-gateways
        enabled: true
        label:
          istio: ingressgateway
          version: 1-10-3
          app: istio-ingressgateway
          # matches spec.values.global.network in istiod deployment
          topology.istio.io/network: production-cluster-network
        k8s:
          hpaSpec:
            maxReplicas: 5
            metrics:
              - resource:
                  name: cpu
                  targetAverageUtilization: 60
                type: Resource
            minReplicas: 2
            scaleTargetRef:
              apiVersion: apps/v1
              kind: Deployment
              name: istio-ingressgateway-1-10-3
          strategy:
            rollingUpdate:
              maxSurge: 100%
              maxUnavailable: 25%
          resources:
            limits:
              cpu: 2000m
              memory: 1024Mi
            requests:
              cpu: 100m
              memory: 40Mi
          service:
            # Since we created our own LoadBalanced service, tell istio to create a ClusterIP service for this gateway
            type: ClusterIP
            # match the LoadBalanced Service
            ports:
              - name: status-port
                port: 15021
                targetPort: 15021
              - name: http2
                port: 80
                targetPort: 8080
              - name: https
                port: 443
                targetPort: 8443
              - name: tcp
                port: 31400
                targetPort: 31400
              - name: tls
                port: 15443
                targetPort: 15443
  values:
    global:
      # needed for connecting VirtualMachines to the mesh
      network: production-cluster-network
      # needed for annotating istio metrics with cluster
      multiCluster:
        clusterName: production-cluster
---
apiVersion: v1
kind: Service
metadata:
  name: istio-eastwestgateway
  namespace: istio-gateways
spec:
  type: LoadBalancer
  selector:
    istio: eastwestgateway
    # select the 1-10-3 revision
    version: "1-10-3"
    ports:
      - name: status-port
        port: 15021
        targetPort: 15021
      - name: http2
        port: 80
        targetPort: 8080
      - name: https
        port: 443
        targetPort: 8443
      - name: tcp
        port: 31400
        targetPort: 31400
      - name: tls
        port: 15443
        targetPort: 15443
---
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: eastwest-gateway-1-10-3
  namespace: istio-gateways
spec:
  profile: empty
  revision: 1-10-3
  # This value is required for Gloo Mesh Istio
  hub: gcr.io/istio-enterprise
  # This value can be any Gloo Mesh Istio tag
  tag: 1.10.3
  components:
    ingressGateways:
      - name: istio-eastwestgateway
        namespace: istio-gateways
        enabled: true
        label:
          istio: eastwestgateway
          version: 1-10-3
          app: istio-eastwestgateway
          # matches spec.values.global.network in istiod deployment
          topology.istio.io/network: production-cluster-network
        k8s:
          env:
            # traffic through this gateway should be routed inside the network
            # matches spec.values.global.network in istiod deployment
            - name: ISTIO_META_REQUESTED_NETWORK_VIEW
              value: production-cluster-network
          hpaSpec:
            maxReplicas: 5
            metrics:
              - resource:
                  name: cpu
                  targetAverageUtilization: 60
                type: Resource
            minReplicas: 2
            scaleTargetRef:
              apiVersion: apps/v1
              kind: Deployment
              name: istio-eastwest-1-10-3
          strategy:
            rollingUpdate:
              maxSurge: 100%
              maxUnavailable: 25%
          resources:
            limits:
              cpu: 2000m
              memory: 1024Mi
            requests:
              cpu: 100m
              memory: 40Mi
          service:
            ports:
              - name: status-port
                port: 15021
                targetPort: 15021
              - name: tls
                port: 15443
                targetPort: 15443
              - name: tls-istiod
                port: 15012
                targetPort: 15012
              - name: tls-webhook
                port: 15017
                targetPort: 15017
  values:
    global:
      # needed for connecting VirtualMachines to the mesh
      network: production-cluster-network
      # needed for annotating istio metrics with cluster
      multiCluster:
        clusterName: production-cluster
```

### Navigate

* [Deploying Istio Control Plane](./istiod_deployment.md)
* [Deploying Istio Operator](./operator_deployment.md)