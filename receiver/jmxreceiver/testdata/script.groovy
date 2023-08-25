/*
 * Copyright The OpenTelemetry Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import io.opentelemetry.api.metrics.common.Labels

def loadMatches = otel.queryJmx("org.apache.cassandra.metrics:type=Storage,name=Load")
if (loadMatches.size() > 0) {
    def load = loadMatches.first()
    otel.longUpDownCounter(
        "cassandra.storage.load",
        "Size, in bytes, of the on disk data size this node manages",
        "By"
    ).add(load.Count, Labels.of(
        "myKey", "myVal",
        System.properties["my.label.name"], System.properties["my.label.value"],
        System.properties["my.other.label.name"], System.properties["my.other.label.value"],
    ))
}
