# Scripts

This folder contains various scripts for testing different scenarios. Number in the bracket describes how mature that script is.

## Overview

* `benchmarking` - measures throughput by continuously generating logs and sending them via OTLP. Requires manual effort to find out the real numbers (0/5).
* `e2e` - measures maximum throughput. Log files are generator in advance, and it measures time until all the lines are accepted by the server (3/5).
* `reliability` - tests whether all logs are processed even when there is server outage. Logs are generated continuously, server is for some time period returning 530, and then starts accepting data again (2/5).
