module github.com/scalyr/opentelemetry-exporter-dataset

replace (
	github.com/open-telemetry/opentelemetry-collector-contrib => ./opentelemetry-collector-contrib/
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/datasetexporter => ./datasetexporter
	github.com/scalyr/dataset-go => ./dataset-go
)

go 1.21
