# example run of istio test framework

* currently installs istio operator based on `pkg/test/manifests/operator/gloo-mesh-istio.yaml` for every cluster provided in the kube config below.
* then installs 3 echo applications to those clusters and tests that they can talk to each other within the same cluster.

example run
```shell
FLAT_NETWORKING_ENABLED=false RUN_INTEGRATION=true GLOO_MESH_ENTERPRISE_VERSION=1.1.0-rc2 GLOO_MESH_LICENSE_KEY=$GLOO_MESH_LICENSE_KEY go test -v github.com/solo-io/gloo-mesh/test/integration/multi-cluster/routing \
  -args --istio.test.kube.config=/Users/nick/.kube/mp,/Users/nick/.kube/cp \
    --istio.test.nocleanup=false \
    --istio.test.hub=gcr.io/istio-enterprise \
    --istio.test.tag=1.10.3
```

## Cluster setup using kind (works better for virtualGateway tests)

```sh
deploy-kind() {



  number=$1
  name=$2
  region=$3
  zone=$4
  twodigits=$(printf "%02d\n" $number)

  if [ -z "$3" ]; then
    region=us-east-1
  fi

  if [ -z "$4" ]; then
    zone=us-east-1a
  fi

  if hostname -I 2>/dev/null; then
    myip=$(hostname -I | awk '{ print $1 }')
  else
    myip=$(ipconfig getifaddr en0)
  fi

  reg_name='kind-registry'
  reg_port='5000'
  running="$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)"
  if [ "${running}" != 'true' ]; then
    docker run \
      -d --restart=always -p "127.0.0.1:${reg_port}:5000" --name "${reg_name}" \
      registry:2
  fi

  cache_name='kind-cache'
  cache_port='5000'
  running="$(docker inspect -f '{{.State.Running}}' "${cache_name}" 2>/dev/null || true)"
  if [ "${running}" != 'true' ]; then
    cat > $HOME/.kube/kind/config.yml <<EOF
  version: 0.1
  proxy:
    remoteurl: https://registry-1.docker.io
  log:
    fields:
      service: registry
  storage:
    cache:
      blobdescriptor: inmemory
    filesystem:
      rootdirectory: /var/lib/registry
  http:
    addr: :5000
    headers:
      X-Content-Type-Options: [nosniff]
  health:
    storagedriver:
      enabled: true
      interval: 10s
      threshold: 3
EOF

    docker run \
      -d --restart=always -v $HOME/.kube/kind/config.yml:/etc/docker/registry/config.yml --name "${cache_name}" \
      registry:2
  fi

  cat << EOF > $HOME/.kube/kind/kind${number}.yaml
  kind: Cluster
  apiVersion: kind.x-k8s.io/v1alpha4
  featureGates:
  #  TokenRequest: true
    EphemeralContainers: true
  nodes:
  - role: control-plane
    extraPortMappings:
    - containerPort: 6443
      hostPort: 70${twodigits}
  networking:
    serviceSubnet: "10.0${twodigits}.0.0/16"
    podSubnet: "10.1${twodigits}.0.0/16"
  kubeadmConfigPatches:
  - |
    apiVersion: kubeadm.k8s.io/v1beta2
    kind: ClusterConfiguration
    apiServer:
      extraArgs:
        service-account-signing-key-file: /etc/kubernetes/pki/sa.key
        service-account-key-file: /etc/kubernetes/pki/sa.pub
        service-account-issuer: api
        service-account-api-audiences: api,vault,factors
    metadata:
      name: config
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true,topology.kubernetes.io/region=${region},topology.kubernetes.io/zone=${zone}"
  containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
      endpoint = ["http://${reg_name}:${reg_port}"]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
      endpoint = ["http://${cache_name}:${cache_port}"]
EOF

  kind create cluster --name kind${number} --config $HOME/.kube/kind/kind${number}.yaml

  ipkind=$(docker inspect kind${number}-control-plane | jq -r '.[0].NetworkSettings.Networks[].IPAddress')
  networkkind=$(echo ${ipkind} | awk -F. '{ print $1"."$2 }')

  kubectl config set-cluster kind-kind${number} --server=https://${myip}:70${twodigits} --insecure-skip-tls-verify=true

  kubectl --context=kind-kind${number} apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.3/manifests/namespace.yaml
  kubectl --context=kind-kind${number} apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.3/manifests/metallb.yaml
  kubectl --context=kind-kind${number} create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)"

  cat << EOF > $HOME/.kube/kind/metallb${number}.yaml
  apiVersion: v1
  kind: ConfigMap
  metadata:
    namespace: metallb-system
    name: config
  data:
    config: |
      address-pools:
      - name: default
        protocol: layer2
        addresses:
        - ${networkkind}.0${twodigits}.1-${networkkind}.0${twodigits}.254
EOF

  kubectl --context=kind-kind${number} apply -f $HOME/.kube/kind/metallb${number}.yaml

  docker network connect "kind" "${reg_name}" || true
  docker network connect "kind" "${cache_name}" || true

  cat <<EOF | kubectl apply -f -
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: local-registry-hosting
    namespace: kube-public
  data:
    localRegistryHosting.v1: |
      host: "localhost:${reg_port}"
      help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

  kubectl config delete-cluster ${name} || true
  kubectl config delete-user ${name} || true
  kubectl config delete-context ${name} || true

  kubectl config rename-context kind-kind${number} ${name} || true

}

kind-up() {
  deploy-kind 1 mp
  kind get kubeconfig --name kind1 > $HOME/.kube/mp || true
  deploy-kind 2 cp us-west us-west-1
  kind get kubeconfig --name kind2 > $HOME/.kube/cp || true
}

kind-down(){
  kind delete cluster --name kind1
  kind delete cluster --name kind2
  rm  $HOME/.kube/mp || true
  rm  $HOME/.kube/cp || true
}


```

