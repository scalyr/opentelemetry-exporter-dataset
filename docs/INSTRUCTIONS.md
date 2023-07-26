# Instructions

We assume you have some familiarity with OpenTelemetry configuration. You can consult the [official documentation](https://opentelemetry.io/docs/collector/configuration/) for more information, and best practices.

For help, contact <support@dataset.com>.


## Installation

You can build from source, or pull a docker image.


### Build From Source

You can consult the documentation to [build a custom collector](https://opentelemetry.io/docs/collector/custom-collector/).

**(1)** [Go](https://go.dev/doc/install) is a prerequisite.

**(2)** Install the [builder](https://pkg.go.dev/go.opentelemetry.io/collector/cmd/builder) for the OpenTelemetry Collector. This generates a binary for a given configuration, and only includes necessary components:

```bash
go install go.opentelemetry.io/collector/cmd/builder@v0.77.0
```

**(3)** Configure components for the collector. An example with the minimum configuration. Data is received from  OTLP or log files, and logs export to DataSet:

```yaml
# For a list of all components:
# https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/cmd/otelcontribcol/builder-config.yaml

dist:
  name: otelcol-dataset
  description: Local OpenTelemetry Collector binary
  output_path: ./otelcol-dataset
  otelcol_version: 0.77.0

exporters:
  - gomod: go.opentelemetry.io/collector/exporter/loggingexporter v0.77.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/exporter/datasetexporter v0.77.0

receivers:
  - gomod: go.opentelemetry.io/collector/receiver/otlpreceiver v0.77.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.77.0

processors:
  - gomod: go.opentelemetry.io/collector/processor/batchprocessor v0.77.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.77.0
  - gomod: go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.77.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanmetricsprocessor v0.77.0

extensions:
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.77.0
  - gomod: go.opentelemetry.io/collector/extension/ballastextension v0.77.0
  - gomod: go.opentelemetry.io/collector/extension/zpagesextension v0.77.0
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage v0.77.0
    import: github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage

replaces:
  github.com/open-telemetry/opentelemetry-collector-contrib/exporter/datasetexporter v0.77.0 => github.com/scalyr/opentelemetry-collector-contrib/exporter/datasetexporter datasetexporter-latest
```

Note the `replaces` directive. The `datasetexporter` is merged to OpenTelemetry, but we recommend you pull the latest version from our repository, at `https://github.com/scalyr/opentelemetry-collector-contrib`.

You can add more components from the [contrib repository](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/cmd/otelcontribcol/builder-config.yaml).

**(4)** Save the config file, for example as `otelcol-builder.yaml`.

**(5)** Build the collector

Go to the `/bin` folder for the `builder`:

```bash
cd ~/go/bin
```

Then run:

```bash
./builder --config=/path/to/otelcol-builder.yaml
```

This creates an `otelcol-dataset` collector in the folder. See below to configure and run the collector.


### Pull a Docker Image

Docker images with the minimal configuration (see **3** above):

* GitHub - https://github.com/scalyr/opentelemetry-exporter-dataset/pkgs/container/opentelemetry-exporter-dataset
    ```bash
    docker pull ghcr.io/scalyr/opentelemetry-exporter-dataset:latest
    ```
For the latest version of the `datasetexporter`, we recommend you pull an image from these repositories, instead of `otel/opentelemetry-collector`.

The build datetime is a tag, in the format `YYYYMMDDtHHMMSS`.


## Configure the Collector

Your specific configuration depends on your pipeline. Consult the [official documentation](https://opentelemetry.io/docs/collector/configuration/) for more.

An example with the minimal configuration. The collector accepts logs from OTLP, batches them, logs to file, and exports to DataSet:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
      http:

processors:
  batch:

exporters:
  logging:
    loglevel: debug
  dataset:
    # read from env DATASET_URL
    dataset_url:
    # read from env DATASET_API_KEY
    api_key:

service:
  pipelines:
    logs:
      receivers: [ otlp ]
      processors: [ batch ]
      exporters: [ logging, dataset ]
```

### DataSet Configuration

You must set the `dataset_url` and `api_key` in the YAML, or as environmental variables:

- `dataset_url` (no default): The URL of the DataSet API. Usually `https://app.scalyr.com`. For EU accounts, set this to `https://eu.scalyr.com`. If not set, the `DATASET_URL` environmental variable is used.
- `api_key` (no default): Log into your DataSet account and go to the [API Keys](/keys) page. There you can view or generate a "Log Write Access" key. If not set, the `DATASET_API_KEY` environmental variable is used.


**More settings:**

- `buffer`
  - `max_lifetime` (default = `5`): The maximum delay in seconds between sending batches from the same source.
  - `group_by` (default = []): A list of attributes in the data, used to map data into sessions. The DataSet [addEvents API](https://app.scalyr.com/help/api#addEvents) requires all data to be part of a session. Each session cannot exceed 10 MB/sec, and we conservatively recommend 2.5 MB/sec per session. When `group_by` is not set, all of your data is exported in a single session. See below for more on setting this property.
  - `retry_initial_interval` (default = `5`): Time to wait in seconds after the first failure before retrying.
  - `retry_max_interval` (default = `30`): The upper bound on backoff for the retry logic, in seconds.
  - `max_elapsed_time` (default = `300`): The maximum amount of time, in seconds, to spend trying to send a buffer.

**Options from the [exporterhelper](https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/exporterhelper/README.md) exporter:**

- `retry_on_failure`
    - `enabled` (default = true)
    - `initial_interval` (default = `5`): Time to wait after the first failure before retrying, in seconds; ignored if `enabled` is `false`
    - `max_interval` (default = `30`): The upper bound on backoff for the retry logic, in seconds; ignored if `enabled` is `false`
    - `max_elapsed_time` (default = `300`): The maximum time to try to send a batch, in seconds; ignored if `enabled` is `false`
- `sending_queue`
    - `enabled` (default = true)
    - `num_consumers` (default = `10`): Number of consumers that dequeue batches; ignored if `enabled` is `false`
    - `queue_size` (default = `5000`): Maximum number of batches kept in memory; ignored if `enabled` is `false`. Calculate as `num_seconds * requests_per_second / requests_per_batch`, where:
        - `num_seconds` is the number of seconds to buffer, in case of a backend outage
        - `requests_per_second` is the average number of requests per second
        - `requests_per_batch` is the average number of requests per batch (if you use the
          [batch processor](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/batchprocessor), you can estimate this with the `batch_send_size` metric)
- `timeout` (default = `5`): Time to wait for each attempt to send data, in seconds


**For the persistent queue:**

- `sending_queue`
    - `storage` (default = none): When set, enables persistence and uses the component specified as a storage extension for the persistent queue. For example, `storage: file_storage`
- `sending_queue.queue_size` (default = `5000`): The maximum number of batches stored to disk. The default (5000) is the same default for in-memory buffering.


**More on `buffer.group_by`**:

A list of attributes in the data, used to map data into sessions. The DataSet [addEvents API](https://app.scalyr.com/help/api#addEvents) requires all data to be part of a session. Each session cannot exceed 10 MB/sec, and we conservatively recommend 2.5 MB/sec per session. When `group_by` is not set, all of your data is exported in a single session.

If your throughput is above 10 MB/sec, you must set this property, and we recommend you set it for a throughput of approximately 2.5 MB/sec per session. You can initially not set the property, view the attributes you can set in the DataSet UI (see the [fieldList](https://app.scalyr.com/help/log-overview#fieldList)), and then set the property. An example for a high-throughput Kubernetes implementation:
```yaml
group_by:
- attributes.container_id
- attributes.log.file.path
- body.map.kubernetes.container_hash
- body.map.kubernetes.pod_id
- body.map.kubernetes.docker_id
- body.map.stream
```
Note that if you have many containers and set the container id, you may separate your data into too many sessions, and lower throughput.

## Recommended Components

We recommend these components. You must include them in the collector build, and the collector configuration:

* [File Log Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver)
    * Tails and parses logs from files.
* [OTLP Receiver](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver/otlpreceiver)
    * Receives data from gRPC or HTTP in OTLP format.
* [Batch Processor](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/batchprocessor)
    * Accepts spans, metrics, or logs, and batches them. Batching compresses the data, and decreases the number of outgoing connections.
* [Logging Exporter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/loggingexporter)
    * Exports data to the console via zap.Logger.
* [File Storage](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/storage/filestorage)
    * Persists state to the local file system.
* [Memory Limiter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/memorylimiterprocessor)
    * Performs periodic checks of memory usage; refuses data if memory exceeds the limit, and forces Garbage Collection to decrease memory consumption.
* [zPages](https://github.com/open-telemetry/opentelemetry-collector/tree/main/extension/zpagesextension)
    * Enables an HTTP endpoint with live data to debug components. All core exporters and receivers have some zPage instrumentation.
* [Memory Ballast](https://github.com/open-telemetry/opentelemetry-collector/tree/main/extension/ballastextension)
    * Enables applications to configure memory ballast for the process.
* [Health Check](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/healthcheckextension)
    * Enables an HTTP URL to check the status of the OpenTelemetry Collector. Can be used as a liveness and/or readiness probe on Kubernetes.

### Example

To build a collector that:
* Receives data from OTLP, and also reads logs from specified log files
* Has zPages and Health Check for monitoring and alerting
* Sends a batch at least every 2s, with a a maximum of 10000 records per batch
* Logs data to the console
* Sends data to DataSet
    * Sends to `https://app.scalyr.com`
    * Sends data from a single source at least every 5s
    * Does not loose data when DataSet is temporarily unavailable, for up to 5 minutes

      We expect to receive 1000 logs per second. With a batch timeout of 2s, each batch will have 2000 logs. To keep data up to 5 minutes (300s) when DataSet is unavailable, we must retain 30,0000 logs, and 150 batches. The default for `sending_queue.queue_size` is 5000 batches, which is significantly more than the 150 expected.

      The configuration:

```yaml
receivers:
  # receive data from OTLP
  otlp:
    protocols:
      grpc:
      http:

  # receive also data produced by docker containers
  filelog/containers:
    include: [ "/var/lib/docker/containers/*/*.log" ]
    start_at: end
    include_file_path: true
    include_file_name: false
    operators:
      - type: json_parser
        id: parser-docker
        output: extract_metadata_from_filepath
        timestamp:
          parse_from: attributes.time
          layout: '%Y-%m-%dT%H:%M:%S.%LZ'
      # Extract metadata from file path
      - type: regex_parser
        id: extract_metadata_from_filepath
        regex: '^.*containers/(?P<container_id>[^_]+)/.*log$'
        parse_from: attributes["log.file.path"]
        output: parse_body
      - type: move
        id: parse_body
        from: attributes.log
        to: body
        output: add_source

processors:
  # create batches with at most 10k records
  # but sent them every 2s even if they are smaller
  batch:
    send_batch_size: 10000
    timeout: 2s

extensions:
  # enable debugging
  zpages:
  # enable health check
  health_check:
  # store unprocessed data persistently
  # do not forget to create this directory
  # and make sure that its readable and writeable
  file_storage:
    directory: /tmp/otc/

exporters:
  # log to the console
  logging:
  # log to the dataset
  dataset:
    # send data to https://app.scalyr.com
    dataset_url: https://app.scalyr.com
    api_key: <your-api-key>
    buffer:
      # each log in each container will be considered as separate source
      group_by:
        - attributes.container_id
        - attributes.log.file.path
    retry_on_failure:
      max_interval: 300
      max_elapsed_time: 300
    sending_queue:
      storage: file_storage

service:
  extensions: [zpages, health_check, file_storage]
  pipelines:
    logs:
      receivers: [ otlp, filelog/containers ]
      processors: [ batch ]
      exporters: [ logging, dataset ]
```

## Run the Collector

To run the collector:

```bash
./otelcol-dataset --config=/path/to/config.yaml
```
