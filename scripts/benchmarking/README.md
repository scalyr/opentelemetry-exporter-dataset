# Benchmarking Test

This is benchmark that:
* uses [telemetrygen](https://pkg.go.dev/github.com/open-telemetry/opentelemetry-collector-contrib/cmd/telemetrygen) to generate logs
  * you should modify [docker-compose.yaml](docker-compose.yaml) section of `telemetrygen-logs-dataset` and find values that ends up with 0 waiting buffers
* you should modify [docker-compose.yaml](docker-compose.yaml) section `telemetrygen-logs-dataset` to find parameters where all buffers are being processed and nothing is piling up.

## Usage

* Build image, if you have made some changes:
  * ```bash
    (cd ../..; make docker-build)
    ```
* Set environmental variables:
  * Real:
      ```bash
      export DATASET_API_KEY=your-api-key;
      export DATASET_URL='https://app-qatesting.scalyr.com';
      ```
  * [Dummy API server](../../dummy-server):
      ```bash
      export DATASET_API_KEY=dummy;
      export DATASET_URL='http://dataset-dummy-server-bench:8000';
      ```
* Run test:
  * ```bash
    ./test.sh
    ```