## cluster setup script using k3d

```sh
#!/bin/bash
network=demo-1

# create docker network if it does not exist
docker network create $network || true

## management plane cluster exposes port 9000 (unused currently)
k3d cluster create mp --image "rancher/k3s:v1.20.2-k3s1"  --k3s-server-arg "--disable=traefik" --network $network
kube_ctx=k3d-mp
k3d kubeconfig get mp > ~/.kube/mp

kubectl label node $kube_ctx-server-0 topology.kubernetes.io/region=us-west-1 --context $kube_ctx
kubectl label node $kube_ctx-server-0 topology.kubernetes.io/zone=us-west-1b --context $kube_ctx

## control plane cluster (us-east) exposes port 9010 (unused currently)
k3d cluster create cp-us-east --image "rancher/k3s:v1.20.2-k3s1"  --k3s-server-arg "--disable=traefik" --network $network
k3d kubeconfig get cp-us-east > ~/.kube/cp-us-east
kube_ctx=k3d-cp-us-east

kubectl label node $kube_ctx-server-0 topology.kubernetes.io/region=us-east-1 --context $kube_ctx
kubectl label node $kube_ctx-server-0 topology.kubernetes.io/zone=us-east-1a --context $kube_ctx

```

### teardown
```shell
#!/bin/bash
network=demo-1
k3d cluster delete mp
rm  ~/.kube/mp
  
k3d cluster delete cp-us-east
rm  ~/.kube/cp-us-east

docker network rm $network

```

## Cluster / App Topology
This is the default deployment of the test engine. there are no assumptions about the connectivity of the applications.
Each test will determine the routing between them. 
![Application Topology](images/istio-test-framework-cluster-arch.png)


## Use Cases


### Outside client
The client is outside the cluster and outside the mesh. They need to rely on third party DNS to resolve a gateway that hosts the API.
![Outside Client](images/use-case-1-outside-client.png)


### In-Cluster In-Mesh
The client is inside the same cluster as the application and is a part of the mesh.
![In-Cluster In-Mesh](images/use-case-2-in-cluster-in-mesh.png)


### Outside cluster In-Mesh
The client is a part of the mesh but does not contain a target application in its local cluster
![Outside cluster In-Mesh](images/use-case-3-out-of-cluster-in-mesh.png)
  

### Multi-Cluster Application 
The application exists in many clusters.
![Multi-Cluster Application](images/use-case-4-multi-cluster-app.png)


### Multi-Cluster Application Hybrid Topology 
The application exists in many clusters but some are flat networks and some require ingress gateways.
![Multi-Cluster Application Hybrid Topology](images/use-case-5-multi-cluster-app-hybrid.png)


### Multi-Cluster Application With ELB 
The application exists in many clusters but not all. The ELB will still route requests to clusters without an application. 
![Multi-Cluster Application With ELB](images/use-case-6-multi-cluster-app-elb.png)

  
### In-Cluster Out of Mesh 
The client is in the same cluster as the application but not a part of the mesh. 
![Multi-Cluster Application With ELB](images/use-case-7-in-cluster-out-of-mesh.png)
  

### Tiered Gateways 
Gateway that routes to other gateways within the mesh.
![Tiered Gateways](images/use-case-7-in-cluster-out-of-mesh.png)
  

### Egress Gateways 
Routing egress traffic through a specified gateway
![Egress Gateways](images/use-case-8-tiered-gateways.png)