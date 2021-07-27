# Full IstioOperator Spec Examples

Below is a full IstioOperator spec examples that shows you how to set a number of common values.

## References

* [Istio default profiles](https://github.com/istio/istio/tree/master/manifests/profiles)
* [Istio Operator Spec](https://istio.io/latest/docs/reference/config/istio.operator.v1alpha1/)
* [Istio MeshConfig Spec](https://istio.io/latest/docs/reference/config/istio.mesh.v1alpha1/#MeshConfig)


## Example IstioOperator

```yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-1
  namespace: istio-system
spec:
  profile: minimal
  hub: docker.io/istio
  tag: 1.10.3
  revision: 1-10-3

  # You may override parts of meshconfig by uncommenting the following lines.
  meshConfig:
    # enable access logging. Empty value disables access logging.
    accessLogFile: /dev/stdout
    # Encoding for the proxy access log (TEXT or JSON). Default value is TEXT.
    accessLogEncoding: TEXT
    enableTracing: false

    defaultConfig:
      tracing:
        # sample 10% of traffic
        sampling: 10.0
        # the maximum length for the request path included as part of the HttpUrl span tag is 256
        max_path_tag_length: 256
        custom_tags:
          # Literal represents a static value that gets added to each span
          my-company:
            literal:
              value: solo.io
          # In order to add custom tags based on environmental variables, you must modify the istio-sidecar-injector ConfigMap in your root Istio system namespace.
          clusterID:
            environment:
              name: ISTIO_META_CLUSTER_ID
              defaultValue: unknown
          # Client request header option can be used to populate tag value from an incoming client request header.
          user_id:
            header:
              name: x-user-id
              defaultValue: unknown
      proxyMetadata:
        # Enable Istio agent to handle DNS requests for known hosts
        # Unknown hosts will automatically be resolved using upstream dns servers in resolv.conf
        ISTIO_META_DNS_CAPTURE: "true"
        # Enable automatic address allocation, optional
        ISTIO_META_DNS_AUTO_ALLOCATE: "true"
    # Specify if http1.1 connections should be upgraded to http2 by default. 
    # Can be overridden using DestinationRule
    h2UpgradePolicy: UPGRADE
    # Set the default behavior of the sidecar for handling outbound traffic from the application.
    outboundTrafficPolicy: 
      mode: ALLOW_ANY
    # The trust domain corresponds to the trust root of a system. For Gloo Mesh this should be the name of the cluster that cooresponds with the CA certificate CommonName identity
    trustDomain: cluster-1.solo
    # The namespace to treat as the administrative root namespace for Istio configuration.
    rootNamespace: istio-config

  # Traffic management feature
  components:
    base:
      enabled: true
    pilot:
      enabled: true
      k8s:
        # Recommended to be >1 in production
        replicaCount: 2
        # The Istio load tests mesh consists of 1000 services and 2000 sidecars with 70,000 mesh-wide requests per second and Istiod used 1 vCPU and 1.5 GB of memory (1.10.3).
        resources:
          requests:
            cpu: 200m
            memory: 200Mi
        strategy:
          rollingUpdate:
            maxSurge: 100%
            maxUnavailable: 25%
        # recommended to scale istiod under load
        hpaSpec:
          maxReplicas: 5
          minReplicas: 2
          scaleTargetRef:
            apiVersion: apps/v1
            kind: Deployment
            # matches the format istiod-<revision_label>
            name: istiod-1-10-3
          metrics:
            - resource:
                name: cpu
                targetAverageUtilization: 60
              type: Resource
    # Istio CNI feature
    cni:
      enabled: false
      namespace: kube-system

    # Istio Gateway feature
    # Disable gateways deployments because they will be in separate IstioOperator configs
    ingressGateways:
    - name: istio-ingressgateway
      enabled: false
    - name: istio-eastwestgateway
      enabled: false
    egressGateways:
    - name: istio-egressgateway
      enabled: false

    # istiod remote configuration when istiod isn't installed on the cluster
    istiodRemote:
      enabled: false

  # CNI options if using OpenShift
  # ONLY used if spec.components.cni.enabled == true
  values:
    cni:
      cniBinDir: /var/lib/cni/bin
      cniConfDir: /etc/cni/multus/net.d
      chained: false
      cniConfFileName: "istio-cni.conf"
      excludeNamespaces:
       - istio-system
       - kube-system
      logLevel: info
```