# Release Notes
## 0.18.0

* break: drop support for Go 1.20 and add testing for Go 1.22
* fix: release resources associated with inactive sessions

## 0.17.0

* break: export metrics as attributes and as snake_case

## 0.16.0

* feat: introduce open telemetry metrics to increase observability

## 0.15.0

* break: move grouped attributes to session info


## 0.14.0

* break: drop support for Go 1.19 and add testing for Go 1.21

## 0.13.0

* fix: do not drop event when buffer is being published
* fix: do not double count dropped buffers

## 0.12.0

* fix: make client shutdown timeout configurable.

## 0.0.11

* feat: make client shutdown timeout configurable.

## 0.0.10

* feat: make `serverHost` explicit field of the `Event`.

## 0.0.9 Add more retry parameters

* fix: recovery from error mode - when DataSet HTTP req sending fails consuming of additional input events is blocked. Fix issue with rejected input events once retry attempts times out.
* feat: change the content of User-Agent header in order to send more details. Also allow library consumer to define additional details.
* Refactoring and documentation

## 0.0.8 Improve logging

* Log amount of transferred bytes to DataSet

## 0.0.7 Add more retry parameters

* To make OpenTelemetry configuration more stable we have introduced more retry options. We have to propagate them to the library.
* Fix another data race caused by `client.worker`.

## 0.0.6 Fix Concurrency Issues

* OpenTelemetry can call AddEvents multiple times in parallel. Let's use another Pub/Sub to publish events into topic and then consume them independently.

## 0.0.5 Quick Hack

* OpenTelemetry can call AddEvents multiple times in parallel. Add lock so only one of them is in progress in any given time.

## 0.0.4 Fix Concurrency Issues

* sometimes not all events have been delivered exactly once

## 0.0.3 Fix Data Races

* fixed [data races](https://go.dev/doc/articles/race_detector)

## 0.0.2 Fix Data Races

* fixed [data races](https://go.dev/doc/articles/race_detector)
* added function `client.Finish` to allow waiting on processing all events
## 0.0.1 Initial Release

* support for API call [addEvents](https://app.scalyr.com/help/api#addEvents)
