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

Rotating the root-certificate requires 

For more information on rotating certificates in Istio, see [Christian Posta's blog](https://blog.christianposta.
com/diving-into-istio-1-6-certificate-rotation/). This blog is for a deprecated Istio version (1.6), but the concept is 
still the same.

