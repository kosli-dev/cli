module github.com/merkely-development/reporter

go 1.13

require (
	github.com/aws/aws-sdk-go v1.40.17
	github.com/aws/aws-sdk-go-v2/config v1.6.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.7.0
	github.com/containerd/containerd v1.5.8 // indirect
	github.com/docker/docker v20.10.10+incompatible
	github.com/go-git/go-git/v5 v5.4.2
	github.com/google/go-github/v42 v42.0.0
	github.com/hashicorp/go-retryablehttp v0.6.8
	github.com/joshdk/go-junit v0.0.0-20210226021600-6145f504ca0d
	github.com/mattn/go-shellwords v1.0.3
	github.com/maxcnunes/httpfake v1.2.4
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
	k8s.io/kubernetes v1.22.3
	sigs.k8s.io/kind v0.11.1
)

replace k8s.io/client-go => k8s.io/client-go v0.22.3

replace k8s.io/api => k8s.io/api v0.22.3

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.3

replace k8s.io/apimachinery => k8s.io/apimachinery v0.22.4-rc.0

replace k8s.io/apiserver => k8s.io/apiserver v0.22.3

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.3

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.22.3

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.22.3

replace k8s.io/code-generator => k8s.io/code-generator v0.22.4-rc.0

replace k8s.io/component-base => k8s.io/component-base v0.22.3

replace k8s.io/component-helpers => k8s.io/component-helpers v0.22.3

replace k8s.io/controller-manager => k8s.io/controller-manager v0.22.3

replace k8s.io/cri-api => k8s.io/cri-api v0.23.0-alpha.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.22.3

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.22.3

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.22.3

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.22.3

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.22.3

replace k8s.io/kubectl => k8s.io/kubectl v0.22.3

replace k8s.io/kubelet => k8s.io/kubelet v0.22.3

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.22.3

replace k8s.io/metrics => k8s.io/metrics v0.22.3

replace k8s.io/mount-utils => k8s.io/mount-utils v0.22.4-rc.0

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.22.3

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.22.3

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.22.3

replace k8s.io/sample-controller => k8s.io/sample-controller v0.22.3

replace github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2

replace github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.3

replace github.com/containerd/containerd => github.com/containerd/containerd v1.5.9
