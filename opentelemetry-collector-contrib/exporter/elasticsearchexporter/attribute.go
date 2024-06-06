// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package elasticsearchexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/elasticsearchexporter"

import "go.opentelemetry.io/collector/pdata/pcommon"

// dynamic index attribute key constants
const (
	indexPrefix = "elasticsearch.index.prefix"
	indexSuffix = "elasticsearch.index.suffix"
)

// resource is higher priotized than record attribute
type attrGetter interface {
	Attributes() pcommon.Map
}

// retrieve attribute out of resource, scope, and record (span or log, if not found in resource)
func getFromAttributes(name string, resource, scope, record attrGetter) string {
	var str string
	val, exist := resource.Attributes().Get(name)
	if !exist {
		val, exist = scope.Attributes().Get(name)
		if !exist {
			val, exist = record.Attributes().Get(name)
			if exist {
				str = val.AsString()
			}
		}
		if exist {
			str = val.AsString()
		}
	}
	if exist {
		str = val.AsString()
	}
	return str
}
