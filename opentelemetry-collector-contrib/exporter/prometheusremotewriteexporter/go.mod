module github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter

go 1.20

require (
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/fsnotify/fsnotify v1.6.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/snappy v0.0.4
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.83.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.83.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.83.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheusremotewrite v0.83.0
	github.com/prometheus/prometheus v0.44.0
	github.com/stretchr/testify v1.8.4
	github.com/tidwall/wal v1.1.7
	go.opentelemetry.io/collector/component v0.83.0
	go.opentelemetry.io/collector/config/confighttp v0.83.0
	go.opentelemetry.io/collector/config/configopaque v0.83.0
	go.opentelemetry.io/collector/config/configtls v0.83.0
	go.opentelemetry.io/collector/confmap v0.83.0
	go.opentelemetry.io/collector/consumer v0.83.0
	go.opentelemetry.io/collector/exporter v0.83.0
	go.opentelemetry.io/collector/pdata v1.0.0-rcv0014
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.25.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/knadh/koanf v1.5.0 // indirect
	github.com/knadh/koanf/v2 v2.0.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20220423185008-bf980b35cac4 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/rs/cors v1.9.0 // indirect
	github.com/tidwall/gjson v1.10.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/tinylru v1.1.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/collector v0.83.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.83.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v0.83.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.83.0 // indirect
	go.opentelemetry.io/collector/config/internal v0.83.0 // indirect
	go.opentelemetry.io/collector/extension v0.83.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.83.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.0.0-rcv0014 // indirect
	go.opentelemetry.io/collector/processor v0.83.0 // indirect
	go.opentelemetry.io/collector/receiver v0.83.0 // indirect
	go.opentelemetry.io/collector/semconv v0.83.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.42.0 // indirect
	go.opentelemetry.io/otel v1.16.0 // indirect
	go.opentelemetry.io/otel/metric v1.16.0 // indirect
	go.opentelemetry.io/otel/trace v1.16.0 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/grpc v1.57.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/common => ../../internal/common

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal => ../../internal/coreinternal

replace github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry => ../../pkg/resourcetotelemetry

replace github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus => ../../pkg/translator/prometheus

replace github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheusremotewrite => ../../pkg/translator/prometheusremotewrite

retract (
	v0.76.2
	v0.76.1
	v0.65.0
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil => ../../pkg/pdatautil

replace github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest => ../../pkg/pdatatest
