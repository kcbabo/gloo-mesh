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
        resources:
          requests:
            cpu: 10m
            memory: 100Mi
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






























```yaml
#From Tanzu

apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: installed-state
  namespace: istio-system
spec:
  components:
    base:
      enabled: true
    cni:
      enabled: true
      namespace: kube-system
    egressGateways:
      - enabled: true
        k8s:
          env:
            - name: ISTIO_META_ROUTER_MODE
              value: sni-dnat
          hpaSpec:
            maxReplicas: 5
            metrics:
              - resource:
                  name: cpu
                  targetAverageUtilization: 80
                type: Resource
            minReplicas: 1
            scaleTargetRef:
              apiVersion: apps/v1
              kind: Deployment
              name: istio-egressgateway
          resources:
            limits:
              cpu: 2000m
              memory: 1024Mi
            requests:
              cpu: 10m
              memory: 40Mi
          service:
            ports:
              - name: http2
                port: 80
                targetPort: 8080
              - name: https
                port: 443
                targetPort: 8443
              - name: tls
                port: 15443
                targetPort: 15443
          strategy:
            rollingUpdate:
              maxSurge: 100%
              maxUnavailable: 25%
        name: istio-egressgateway
    ingressGateways:
      - enabled: true
        k8s:
          env:
            - name: ISTIO_META_ROUTER_MODE
              value: sni-dnat
          hpaSpec:
            maxReplicas: 5
            metrics:
              - resource:
                  name: cpu
                  targetAverageUtilization: 80
                type: Resource
            minReplicas: 1
            scaleTargetRef:
              apiVersion: apps/v1
              kind: Deployment
              name: istio-ingressgateway
          resources:
            limits:
              cpu: 2000m
              memory: 1024Mi
            requests:
              cpu: 10m
              memory: 40Mi
          service:
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
          strategy:
            rollingUpdate:
              maxSurge: 100%
              maxUnavailable: 25%
        name: istio-ingressgateway
    istiodRemote:
      enabled: false
    pilot:
      enabled: true
      k8s:
        env:
          - name: PILOT_TRACE_SAMPLING
            value: "100"
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 1
          periodSeconds: 3
          timeoutSeconds: 5
        resources:
          requests:
            cpu: 10m
            memory: 100Mi
        strategy:
          rollingUpdate:
            maxSurge: 100%
            maxUnavailable: 25%
    policy:
      enabled: false
      k8s:
        env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
        hpaSpec:
          maxReplicas: 5
          metrics:
            - resource:
                name: cpu
                targetAverageUtilization: 80
              type: Resource
          minReplicas: 1
          scaleTargetRef:
            apiVersion: apps/v1
            kind: Deployment
            name: istio-policy
        strategy:
          rollingUpdate:
            maxSurge: 100%
            maxUnavailable: 25%
    telemetry:
      enabled: false
      k8s:
        env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: GOMAXPROCS
            value: "6"
        hpaSpec:
          maxReplicas: 5
          metrics:
            - resource:
                name: cpu
                targetAverageUtilization: 80
              type: Resource
          minReplicas: 1
          scaleTargetRef:
            apiVersion: apps/v1
            kind: Deployment
            name: istio-telemetry
        replicaCount: 1
        resources:
          limits:
            cpu: 4800m
            memory: 4G
          requests:
            cpu: 1000m
            memory: 1G
        strategy:
          rollingUpdate:
            maxSurge: 100%
            maxUnavailable: 25%
  hub: docker.io/istio
  meshConfig:
    accessLogFile: /dev/stdout
    defaultConfig:
      proxyMetadata: {}
    enablePrometheusMerge: true
    outboundTrafficPolicy:
      mode: REGISTRY_ONLY
  profile: demo
  revision: 1-7-3
  values:
    base:
      enableCRDTemplates: false
      validationURL: ""
    clusterResources: true
    gateways:
      istio-egressgateway:
        autoscaleEnabled: false
        env: {}
        name: istio-egressgateway
        secretVolumes:
          - mountPath: /etc/istio/egressgateway-certs
            name: egressgateway-certs
            secretName: istio-egressgateway-certs
          - mountPath: /etc/istio/egressgateway-ca-certs
            name: egressgateway-ca-certs
            secretName: istio-egressgateway-ca-certs
        type: ClusterIP
        zvpn: {}
      istio-ingressgateway:
        applicationPorts: ""
        autoscaleEnabled: false
        debug: info
        domain: ""
        env: {}
        meshExpansionPorts:
          - name: tcp-istiod
            port: 15012
            targetPort: 15012
          - name: tcp-dns-tls
            port: 853
            targetPort: 8853
        name: istio-ingressgateway
        secretVolumes:
          - mountPath: /etc/istio/ingressgateway-certs
            name: ingressgateway-certs
            secretName: istio-ingressgateway-certs
          - mountPath: /etc/istio/ingressgateway-ca-certs
            name: ingressgateway-ca-certs
            secretName: istio-ingressgateway-ca-certs
        type: LoadBalancer
        zvpn: {}
    global:
      arch:
        amd64: 2
        ppc64le: 2
        s390x: 2
      configValidation: true
      controlPlaneSecurityEnabled: true
      defaultNodeSelector: {}
      defaultPodDisruptionBudget:
        enabled: true
      defaultResources:
        requests:
          cpu: 10m
      enableHelmTest: false
      hub: harbor-k8s.shk8s.de/istio
      imagePullPolicy: ""
      imagePullSecrets: []
      istioNamespace: istio-system
      istiod:
        enableAnalysis: false
      jwtPolicy: third-party-jwt
      logAsJson: false
      logging:
        level: default:info
      meshExpansion:
        enabled: false
        useILB: false
      meshNetworks: {}
      mountMtlsCerts: false
      multiCluster:
        clusterName: ""
        enabled: false
      network: ""
      omitSidecarInjectorConfigMap: false
      oneNamespace: false
      operatorManageWebhooks: false
      pilotCertProvider: istiod
      priorityClassName: ""
      proxy:
        autoInject: enabled
        clusterDomain: cluster.local
        componentLogLevel: misc:error
        enableCoreDump: false
        excludeIPRanges: ""
        excludeInboundPorts: ""
        excludeOutboundPorts: ""
        image: proxyv2
        includeIPRanges: "*"
        logLevel: warning
        privileged: false
        readinessFailureThreshold: 30
        readinessInitialDelaySeconds: 1
        readinessPeriodSeconds: 2
        resources:
          limits:
            cpu: 2000m
            memory: 1024Mi
          requests:
            cpu: 10m
            memory: 40Mi
        statusPort: 15020
        tracer: zipkin
      proxy_init:
        image: proxyv2
        resources:
          limits:
            cpu: 2000m
            memory: 1024Mi
          requests:
            cpu: 10m
            memory: 10Mi
      sds:
        token:
          aud: istio-ca
      sts:
        servicePort: 0
      tracer:
        datadog:
          address: $(HOST_IP):8126
        lightstep:
          accessToken: ""
          address: ""
        stackdriver:
          debug: false
          maxNumberOfAnnotations: 200
          maxNumberOfAttributes: 200
          maxNumberOfMessageEvents: 200
        zipkin:
          address: ""
      trustDomain: cluster.local
      useMCP: false
    grafana:
      accessMode: ReadWriteMany
      contextPath: /grafana
      dashboardProviders:
        dashboardproviders.yaml:
          apiVersion: 1
          providers:
            - disableDeletion: false
              folder: istio
              name: istio
              options:
                path: /var/lib/grafana/dashboards/istio
              orgId: 1
              type: file
      datasources:
        datasources.yaml:
          apiVersion: 1
      env: {}
      envSecrets: {}
      image:
        repository: grafana/grafana
        tag: 7.0.5
      nodeSelector: {}
      persist: false
      podAntiAffinityLabelSelector: []
      podAntiAffinityTermLabelSelector: []
      security:
        enabled: false
        passphraseKey: passphrase
        secretName: grafana
        usernameKey: username
      service:
        annotations: {}
        externalPort: 3000
        name: http
        type: ClusterIP
      storageClassName: ""
      tolerations: []
    istiocoredns:
      coreDNSImage: coredns/coredns
      coreDNSPluginImage: istio/coredns-plugin:0.2-istio-1.1
      coreDNSTag: 1.6.2
    istiodRemote:
      injectionURL: ""
    mixer:
      adapters:
        kubernetesenv:
          enabled: true
        prometheus:
          enabled: true
          metricsExpiryDuration: 10m
        stackdriver:
          auth:
            apiKey: ""
            appCredentials: false
            serviceAccountPath: ""
          enabled: false
          tracer:
            enabled: false
            sampleProbability: 1
        stdio:
          enabled: false
          outputAsJson: false
        useAdapterCRDs: false
      policy:
        adapters:
          kubernetesenv:
            enabled: true
          useAdapterCRDs: false
        autoscaleEnabled: true
        image: mixer
        sessionAffinityEnabled: false
      telemetry:
        autoscaleEnabled: true
        env:
          GOMAXPROCS: "6"
        image: mixer
        loadshedding:
          latencyThreshold: 100ms
          mode: enforce
        nodeSelector: {}
        podAntiAffinityLabelSelector: []
        podAntiAffinityTermLabelSelector: []
        replicaCount: 1
        sessionAffinityEnabled: false
        tolerations: []
    pilot:
      appNamespaces: []
      autoscaleEnabled: false
      autoscaleMax: 5
      autoscaleMin: 1
      configMap: true
      configNamespace: istio-config
      cpu:
        targetAverageUtilization: 80
      enableProtocolSniffingForInbound: true
      enableProtocolSniffingForOutbound: true
      env: {}
      image: pilot
      keepaliveMaxServerConnectionAge: 30m
      nodeSelector: {}
      podAntiAffinityLabelSelector: []
      podAntiAffinityTermLabelSelector: []
      policy:
        enabled: false
      replicaCount: 1
      tolerations: []
      traceSampling: 1
    prometheus:
      contextPath: /prometheus
      hub: docker.io/prom
      nodeSelector: {}
      podAntiAffinityLabelSelector: []
      podAntiAffinityTermLabelSelector: []
      provisionPrometheusCert: true
      retention: 6h
      scrapeInterval: 15s
      security:
        enabled: true
      tag: v2.19.2
      tolerations: []
    sidecarInjectorWebhook:
      enableNamespacesByDefault: false
      injectLabel: istio-injection
      objectSelector:
        autoInject: true
        enabled: false
      rewriteAppHTTPProbe: true
    telemetry:
      enabled: true
      v1:
        enabled: false
      v2:
        enabled: true
        metadataExchange:
          wasmEnabled: false
        prometheus:
          enabled: true
          wasmEnabled: false
        stackdriver:
          configOverride: {}
          enabled: false
          logging: false
          monitoring: false
          topology: false
    tracing:
      jaeger:
        accessMode: ReadWriteMany
        hub: docker.io/jaegertracing
        memory:
          max_traces: 50000
        persist: false
        spanStorageType: badger
        storageClassName: ""
        tag: "1.18"
      nodeSelector: {}
      opencensus:
        exporters:
          stackdriver:
            enable_tracing: true
        hub: docker.io/omnition
        resources:
          limits:
            cpu: "1"
            memory: 2Gi
          requests:
            cpu: 200m
            memory: 400Mi
        tag: 0.1.9
      podAntiAffinityLabelSelector: []
      podAntiAffinityTermLabelSelector: []
      provider: jaeger
      service:
        annotations: {}
        externalPort: 9411
        name: http-query
        type: ClusterIP
      zipkin:
        hub: docker.io/openzipkin
        javaOptsHeap: 700
        maxSpans: 500000
        node:
          cpus: 2
        probeStartupDelay: 10
        queryPort: 9411
        resources:
          limits:
            cpu: 1000m
            memory: 2048Mi
          requests:
            cpu: 150m
            memory: 900Mi
        tag: 2.20.0
    version: ""

```