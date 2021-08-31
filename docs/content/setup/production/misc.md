# Misc Istio Advice

## Tuning Istio Service Discovery

## Tuning Sidecar

* pre-stop delay
* enableProtocolSniffing false

## Adding Istio to an Existing Production Cluster

Avoid

* STRICT PeerAuthentication
* outboundTrafficPolicy REGISTRY_ONLY mode
* Global Authorization Policy

## EnvoyFilter

* Naming
* Scope down as much as possible

## Gateways

* Separate IstioOperator
* Different namespace than `istio-system`
* Multilple gateways with zone affinity

## Tracing and Telemetry

* Disable if not used
* Tracing sampling %
* Disable logging

# Use revisions
