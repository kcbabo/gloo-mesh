module github.com/solo-io/gloo-mesh

go 1.16

replace (
	// pinned to solo-io's fork of cue version 95a50cebaffb4bdba8c544601d8fb867990ad1ad
	cuelang.org/go => github.com/solo-io/cue v0.4.1-0.20210622213027-95a50cebaffb

	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	github.com/envoyproxy/go-control-plane => github.com/envoyproxy/go-control-plane v0.9.9-0.20210115003313-31f9241a16e6
	github.com/spf13/viper => github.com/istio/viper v1.3.3-0.20190515210538-2789fed3109c

	k8s.io/api => k8s.io/api v0.20.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.4
	k8s.io/client-go => k8s.io/client-go v0.20.4

	k8s.io/kube-openapi => github.com/howardjohn/kube-openapi v0.0.0-20210104181841-c0b40d2cb1c8
	k8s.io/kubectl => k8s.io/kubectl v0.20.4
	k8s.io/utils => k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.7.0
)

require (
	cloud.google.com/go v0.80.0 // indirect
	cuelang.org/go v0.4.0
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/Masterminds/sprig/v3 v3.2.2 // indirect
	github.com/aws/aws-app-mesh-controller-for-k8s v1.1.1
	github.com/gertd/go-pluralize v0.1.1
	github.com/go-openapi/swag v0.19.7 // indirect
	github.com/gobuffalo/packd v1.0.0 // indirect
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-github/v32 v32.0.0
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1
	github.com/iancoleman/strcase v0.1.3
	github.com/lib/pq v1.9.0 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mitchellh/copystructure v1.1.1 // indirect
	github.com/mitchellh/hashstructure v1.0.0
	github.com/onsi/ginkgo v1.15.0 // indirect
	github.com/onsi/gomega v1.11.0 // indirect
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0 // indirect
	github.com/prometheus/common v0.20.0 // indirect
	github.com/pseudomuto/protoc-gen-doc v1.4.1
	github.com/pseudomuto/protokit v0.2.0
	github.com/rotisserie/eris v0.4.0
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/servicemeshinterface/smi-sdk-go v0.4.1
	github.com/sirupsen/logrus v1.7.0
	github.com/solo-io/anyvendor v0.0.4
	github.com/solo-io/external-apis v0.1.6
	github.com/solo-io/go-list-licenses v0.1.3
	github.com/solo-io/go-utils v0.21.9
	github.com/solo-io/k8s-utils v0.0.11 // indirect
	github.com/solo-io/protoc-gen-ext v0.0.15
	github.com/solo-io/skv2 v0.19.3
	github.com/solo-io/solo-apis v1.6.31
	github.com/solo-io/solo-kit v0.21.0
	github.com/spf13/cobra v1.1.3
	github.com/stoewer/go-strcase v1.2.0
	golang.org/x/oauth2 v0.0.0-20210413134643-5e61552d6c78 // indirect
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	google.golang.org/grpc v1.37.0
	google.golang.org/protobuf v1.27.1
	honnef.co/go/tools v0.0.1-2020.1.5 // indirect
	istio.io/api v0.0.0-20210420211535-1c598ea4139c
	istio.io/client-go v1.9.3-0.20210420212959-a7d9bf5db101
	istio.io/gogo-genproto v0.0.0-20210420211914-9cbf6943c732 // indirect
	k8s.io/api v0.21.0
	k8s.io/apiextensions-apiserver v0.21.0 // indirect
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v8.0.0+incompatible
	k8s.io/utils v0.0.0-20210305010621-2afb4311ab10 // indirect
	sigs.k8s.io/controller-runtime v0.9.0-alpha.1.0.20210412152200-442d3cad1e99
)
