# Istio In Production

# Recommended Architecture

## Namespaces

* `istio-config` - Istios "rootNamespace" where configuration will be read
* `istio-system-<revision>` - 
* `istio-gateways` - Default namespace for deploying gateway resources

# Deployment

You can deploy Istio a number of ways but it is recommended to deploy the Operator and configure it with the `IstioOperator` config. If you use a helm based deployment model you can still deploy it with a helm chart


Deploying Istio Operator via helm
```
helm install istio-operator manifests/charts/istio-operator \
  --set operatorNamespace=istio-operator \
  --set watchedNamespaces="istio-namespace1\,istio-namespace2"
```


# Upgrading

# Tuning Istio Service Discovery

# Sidecar Properties

# Access Logging

# Metrics

# Adding Istio to an Existing Production Cluster

Avoid 
* STRICT PeerAuthentication
* outbound REGISTRY_ONLY mode
* GLobal Authorization Policy

EnvoyFilter Naming



# Unknowns
* Passing gateway IstioOperator specs between versions of the operator
* Root namespace vs config namespace
 * If a company opts for using a single namespace for all istio confiuguration how does that work with a root namespace? For example if they want specific PeerAuthentication Policies in a namespace does it have to exist there?


* Labelling pods to get envoy filters is recommended by label selector
* Ops people manage envoy filters, service owner chooses them

* Label based configurations recommendation
  *  Examples for global resources that could be selected by service owners
  * I have a Access log EnvoyFilter and devs can label istio-logs: "enabled" and gets logs