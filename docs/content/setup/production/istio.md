# Istio In Production

## Recommended Architecture

This is the architecture diagram of our prodcuction Istio deployment. The following documentation will go into more detail about how to deploy and mange Istio with this configuration.

![Istio Production Architecture](../../img/production-istio_architecture.png)

## Namespaces

Below is the recommended naming scheme for Istio namespaces.

* `istio-config` - Istios "rootNamespace" where configuration will be read
* `istio-system` - Deployment of the istio control plane
* `istio-gateways` - Default namespace for deploying gateway resources
  * `istio-ingress` - Non-shared istio-ingressgateway deployment namespace
  * `istio-egress` - Non-shared istio-egressgateway deployment namespace
  * `istio-eastwest` - Non-shared istio-eastwestgateway deployment namespace

* For more Information see: [Istio Namespaces](./namespaces.md)

## Deployment

You can deploy Istio a number of ways but it is recommended to deploy the Operator and configure it with the `IstioOperator` config. If you use a helm based deployment model you can still deploy it with a helm chart provided by Istio.

As shown in the above diagram. We first deploy the IstioOperator to the `istio-operator` namespace. We then can deploy the Istio control plane with an IstioOperator configuration. Once completed we can go ahead and deploy the required gateways to the `istio-gateways` namespace. For a full set of instructions on how to set these up, see below.

### Full Installation Details

1. [Deploying the Istio Operator](./operator_deployment.md)
2. [Deploying Istio Control Plane](./istiod_deployment.md)
3. [Deploying Gateways](./gateway_deployment.md)


## Configuration Management

Istio has implemented its configuration in a way to allow admins to set mesh wide policies that allows individual service owners to override them for their workloads. For further reading on how to manage configuration see below. 

* [Configuration Management](./config_management.md)

## Upgrading

## Tuning Istio Service Discovery

## Sidecar Properties

## Access Logging

## Metrics

## Adding Istio to an Existing Production Cluster

Avoid 

* STRICT PeerAuthentication
* outbound REGISTRY_ONLY mode
* GLobal Authorization Policy

EnvoyFilter Naming