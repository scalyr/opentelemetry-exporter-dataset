// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ottlfuncs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
)

func Test_substring(t *testing.T) {
	tests := []struct {
		name     string
		target   ottl.StringGetter[interface{}]
		start    int64
		length   int64
		expected interface{}
	}{
		{
			name: "substring",
			target: &ottl.StandardStringGetter[interface{}]{
				Getter: func(ctx context.Context, tCtx interface{}) (interface{}, error) {
					return "123456789", nil
				},
			},
			start:    3,
			length:   3,
			expected: "456",
		},
		{
			name: "substring with result of total string",
			target: &ottl.StandardStringGetter[interface{}]{
				Getter: func(ctx context.Context, tCtx interface{}) (interface{}, error) {
					return "123456789", nil
				},
			},
			start:    0,
			length:   9,
			expected: "123456789",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprFunc, err := substring(tt.target, tt.start, tt.length)
			assert.NoError(t, err)
			result, err := exprFunc(nil, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_substring_validation(t *testing.T) {
	tests := []struct {
		name   string
		target ottl.StringGetter[interface{}]
		start  int64
		length int64
	}{
		{
			name: "substring with result of empty string",
			target: &ottl.StandardStringGetter[interface{}]{
				Getter: func(ctx context.Context, tCtx interface{}) (interface{}, error) {
					return "123456789", nil
				},
			},
			start:  0,
			length: 0,
		},
		{
			name: "substring with invalid start index",
			target: &ottl.StandardStringGetter[interface{}]{
				Getter: func(ctx context.Context, tCtx interface{}) (interface{}, error) {
					return "123456789", nil
				},
			},
			start:  -1,
			length: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := substring(tt.target, tt.start, tt.length)
			assert.Error(t, err)
		})
	}
}

func Test_substring_error(t *testing.T) {
	tests := []struct {
		name   string
		target ottl.StringGetter[interface{}]
		start  int64
		length int64
	}{
		{
			name: "substring empty string",
			target: &ottl.StandardStringGetter[interface{}]{
				Getter: func(ctx context.Context, tCtx interface{}) (interface{}, error) {
					return "", nil
				},
			},
			start:  3,
			length: 6,
		},
		{
			name: "substring with invalid length index",
			target: &ottl.StandardStringGetter[interface{}]{
				Getter: func(ctx context.Context, tCtx interface{}) (interface{}, error) {
					return "123456789", nil
				},
			},
			start:  3,
			length: 20,
		},
		{
			name: "substring non-string",
			target: &ottl.StandardStringGetter[interface{}]{
				Getter: func(ctx context.Context, tCtx interface{}) (interface{}, error) {
					return 123456789, nil
				},
			},
			start:  3,
			length: 6,
		},
		{
			name: "substring nil string",
			target: &ottl.StandardStringGetter[interface{}]{
				Getter: func(ctx context.Context, tCtx interface{}) (interface{}, error) {
					return nil, nil
				},
			},
			start:  3,
			length: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprFunc, err := substring(tt.target, tt.start, tt.length)
			assert.NoError(t, err)
			result, err := exprFunc(nil, nil)
			assert.Error(t, err)
			assert.Equal(t, nil, result)
		})
	}
}
