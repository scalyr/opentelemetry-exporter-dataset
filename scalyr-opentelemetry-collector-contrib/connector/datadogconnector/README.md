# Datadog Connector

## Description

The Datadog Connector is a connector component that computes Datadog APM Stats pre-sampling in the event that your traces pipeline is sampled using components such as the tailsamplingprocessor or probabilisticsamplerprocessor.

The connector is most applicable when using the sampling components such as the [tailsamplingprocessor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/tailsamplingprocessor#tail-sampling-processor), or the [probabilisticsamplerprocessor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/probabilisticsamplerprocessor) in one of your pipelines. The sampled pipeline should be duplicated and the `datadog` connector should be added to the the pipeline that is not being sampled to ensure that Datadog APM Stats are accurate in the backend.

## Usage

To use the Datadog Connector, add the connector to one set of the duplicated pipelines while sampling the other. The Datadog Connector will compute APM Stats on all spans that it sees. Here is an example on how to add it to a pipeline using the [probabilisticsampler]:

<table>
<tr>
<td> Before </td> <td> After </td>
</tr>
<tr>
<td valign="top">

```yaml
# ...
processors:
  # ...
  probabilistic_sampler:
    sampling_percentage: 20
  # add the "datadog" processor definition
  datadog:

exporters:
  datadog:
    api:
      key: ${env:DD_API_KEY}

service:
  pipelines:
    traces:
      receivers: [otlp]
      # prepend it to the sampler in your pipeline:
      processors: [batch, datadog, probabilistic_sampler]
      exporters: [datadog]

    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [datadog]
```

</td><td valign="top">

```yaml
# ...
processors:
  probabilistic_sampler:
    sampling_percentage: 20

connectors:
    # add the "datadog" connector definition and further configurations
    datadog/connector:

exporters:
  datadog:
    api:
      key: ${env:DD_API_KEY}

service:
  pipelines:
   traces:
     receivers: [otlp]
     processors: [batch]
     exporters: [datadog/connector]

   traces/2: # this pipeline uses sampling
     receivers: [otlp]
     processors: [batch, probabilistic_sampler]
     exporters: [datadog]

  metrics:
    receivers: [datadog/connector]
    processors: [batch]
    exporters: [datadog]
```
</tr></table>

Here we have two traces pipelines that ingest the same data but one is being sampled. The one that is sampled has its data sent to the datadog backend for you to see the sampled subset of the total traces sent across. The other non-sampled pipeline of traces sends its data to the metrics pipeline to be used in the APM stats. This unsampled pipeline gives the full picture of how much data the application emits in traces. 

