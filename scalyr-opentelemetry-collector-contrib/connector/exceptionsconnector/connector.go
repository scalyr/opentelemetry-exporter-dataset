// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exceptionsconnector // import "github.com/open-telemetry/opentelemetry-collector-contrib/connector/exceptionsconnector"

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	conventions "go.opentelemetry.io/collector/semconv/v1.18.0"
)

const (
	serviceNameKey         = conventions.AttributeServiceName
	exceptionTypeKey       = conventions.AttributeExceptionType
	exceptionMessageKey    = conventions.AttributeExceptionMessage
	exceptionStacktraceKey = conventions.AttributeExceptionStacktrace
	// TODO(marctc): formalize these constants in the OpenTelemetry specification.
	spanKindKey   = "span.kind"   // OpenTelemetry non-standard constant.
	statusCodeKey = "status.code" // OpenTelemetry non-standard constant.
	eventNameExc  = "exception"   // OpenTelemetry non-standard constant.
)

type dimension struct {
	name  string
	value *pcommon.Value
}

func newDimensions(cfgDims []Dimension) []dimension {
	if len(cfgDims) == 0 {
		return nil
	}
	dims := make([]dimension, len(cfgDims))
	for i := range cfgDims {
		dims[i].name = cfgDims[i].Name
		if cfgDims[i].Default != nil {
			val := pcommon.NewValueStr(*cfgDims[i].Default)
			dims[i].value = &val
		}
	}
	return dims
}

// getDimensionValue gets the dimension value for the given configured dimension.
// It searches through the span's attributes first, being the more specific;
// falling back to searching in resource attributes if it can't be found in the span.
// Finally, falls back to the configured default value if provided.
//
// The ok flag indicates if a dimension value was fetched in order to differentiate
// an empty string value from a state where no value was found.
func getDimensionValue(d dimension, spanAttrs pcommon.Map, eventAttrs pcommon.Map) (v pcommon.Value, ok bool) {
	// The more specific span attribute should take precedence.
	if attr, exists := spanAttrs.Get(d.name); exists {
		return attr, true
	}
	if attr, exists := eventAttrs.Get(d.name); exists {
		return attr, true
	}
	// Set the default if configured, otherwise this metric will have no value set for the dimension.
	if d.value != nil {
		return *d.value, true
	}
	return v, ok
}
