module github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/awsutil

go 1.20

require (
	github.com/aws/aws-sdk-go v1.44.329
	github.com/stretchr/testify v1.8.4
	go.uber.org/zap v1.25.0
	golang.org/x/net v0.14.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract (
	v0.76.2
	v0.76.1
	v0.65.0
)
