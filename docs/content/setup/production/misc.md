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

## EnvoyFilter Naming

## Gateways

* Separate IstioOperator
* Different namespace than `istio-system`
* Multilple gateways with zone affinity

## Performance

* Tracing sampling %

# Use revisions
