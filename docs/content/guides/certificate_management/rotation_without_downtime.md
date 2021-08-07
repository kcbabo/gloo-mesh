---
title: Rotation Without Downtime
menuTitle: Rotation Without Downtime
description: Rotating root certificates without downtime using Gloo Mesh Enterprise.
weight: 50
---

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

Root cert rotation without downtime is an incredibly important aspect to any secure system. There are 2 main instances 
when this feature becomes necessary.
1. The root-cert, or an intermediate certificate authority signed by the root, has been compromised. 
2. The root-cert is going to expire.

The first situation comes with more dire consequences if not handled immediately, but both situations require the 
same work to be done. 

## What is Cert Rotation.

We briefly touched on why it might be necessary rotate our root-cert, but what exactly does it entail?

Rotating the root-certificate is the process of swapping the system from trusting one root-certificate, to another. 
That sounds simple on it's face, but the actual process requires quite a few steps. The most important part of any 
root-cert rotation is zero downtime. What does that mean? It means that throughout the rotation process, all 
services in the system should continue to trust each other. But how can we do that with two different certificates?

There are two main ways to accomplish this, but we will focus on the more popular one today. The one we will use is 
a three-step process called mutli-root. The most important part of this process is that if traffic begins to fail 
for any reason during one of the steps, the process can be rolled back. As mentioned this method has three steps, 
they are:
1. Adding new root. During this first stage all services/workloads in the system are given the new root-cert in 
   addition to the old root-cert, to ensure that they can trust either of them. This is especially important for 
   systems where mTLS is used heavily, as both client and server need to trust both roots.
2. Propogating new cert chain. During this second stage the services/workloads are given new certificates, signed by 
   the new root-certs private key. In order to ensure maximum safety, this should be done one at a time. Without 
   step 1, this step would fail, but since all of our services/workloads trust both roots, they will be able to 
   continue communicating during the entirety of step 2. This is the most dangerous step, and should be done with 
   the most caution.
3. Deleting the old root. During this third stage we remove the old root-certificate from the list of trusted roots. 
   This ensures that all workloads/services only trust certificates signed by the new root private key. Once this 
   has been completed, rotation is done, and the old root may be deleted entirely.

### Cert Rotation in Istio

Cert rotation is Istio follows the same basic principle outlined in the previous section, but with an Istio twist!

For more information on rotating certificates in Istio, see [Christian Posta's blog](https://blog.christianposta.
com/diving-into-istio-1-6-certificate-rotation/). This blog is for a deprecated Istio version (1.6), but the concept is 
still the same.

The Gloo Mesh rotation workflow follows the same process as the aforementioned blog, but on a multi cluster scale.

## Demo!

Now that we have briefly explained what cert rotation is, and why it's so important, let's quickly go over how we 
can accomplish this complex task easily with Gloo Mesh Enterprise.

### Before you begin

This guide assumes the following:

* Gloo Mesh Enterprise is [installed in relay mode and running on the `cluster-1`]({{% versioned_link_path fromRoot="/setup/install-gloo-mesh" %}})
* `gloo-mesh` is the installation namespace for Gloo Mesh
* `enterprise-networking` is deployed on `cluster-1` in the `gloo-mesh` namespace and exposes its gRPC server on port 9900
* `enterprise-agent` is deployed on both clusters and exposes its gRPC server on port 9977
* Both `cluster-1` and `cluster-2` are [registered with Gloo Mesh]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
* Istio is [installed on both clusters]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
* `istio-system` is the root namespace for both Istio deployments
* The `bookinfo` app is [installed into the two clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}}) under the `bookinfo` namespace
* the following environment variables are set:

```shell
CONTEXT_1="cluster_1's_context"
CONTEXT_2="cluster_2's_context"
```

### Generated root of trust.

First things first, let's take a look at our system's root of trust. In our `VirtualMesh` we see that our mtlsConfig 
has what we call a generated certificate authority. What does that mean exactly? Well, similar to Istio, Gloo Mesh 
has the ability to generate a CA certificate, if one isn't provided to us. This is the default setting for most 
initial Gloo Mesh installations. However, this type of root-cert most definitely should not be used in a production 
use-case, as deleting the root-cert secret created by Gloo Mesh, could break trust for the entire system. For this 
reason we recommend storing the root-certificate in an external secure location, and allowing Gloo Mesh to access it,
or a CA signed by that root. In this demo we are going to migrate away from the Gloo Mesh generated root-cert to a 
more permanent and secure root.

The root-cert we will be using for this example will be generated and store in [AWS ACM](https://aws.amazon.
com/certificate-manager/), but there are many other 
services that can be used for this purpose.

# TODO
kubectl get virtualmesh

