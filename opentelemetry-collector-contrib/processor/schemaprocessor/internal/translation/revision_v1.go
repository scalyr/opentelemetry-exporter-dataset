// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package translation // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/schemaprocessor/internal/translation"

import (
	"go.opentelemetry.io/otel/schema/v1.0/ast"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/schemaprocessor/internal/migrate"
)

// RevisionV1 represents all changes that are to be
// applied to a signal at a given version.
type RevisionV1 struct {
	ver              *Version
	all              *migrate.AttributeChangeSetSlice
	resource         *migrate.AttributeChangeSetSlice
	spans            *migrate.ConditionalAttributeSetSlice
	eventNames       *migrate.SignalNameChangeSlice
	eventAttrsOnSpan *migrate.ConditionalAttributeSetSlice
	eventAttrsOnName *migrate.ConditionalAttributeSetSlice
	metricsAttrs     *migrate.ConditionalAttributeSetSlice
	metricNames      *migrate.SignalNameChangeSlice
}

// NewRevision processes the VersionDef and assigns the version to this revision
// to allow sorting within a slice.
// Since VersionDef uses custom types for various definitions, it isn't possible
// to cast those values into the primitives so each has to be processed together.
// Generics would be handy here.
func NewRevision(ver *Version, def ast.VersionDef) *RevisionV1 {
	return &RevisionV1{
		ver:              ver,
		all:              newAttributeChangeSetSliceFromChanges(def.All),
		resource:         newAttributeChangeSetSliceFromChanges(def.Resources),
		spans:            newSpanConditionalAttributeSlice(def.Spans),
		eventNames:       newSpanEventSignalSlice(def.SpanEvents),
		eventAttrsOnSpan: newSpanEventConditionalSpans(def.SpanEvents),
		eventAttrsOnName: newSpanEventConditionalNames(def.SpanEvents),
		metricsAttrs:     newMetricConditionalSlice(def.Metrics),
		metricNames:      newMetricNameSignalSlice(def.Metrics),
	}
}

func newAttributeChangeSetSliceFromChanges(attrs ast.Attributes) *migrate.AttributeChangeSetSlice {
	values := make([]*migrate.AttributeChangeSet, 0, 10)
	for _, at := range attrs.Changes {
		if renamed := at.RenameAttributes; renamed != nil {
			values = append(values, migrate.NewAttributeChangeSet(renamed.AttributeMap))
		}
	}
	return migrate.NewAttributeChangeSetSlice(values...)
}

func newSpanConditionalAttributeSlice(spans ast.Spans) *migrate.ConditionalAttributeSetSlice {
	values := make([]*migrate.ConditionalAttributeSet, 0, 10)
	for _, ch := range spans.Changes {
		if renamed := ch.RenameAttributes; renamed != nil {
			values = append(values, migrate.NewConditionalAttributeSet(
				renamed.AttributeMap,
				renamed.ApplyToSpans...,
			))
		}
	}
	return migrate.NewConditionalAttributeSetSlice(values...)
}

func newSpanEventSignalSlice(events ast.SpanEvents) *migrate.SignalNameChangeSlice {
	values := make([]*migrate.SignalNameChange, 0, 10)
	for _, ch := range events.Changes {
		if renamed := ch.RenameEvents; renamed != nil {
			values = append(values, migrate.NewSignalNameChange(renamed.EventNameMap))
		}
	}
	return migrate.NewSignalNameChangeSlice(values...)
}

func newSpanEventConditionalSpans(events ast.SpanEvents) *migrate.ConditionalAttributeSetSlice {
	values := make([]*migrate.ConditionalAttributeSet, 0, 10)
	for _, ch := range events.Changes {
		if rename := ch.RenameAttributes; rename != nil {
			values = append(values, migrate.NewConditionalAttributeSet(rename.AttributeMap, rename.ApplyToSpans...))
		}
	}
	return migrate.NewConditionalAttributeSetSlice(values...)
}

func newSpanEventConditionalNames(events ast.SpanEvents) *migrate.ConditionalAttributeSetSlice {
	values := make([]*migrate.ConditionalAttributeSet, 0, 10)
	for _, ch := range events.Changes {
		if rename := ch.RenameAttributes; rename != nil {
			values = append(values, migrate.NewConditionalAttributeSet(rename.AttributeMap, rename.ApplyToEvents...))
		}
	}
	return migrate.NewConditionalAttributeSetSlice(values...)
}

func newMetricConditionalSlice(metrics ast.Metrics) *migrate.ConditionalAttributeSetSlice {
	values := make([]*migrate.ConditionalAttributeSet, 0, 10)
	for _, ch := range metrics.Changes {
		if rename := ch.RenameAttributes; rename != nil {
			values = append(values, migrate.NewConditionalAttributeSet(rename.AttributeMap, rename.ApplyToMetrics...))
		}
	}
	return migrate.NewConditionalAttributeSetSlice(values...)
}

func newMetricNameSignalSlice(metrics ast.Metrics) *migrate.SignalNameChangeSlice {
	values := make([]*migrate.SignalNameChange, 0, 10)
	for _, ch := range metrics.Changes {
		values = append(values, migrate.NewSignalNameChange(ch.RenameMetrics))
	}
	return migrate.NewSignalNameChangeSlice(values...)
}
