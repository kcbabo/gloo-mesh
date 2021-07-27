# Full IstioOperator Spec Examples

Below is a full IstioOperator spec examples that shows you how to set a number of common values.

## Deployment with Revisions

Following [Canary Based Deployment](https://istio.io/latest/blog/2017/0.1-canary/) from the Istio website, we will deploy Istio with a revision label that matches its version. This makes it easy to migrate to new versions of Istio control plane when they are available.

Below is a couple of example deployments with their respective production settings. Depending on your environment you may need to edit certain istio functionality.

## References

* [Istio default profiles](https://github.com/istio/istio/tree/master/manifests/profiles)
* [Istio Operator Spec](https://istio.io/latest/docs/reference/config/istio.operator.v1alpha1/)
* [Istio MeshConfig Spec](https://istio.io/latest/docs/reference/config/istio.mesh.v1alpha1/#MeshConfig)

## Example Production IstioOperator

```yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: production-example
  namespace: istio-system
spec:
  profile: minimal
  # This value is required for Gloo Mesh Istio
  hub: gcr.io/istio-enterprise
  # This value can be any Gloo Mesh Istio tag
  tag: 1.10.3
  revision: 1-10-3

  # You may override parts of meshconfig by uncommenting the following lines.
  meshConfig:
    # enable access logging. Empty value disables access logging.
    # accessLogFile: /dev/stdout
    # Encoding for the proxy access log (TEXT or JSON). Default value is TEXT.
    accessLogEncoding: JSON

    enableTracing: false

    defaultConfig:
      # location of istiod service
      # discoveryAddress: istiod-1-10-3.istio-system.svc:15012
      # enable GlooMesh metrics service
      envoyMetricsService:
        address: enterprise-agent.gloo-mesh:9977
       # enable GlooMesh accesslog service
      envoyAccessLogService:
        address: enterprise-agent.gloo-mesh:9977
      # tracing:
      #   # sample 1% of traffic
      #   sampling: 01.0
      proxyMetadata:
        # Enable Istio agent to handle DNS requests for known hosts
        # Unknown hosts will automatically be resolved using upstream dns servers in resolv.conf
        ISTIO_META_DNS_CAPTURE: "true"
        # Enable automatic address allocation, optional
        ISTIO_META_DNS_AUTO_ALLOCATE: "true"
        # Used for gloo mesh metrics aggregation
        # should match trustDomain
        GLOO_MESH_CLUSTER_NAME: production-cluster
    
    # Specify if http1.1 connections should be upgraded to http2 by default. 
    # Can be overridden using DestinationRule
    # h2UpgradePolicy: UPGRADE

    # Set the default behavior of the sidecar for handling outbound traffic from the application.
    outboundTrafficPolicy:
      mode: ALLOW_ANY
    # The trust domain corresponds to the trust root of a system. For Gloo Mesh this should be the name of the cluster that cooresponds with the CA certificate CommonName identity
    trustDomain: production-cluster.solo.io
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

    # Helm values overrides
  values:
    # https://istio.io/v1.5/docs/reference/config/installation-options/#global-options
    global:
      # needed for connecting VirtualMachines to the mesh
      network: production-cluster-network
      # needed for annotating istio metrics with cluster
      multiCluster:
        clusterName: production-cluster
    #   istioNamespace: istio-system
      # proxy:
      #   # The Istio load tests mesh consists of 1000 services and 2000 sidecars with 70,000 mesh-wide requests per second and istio-proxy used 0.35 vCPU and 40 MB memory per 1000 requests per second (1.10.3).
      #   resources:
      #     requests:
      #       cpu: 100m
      #       memory: 128Mi
      #     limits: 
      #       cpu: 2000m
      #       memory: 1024Mi
      #   logLevel: warning
```

## Example Production OpenShift IstioOperator

Openshift requires that the Istio CNI be enabled. Based on the [Istio Openshift profile](https://github.com/istio/istio/blob/master/manifests/profiles/openshift.yaml).

```yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: production-example
  namespace: istio-system
spec:
  profile: minimal
  # This value is required for Gloo Mesh Istio
  hub: gcr.io/istio-enterprise
  # This value can be any Gloo Mesh Istio tag
  tag: 1.10.3
  revision: 1-10-3

  # You may override parts of meshconfig by uncommenting the following lines.
  meshConfig:
    # enable access logging. Empty value disables access logging.
    # accessLogFile: /dev/stdout
    # Encoding for the proxy access log (TEXT or JSON). Default value is TEXT.
    accessLogEncoding: JSON

    enableTracing: false

    defaultConfig:
      # location of istiod service
      # discoveryAddress: istiod-1-10-3.istio-system.svc:15012
      # enable GlooMesh metrics service
      envoyMetricsService:
        address: enterprise-agent.gloo-mesh:9977
       # enable GlooMesh accesslog service
      envoyAccessLogService:
        address: enterprise-agent.gloo-mesh:9977
      # tracing:
      #   # sample 1% of traffic
      #   sampling: 01.0
      proxyMetadata:
        # Enable Istio agent to handle DNS requests for known hosts
        # Unknown hosts will automatically be resolved using upstream dns servers in resolv.conf
        ISTIO_META_DNS_CAPTURE: "true"
        # Enable automatic address allocation, optional
        ISTIO_META_DNS_AUTO_ALLOCATE: "true"
        # Used for gloo mesh metrics aggregation
        # should match trustDomain
        GLOO_MESH_CLUSTER_NAME: production-cluster
    
    # Specify if http1.1 connections should be upgraded to http2 by default. 
    # Can be overridden using DestinationRule
    # h2UpgradePolicy: UPGRADE

    # Set the default behavior of the sidecar for handling outbound traffic from the application.
    outboundTrafficPolicy:
      mode: ALLOW_ANY
    # The trust domain corresponds to the trust root of a system. For Gloo Mesh this should be the name of the cluster that cooresponds with the CA certificate CommonName identity
    trustDomain: production-cluster.solo.io
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
      enabled: true
      namespace: kube-system
      k8s:
        overlays:
          - kind: DaemonSet
            name: istio-cni-node
            patches:
              - path: spec.template.spec.containers[0].securityContext.privileged
                value: true

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

  # CNI options if using OpenShift
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
    sidecarInjectorWebhook:
      injectedAnnotations:
        k8s.v1.cni.cncf.io/networks: istio-cni

    # https://istio.io/v1.5/docs/reference/config/installation-options/#global-options
    global:
      # needed for connecting VirtualMachines to the mesh
      network: production-cluster-network
      # needed for annotating istio metrics with cluster
      multiCluster:
        clusterName: production-cluster
    #   istioNamespace: istio-system
      # proxy:
      #   # The Istio load tests mesh consists of 1000 services and 2000 sidecars with 70,000 mesh-wide requests per second and istio-proxy used 0.35 vCPU and 40 MB memory per 1000 requests per second (1.10.3).
      #   resources:
      #     requests:
      #       cpu: 100m
      #       memory: 128Mi
      #     limits: 
      #       cpu: 2000m
      #       memory: 1024Mi
      #   logLevel: warning
```