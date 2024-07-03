module github.com/open-telemetry/opentelemetry-collector-contrib/cmd/githubgen

go 1.21.0

require (
	github.com/google/go-github/v62 v62.0.0
	go.opentelemetry.io/collector/confmap v0.102.2-0.20240606174409-6888f8f7a45f
	go.opentelemetry.io/collector/confmap/provider/fileprovider v0.102.2-0.20240606174409-6888f8f7a45f
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
)
