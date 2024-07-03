module github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/proxy

go 1.21.0

require (
	github.com/aws/aws-sdk-go v1.53.11
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.102.0
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/collector/config/confignet v0.102.2-0.20240606174409-6888f8f7a45f
	go.opentelemetry.io/collector/config/configtls v0.102.2-0.20240606174409-6888f8f7a45f
	go.uber.org/zap v1.27.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.9.1-0.20240606174409-6888f8f7a45f // indirect
	go.opentelemetry.io/collector/featuregate v1.9.1-0.20240606174409-6888f8f7a45f // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/common => ../../../internal/common

retract (
	v0.76.2
	v0.76.1
	v0.65.0
)
