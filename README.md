# DataSet OpenTelemetry Collector Exporter Build Infrastructure

[![Check code quality](https://github.com/scalyr/opentelemetry-exporter-dataset/actions/workflows/code-quality.yaml/badge.svg)](https://github.com/scalyr/opentelemetry-exporter-dataset/actions/workflows/code-quality.yaml)
[![TruffleHog Secrets Scan](https://github.com/scalyr/opentelemetry-exporter-dataset/actions/workflows/secrets-scanner.yaml/badge.svg)](https://github.com/scalyr/opentelemetry-exporter-dataset/actions/workflows/secrets-scanner.yaml)
[![E2E Test - Reliability](https://github.com/scalyr/opentelemetry-exporter-dataset/actions/workflows/e2e-test.yaml/badge.svg)](https://github.com/scalyr/opentelemetry-exporter-dataset/actions/workflows/e2e-test.yaml)
[![Benchmarks](https://github.com/scalyr/opentelemetry-exporter-dataset/actions/workflows/benchmarks.yaml/badge.svg)](https://github.com/scalyr/opentelemetry-exporter-dataset/actions/workflows/benchmarks.yaml)
[![Release Image](https://github.com/scalyr/opentelemetry-exporter-dataset/actions/workflows/release-image.yaml/badge.svg)](https://github.com/scalyr/opentelemetry-exporter-dataset/actions/workflows/release-image.yaml)
[![codecov](https://codecov.io/github/scalyr/opentelemetry-exporter-dataset/graph/badge.svg?token=EB9MPROGBO)](https://codecov.io/github/scalyr/opentelemetry-exporter-dataset)
[![Docker Pulls](https://img.shields.io/docker/pulls/scalyr/opentelemetry-collector-contrib)](https://hub.docker.com/r/scalyr/opentelemetry-collector-contrib/tags)

Main purpose of this repository is to allow testing of changes in the [dataset-go](./dataset-go) library or
[open-telemetry-contrib](./scalyr-opentelemetry-collector-contrib) exporter before we publish them.
This testing is usually done by either running existing "benchmarks" from the [scripts](scripts) folder or
using them as inspiration for ad hoc testing.

## Dependencies

If you are running targets below on OS X you also need GNU commands aka coreutils package installed
and in your $PATH. See https://formulae.brew.sh/formula/coreutils for more info.

## Prebuilt Docker Images

This repository contains 2 sets of pre-built Docker images for ``amd64`` and ``arm64`` architectures:

1. Pre-built image with the latest version of the dataset exporter plugin and some other basic otel
   components - [GitHub](https://github.com/scalyr/opentelemetry-exporter-dataset/pkgs/container/opentelemetry-exporter-dataset).
2. Pre-built image with the latest development version of the dataset exporter plugin
   (`datasetexporter-latest` branch from [our fork](https://github.com/scalyr/opentelemetry-collector-contrib/))
   with all other upstream otel contrib components - [Docker Hub](https://hub.docker.com/r/scalyr/opentelemetry-collector-contrib/tags).

Second image can act as a direct drop-in replacement for the upstream
`otel/opentelemetry-collector-contrib` image (e.g. in an upstream otel collector helm chart or
similar).

## Official Docker Image

You can use the latest official docker image as well:
1. Prepare config for collector:
    * Update [config.yaml](config.yaml)
    * See [documentation](scalyr-opentelemetry-collector-contrib/exporter/datasetexporter/README.md) for configuration option
2. Run image:
   * Run: `DATASET_URL=https://app.scalyr.com/ DATASET_API_KEY=FOO make exporter-docker-official-run`
3. Verify it:
    * Run: `make push-logs MSG="nejneobhospodarovavatelnejsi"`
    * Search for the word "nejneobhospodarovavatelnejsi" in [DataSet](https://app.scalyr.com/events?filter=nejneobhospodarovavatelnejsi)

## Build Collector

If you do not want to use prebuilt image, you can build your custom image.

Official documentation provides instructions how to build your own custom image:
  * https://opentelemetry.io/docs/collector/custom-collector/ - official documentation
  * https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder - builder

### Without Docker

If you want to use build your own collector, you can use following instructions:
  1. Install builder:
     * Run `go install go.opentelemetry.io/collector/cmd/builder@latest`
  2. Prepare builder config:
     * Example [otelcol-builder.yaml](otelcol-builder.yaml)
  3. Generate source code:
     * Run `make exporter-normal-build`
       * which runs `builder --config=otelcol-builder.yaml`
     * If you receive "command not found" or similar error, this likely indicates `$GOPATH/bin` is not in your search `$PATH`. You can find that by adding contents below to your `~/.bash_profile` or `~/.zshrc` config:
       * ```bash
         export GOPATH=/Users/$USER/go
         export PATH=$GOPATH/bin:$PATH
         ```
  4. Prepare config for collector:
     * Example [config.yaml](config.yaml)
     * See [documentation](scalyr-opentelemetry-collector-contrib/exporter/datasetexporter/README.md) for configuration option
  5. Execute collector:
     * Run: `DATASET_URL=https://app.scalyr.com/ DATASET_API_KEY=FOO otelcol-dataset/otelcol-dataset --config config.yaml`
  6. Verify it:
     * Run: `make push-logs MSG="nejneobhospodarovavatelnejsi"`
     * Search for the word "nejneobhospodarovavatelnejsi" in [DataSet](https://app.scalyr.com/events?filter=nejneobhospodarovavatelnejsi)

Alternatively instead of steps 5 and 6, you may execute `make exporter-normal-run`.

### With Docker

You can use prepared make targets:
  1. Build image using [Dockerfile](Dockerfile):
     * Run: `make exporter-docker-build`
  2. Prepare config for collector:
     * Update [config.yaml](config.yaml)
     * See [documentation](scalyr-opentelemetry-collector-contrib/exporter/datasetexporter/README.md) for configuration option
  3. Run image:
     * Run: `DATASET_URL=https://app.scalyr.com/ DATASET_API_KEY=FOO make exporter-docker-run`
  4. Verify it:
     * Run: `make push-logs MSG="nejneobhospodarovavatelnejsi"`
     * Search for the word "nejneobhospodarovavatelnejsi" in [DataSet](https://app.scalyr.com/events?filter=nejneobhospodarovavatelnejsi)

## Sample Data

To push some sample data to the collector via OTLP protocol, you may use following tasks:
  * `make push-logs` - to push logs with specified message
  * `make push-linked-trace-with-logs` - to push traces that are linked with logs

## Configure

For the configuration option you should check [documentation](datasetexporter/README.md).

## Development

1. Update all the repository subtrees to the latest version:
   * ```bash
     make subtrees-pull
     ```
3. Build all images:
   * ```bash
     make docker-build
     ```
4. Run e2e test:
   * ```bash
     make test-e2e
     ```

## Testing Changes Locally

Once you are familiar with building collectors binary, docker image, and executing e2e
tests from the [scripts](scripts) folder you can use this repository to test changes
from your development branches in the [dataset-go](./dataset-go) library or
[open-telemetry-contrib](./scalyr-opentelemetry-collector-contrib) exporter.

Workflow is:
1. Checkout your development branch in a subtree (e.g. `./dataset-go` or `./scalyr-opentelemetry-collector-contrib` directory)
2. Build collector
3. Run collector with your changes

## Continous Integration and Delivery (CI/CD)

### Benchmarks

We run benchmarks as part of every pull request and main branch run. Results are available in a
pretty formatted Markdown format as part of a job summary - [example run](https://github.com/scalyr/opentelemetry-exporter-dataset/actions/runs/5550337166).

Documentation for the meaning of the columns is in the [documentation](scripts/e2e/README.md).


In addition to that, we also run benchmarks on every main branch push / merge and store those
results in a dedicated [benchmark-results-dont-push-manually](https://github.com/scalyr/opentelemetry-exporter-dataset/blob/benchmark-results-dont-push-manually/benchmark-results.json) branch. This branch is an orphan an
only meant to store benchmark results (only CI/CD should push to this branch).

Results are stored in a line break delimited JSON format (serialized JSON string per line).

## Security

For information on how to report security vulnerabilities, please see [SECURITY.md](SECURITY.md).

## Copyright, License, and Contributors Agreement

Copyright 2023 SentinelOne, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this work except in
compliance with the License. You may obtain a copy of the License in the [LICENSE](LICENSE.txt) file, or at:

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

By contributing you agree that these contributions are your own (or approved by your employer) and you
grant a full, complete, irrevocable copyright license to all users and developers of the project,
present and future, pursuant to the license of the project.
