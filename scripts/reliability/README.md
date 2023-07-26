# Reliability Test

This is reliability test, that:
* executes two images:
  * open telemetry collector
  * [dummy API server](../../dummy-server)
* uses flog to generate data - it generates 10k lines in 100s
* after 5s it sets the server HTTP status 530 for 100s
* then it changes the server to return HTTP status again
* then it wait at most 50s to get all the events
* and then it checks whether number of processed events matched number of created events

## Usage

* Build image, if you have made some changes:
  * ```bash
    (cd ../..; make docker-build)
    ```
* Run tests:
  * ```bash
    ./test.sh
    ```
