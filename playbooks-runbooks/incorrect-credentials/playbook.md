---
Title: "Incorrect Credentials"
Documentation:
  - https://github.com/scalyr/dataset-go/blob/main/docs/ARCHITECTURE.md
  - https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/datasetexporter/README.md
  - https://github.com/scalyr/opentelemetry-exporter-dataset/blob/main/docs/INSTRUCTIONS.md
Github: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/datasetexporter
Revision Date: 2023-07-14
---

# Incorrect Credentials

## Issue Description

You have encountered errors while starting the OpenTelemetry collector or processing data, which can be one of the following:

1. Empty credentials:
   ```
   Error: invalid configuration: exporters::dataset: api_key is required
   2023/07/14 12:56:41 collector server run finished with error: invalid configuration: exporters::dataset: api_key is required
   ```
   * This error suggests that the configuration for the dataset exporter is invalid because the `api_key` is either empty or missing. You need to ensure that the api_key is properly filled in the configuration file.
2. Incorrect API key - malformed token:
   ```
   2023-07-14T13:26:50.253Z        warn    client/add_events.go:381        !!!!! PAYLOAD WAS NOT ACCEPTED !!!!     {"kind": "exporter", "data_type": "logs", "name": "dataset/logs", "statusCode": 401}
   2023-07-14T13:26:50.262Z        error   client/add_events.go:332        Problematic payload     {"kind": "exporter", "data_type": "logs", "name": "dataset/logs", "message": "Couldn't decode API token ...AAA.", "status": "error/client/badParam", "payloadLength": 1692}
   github.com/scalyr/dataset-go/pkg/client.(*DataSetClient).sendAddEventsBuffer
   ```
   * This error (based on `statusCode`, `message`, and `status`) suggests that the used `api_key` has invalid structure.
3. Incorrect API key - Disabled account:
   ```
   2023-07-14T13:28:58.079Z        warn    client/add_events.go:381        !!!!! PAYLOAD WAS NOT ACCEPTED !!!!     {"kind": "exporter", "data_type": "logs", "name": "dataset/logs", "statusCode": 401}
   2023-07-14T13:28:58.081Z        error   client/add_events.go:332        Problematic payload     {"kind": "exporter", "data_type": "logs", "name": "dataset/logs", "message": "this account has been disabled", "status": "error/client/noPermission/accountDisabled", "payloadLength": 1733}
   github.com/scalyr/dataset-go/pkg/client.(*DataSetClient).sendAddEventsBuffer
   ```
   * This error (based on `statusCode`, `message`, and `status`) suggests that the used `api_key` is associated with disabled account.
4. Incorrect API key - Wrong API Key type:
   ```
   2023-07-14T13:43:13.900Z        warn    client/add_events.go:381        !!!!! PAYLOAD WAS NOT ACCEPTED !!!!     {"kind": "exporter", "data_type": "logs", "name": "dataset/logs", "statusCode": 401}
   2023-07-14T13:43:13.900Z        error   client/add_events.go:332        Problematic payload     {"kind": "exporter", "data_type": "logs", "name": "dataset/logs", "message": "authorization token [...1RJA-] does not grant LOG_WRITE permission", "status": "error/client/noPermission", "payloadLength": 1733}
   github.com/scalyr/dataset-go/pkg/client.(*DataSetClient).sendAddEventsBuffer
   ```
   * This error (based on `statusCode`, `message`, and `status`) suggests that the used `api_key` is not a key with `Log Write` permission.

## Procedural Steps to Restore the Service

### Empty Credentials

Follow these steps to troubleshoot and restore the service:

1. Check the configuration file to ensure that the `api_key` for the `datasetexporter` is provided and not empty.
2. If you are using an environmental variable to fetch the api_key (e.g., `${env:DATASET_API_KEY}`), make sure the environmental variable is set and contains a valid value.

### Incorrect API key

Follow these steps to troubleshoot and restore the service:

1. Review the error messages, specifically the `statusCode`, `message`, and `status`, to understand the cause of the problem.
2. Consult the official [Scalyr documentation](https://app.scalyr.com/help/api#addEvents) for information on troubleshooting the specific error messages.
3. Verify that you are using the correct [API key](https://app.scalyr.com/help/api-keys) with write permissions (Log Access Keys).
4. Check for any other issues mentioned in the error message, such as disabled accounts or insufficient permissions.


## Expected Outcome

Once the configuration is correctly set with the required API key and any payload-related issues are resolved, the OpenTelemetry collector should start successfully. You should be able to see the incoming data in the DataSet server specified by the `dataset.dataset_url` configuration option.

## Verification Process Step(s)

To verify if the issue is resolved:

1. Log in to the DataSet server specified in the `dataset_url` configuration option.
2. Check if you can see the incoming data in the DataSet server.

## Miscellanea

Ensure you consult the relevant documentation for additional troubleshooting steps and guidance.
