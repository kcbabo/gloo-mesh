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

## Upgrading Istio

Due to the complexity up upgrading Istio and to prevent downtime, it is recommended that the you follow the `Full Installation Details` above. The following documentation will show you how to upgrade Istio with no downtime using the Istio Operator for both the control plane and gateways.

```txt
Upgrading across more than one minor version (e.g., 1.6.x to 1.8.x) in one step is not officially tested or recommended.
```

### [How To Upgrade Istio Using Operators](./upgrade_istio/upgrade.md)

[Upgrade Istio Official Documenation](https://istio.io/latest/docs/setup/upgrade/)

## Misc Advice

Other good practices can be found here

* [Good Istio Practices](./misc.md)
