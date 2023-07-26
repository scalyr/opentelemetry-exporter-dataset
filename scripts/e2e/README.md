# E2E Test

This is initial dummy E2E test, that:
* executes two images - open telemetry collector and [dummy API server](../../dummy-server)
* uses [logs generator](../../logs-generator) to generate data
* checks, that number of received events is same as number of generated lines

## Usage

* Build image, if you have made some changes:
  * ```bash
    (cd ../..; make docker-build)
    ```
* Run tests:
  * ```bash
    ./test.sh
    ```

## Details

This script runs an E2E test for a specific number of logs files with specified number of lines and their lengths. It works in the following way:
* generates test data using [logs generator](../../logs-generator)
* starts [dummy API server](../../dummy-server)
* expects to find in environmental variables DATASET_URL and DATASET_API_KEY informations needed to call AddEvents API
  * if DATASET_URL is not specified, then dummy server is used
  * if option --server is used then it overwrites DATASET_URL
* collector is configured to read test files and use datasetexporter
* once all the log lines are processed then:
  * JSON describing statistics is appended at the end of benchmark.log
  * if everything was processed then exit code is 0, otherwise it's 1


## Results

During execution and in the final statistics several metrics are produced. Following convention is used:
* unit is usually suffix of the name - MBpS (MB per second)
* input - measures size or speed in terms of input files
* output - measures size of JSON payloads before compression

Example run is [here](https://github.com/scalyr/opentelemetry-exporter-dataset/actions/runs/5550337166).
Latest benchmarks are available in the [benchmarks.yaml](https://github.com/scalyr/opentelemetry-exporter-dataset/actions/workflows/benchmarks.yaml) action.


### Values

* `Time`: The timestamp when the benchmark report was generated.
* `Server url`: The URL of the DataSet server that was used.
* `Server api key`: The last four characters of the API key used.
* `Num files`: The number of input files used for benchmarking.
* `Num lines`: The desired number of lines in each input file.
* `Lines mu`: The mean length of a line, following a normal distribution N(mu, sigma).
* `Lines sigma`: The standard deviation of line lengths, following a normal distribution N(mu, sigma).
* `Seed`: The seed used for the random generator to generate lines.
* `Processing time s`: The duration, in seconds, for which the exporter processed the data, from the moment it received the first data to when the last data were accepted by the server.
* `Input size b`: The size of the generated input log files.
* `Events expected`: The total number of events that should be processed (`Num files` * `Num lines`).
* `Events processed`: The number of events that have been added to the buffers. This number should match the `Events expected` value at the end.
* `Events enqueued`: The number of events that have been accepted by the library.
* `Events waiting`: The number of events being processed by the library (`Events enqueued` - `Events processed`).
* `Buffers processed`: The number of buffers that have been accepted by the server.
* `Buffers enqueued`: The number of buffers that have been created and submitted for further processing.
* `Buffers waiting`: The number of buffers being processed (`Buffers enqueued` - `Buffers processed`).
* `Buffers dropped`: The number of buffers that were dropped after all retries have failed.
* `Bytes accepted mb`: The amount of accepted raw data in megabytes (MB) by the server without compression. If the payload is accepted after a retry, its size is counted once.
* `Bytes sent mb`: The amount of raw data in megabytes (MB) sent to the server without compression. If the payload is accepted after a retry, its size is counted twice.
* `Bytes per buffer mb`: The average size of the buffer.
* `Throughput out m bp s`: The estimated output throughput in megabytes per second (MB/s), measured based on the output data size (`Bytes accepted mb` / `Processing time s`).
* `Success rate`: The ratio of successful API calls to all API calls.
* `Throughput in m bp s`: The estimated input throughput in megabytes per second (MB/s), measured based on the input data size (based on `Processed (MB)`, `Duration (s)`).
* `Throughput in eventsp s`: The estimated input throughput in events per second (events/s), measured based on the input data size (based on `Processed (events)`, `Duration (s)`).
* `Start time`: The timestamp when the benchmark started.
* `End time`: The timestamp when the benchmark finished.
* `This branch`: The name of the branch in this repository.
* `This hash`: The hash of the current commit in this repository.
* `Lib branch`: The name of the branch in the `dataset-go` repository used to build the collector.
* `Lib hash`: The hash of the current commit in the `dataset-go` repository used to build the collector.
* `Otel branch`: The name of the branch in the `opentelemetry-collector-contrib` repository used to build the collector.
* `Otel hash`: The hash of the current commit in the `opentelemetry-collector-contrib` repository used to build the collector.
* `Docker compose hash`: The MD5 checksum of the `docker-compose.yaml` file.
* `Otel config hash`: The MD5 checksum of the `otel-config.yaml` file.
* `Script hash`: The MD5 checksum of the `test.sh` file.
* `Uname`: The username of the user who executed the benchmark.
* `Hostname`: The hostname of the machine where the benchmark was executed.
* `IP info`: The IP address of the machine where the benchmark was executed.
* `CPU info`: Information about the CPU of the machine.
* `Container info`: Information about the Docker container used for the benchmark.
* `Last`: A dummy attribute included to simplify JSON generation or for any other necessary purposes.
