// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package migrate // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/schemaprocessor/internal/migrate"

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/otel/schema/v1.0/ast"
	"go.uber.org/multierr"
)

// ValueMatch defines the expected match type
// that is used on creation of `ConditionalAttributeSet`
type ValueMatch interface {
	~string
}

type ConditionalAttributeSet struct {
	on    *map[string]struct{}
	attrs *AttributeChangeSet
}

type ConditionalAttributeSetSlice []*ConditionalAttributeSet

func NewConditionalAttributeSet[Match ValueMatch](mappings ast.AttributeMap, matches ...Match) *ConditionalAttributeSet {
	on := make(map[string]struct{})
	for _, m := range matches {
		on[string(m)] = struct{}{}
	}
	return &ConditionalAttributeSet{
		on:    &on,
		attrs: NewAttributeChangeSet(mappings),
	}
}

func (ca *ConditionalAttributeSet) Apply(attrs pcommon.Map, values ...string) (errs error) {
	if ca.check(values...) {
		errs = ca.attrs.Apply(attrs)
	}
	return errs
}

func (ca *ConditionalAttributeSet) Rollback(attrs pcommon.Map, values ...string) (errs error) {
	if ca.check(values...) {
		errs = ca.attrs.Rollback(attrs)
	}
	return errs
}

func (ca *ConditionalAttributeSet) check(values ...string) bool {
	if len(*ca.on) == 0 {
		return true
	}
	for _, v := range values {
		if _, ok := (*ca.on)[v]; !ok {
			return false
		}
	}
	return true
}

func NewConditionalAttributeSetSlice(conditions ...*ConditionalAttributeSet) *ConditionalAttributeSetSlice {
	values := new(ConditionalAttributeSetSlice)
	for _, c := range conditions {
		(*values) = append((*values), c)
	}
	return values
}

func (slice *ConditionalAttributeSetSlice) Apply(attrs pcommon.Map, values ...string) error {
	return slice.do(StateSelectorApply, attrs, values)
}

func (slice *ConditionalAttributeSetSlice) Rollback(attrs pcommon.Map, values ...string) error {
	return slice.do(StateSelectorRollback, attrs, values)
}

func (slice *ConditionalAttributeSetSlice) do(ss StateSelector, attrs pcommon.Map, values []string) (errs error) {
	for i := 0; i < len((*slice)); i++ {
		switch ss {
		case StateSelectorApply:
			errs = multierr.Append(errs, (*slice)[i].Apply(attrs, values...))
		case StateSelectorRollback:
			errs = multierr.Append(errs, (*slice)[len((*slice))-i-1].Rollback(attrs, values...))
		}
	}
	return errs
}
