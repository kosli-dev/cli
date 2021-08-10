module github.com/merkely-development/reporter

go 1.13

require (
	github.com/aws/aws-sdk-go v1.40.17
	github.com/aws/aws-sdk-go-v2/config v1.6.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.7.0
	github.com/hashicorp/go-retryablehttp v0.6.8
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/utils v0.0.0-20210305010621-2afb4311ab10 // indirect
)

replace k8s.io/client-go => k8s.io/client-go v0.19.2
