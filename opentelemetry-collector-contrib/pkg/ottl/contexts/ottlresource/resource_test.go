// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ottlresource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/ottltest"
)

func Test_newPathGetSetter(t *testing.T) {
	refResource := createTelemetry()

	newAttrs := pcommon.NewMap()
	newAttrs.PutStr("hello", "world")

	newCache := pcommon.NewMap()
	newCache.PutStr("temp", "value")

	newPMap := pcommon.NewMap()
	pMap2 := newPMap.PutEmptyMap("k2")
	pMap2.PutStr("k1", "string")

	newMap := make(map[string]interface{})
	newMap2 := make(map[string]interface{})
	newMap2["k1"] = "string"
	newMap["k2"] = newMap2

	tests := []struct {
		name     string
		path     []ottl.Field
		orig     interface{}
		newVal   interface{}
		modified func(resource pcommon.Resource, cache pcommon.Map)
	}{
		{
			name: "cache",
			path: []ottl.Field{
				{
					Name: "cache",
				},
			},
			orig:   pcommon.NewMap(),
			newVal: newCache,
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				newCache.CopyTo(cache)
			},
		},
		{
			name: "cache access",
			path: []ottl.Field{
				{
					Name: "cache",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("temp"),
						},
					},
				},
			},
			orig:   nil,
			newVal: "new value",
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				cache.PutStr("temp", "new value")
			},
		},
		{
			name: "attributes",
			path: []ottl.Field{
				{
					Name: "attributes",
				},
			},
			orig:   refResource.Attributes(),
			newVal: newAttrs,
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				newAttrs.CopyTo(resource.Attributes())
			},
		},
		{
			name: "attributes string",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("str"),
						},
					},
				},
			},
			orig:   "val",
			newVal: "newVal",
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				resource.Attributes().PutStr("str", "newVal")
			},
		},
		{
			name: "attributes bool",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("bool"),
						},
					},
				},
			},
			orig:   true,
			newVal: false,
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				resource.Attributes().PutBool("bool", false)
			},
		},
		{
			name: "attributes int",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("int"),
						},
					},
				},
			},
			orig:   int64(10),
			newVal: int64(20),
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				resource.Attributes().PutInt("int", 20)
			},
		},
		{
			name: "attributes float",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("double"),
						},
					},
				},
			},
			orig:   float64(1.2),
			newVal: float64(2.4),
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				resource.Attributes().PutDouble("double", 2.4)
			},
		},
		{
			name: "attributes bytes",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("bytes"),
						},
					},
				},
			},
			orig:   []byte{1, 3, 2},
			newVal: []byte{2, 3, 4},
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				resource.Attributes().PutEmptyBytes("bytes").FromRaw([]byte{2, 3, 4})
			},
		},
		{
			name: "attributes array string",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("arr_str"),
						},
					},
				},
			},
			orig: func() pcommon.Slice {
				val, _ := refResource.Attributes().Get("arr_str")
				return val.Slice()
			}(),
			newVal: []string{"new"},
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				resource.Attributes().PutEmptySlice("arr_str").AppendEmpty().SetStr("new")
			},
		},
		{
			name: "attributes array bool",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("arr_bool"),
						},
					},
				},
			},
			orig: func() pcommon.Slice {
				val, _ := refResource.Attributes().Get("arr_bool")
				return val.Slice()
			}(),
			newVal: []bool{false},
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				resource.Attributes().PutEmptySlice("arr_bool").AppendEmpty().SetBool(false)
			},
		},
		{
			name: "attributes array int",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("arr_int"),
						},
					},
				},
			},
			orig: func() pcommon.Slice {
				val, _ := refResource.Attributes().Get("arr_int")
				return val.Slice()
			}(),
			newVal: []int64{20},
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				resource.Attributes().PutEmptySlice("arr_int").AppendEmpty().SetInt(20)
			},
		},
		{
			name: "attributes array float",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("arr_float"),
						},
					},
				},
			},
			orig: func() pcommon.Slice {
				val, _ := refResource.Attributes().Get("arr_float")
				return val.Slice()
			}(),
			newVal: []float64{2.0},
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				resource.Attributes().PutEmptySlice("arr_float").AppendEmpty().SetDouble(2.0)
			},
		},
		{
			name: "attributes array bytes",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("arr_bytes"),
						},
					},
				},
			},
			orig: func() pcommon.Slice {
				val, _ := refResource.Attributes().Get("arr_bytes")
				return val.Slice()
			}(),
			newVal: [][]byte{{9, 6, 4}},
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				resource.Attributes().PutEmptySlice("arr_bytes").AppendEmpty().SetEmptyBytes().FromRaw([]byte{9, 6, 4})
			},
		},
		{
			name: "attributes pcommon.Map",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("pMap"),
						},
					},
				},
			},
			orig: func() pcommon.Map {
				val, _ := refResource.Attributes().Get("pMap")
				return val.Map()
			}(),
			newVal: newPMap,
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				m := resource.Attributes().PutEmptyMap("pMap")
				m2 := m.PutEmptyMap("k2")
				m2.PutStr("k1", "string")
			},
		},
		{
			name: "attributes mpa[string]interface",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("map"),
						},
					},
				},
			},
			orig: func() pcommon.Map {
				val, _ := refResource.Attributes().Get("map")
				return val.Map()
			}(),
			newVal: newMap,
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				m := resource.Attributes().PutEmptyMap("map")
				m2 := m.PutEmptyMap("k2")
				m2.PutStr("k1", "string")
			},
		},
		{
			name: "attributes nested",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("slice"),
						},
						{
							Int: ottltest.Intp(0),
						},
						{
							String: ottltest.Strp("map"),
						},
					},
				},
			},
			orig: func() string {
				val, _ := refResource.Attributes().Get("slice")
				val, _ = val.Slice().At(0).Map().Get("map")
				return val.Str()
			}(),
			newVal: "new",
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				resource.Attributes().PutEmptySlice("slice").AppendEmpty().SetEmptyMap().PutStr("map", "new")
			},
		},
		{
			name: "attributes nested new values",
			path: []ottl.Field{
				{
					Name: "attributes",
					Keys: []ottl.Key{
						{
							String: ottltest.Strp("new"),
						},
						{
							Int: ottltest.Intp(2),
						},
						{
							Int: ottltest.Intp(0),
						},
					},
				},
			},
			orig: func() interface{} {
				return nil
			}(),
			newVal: "new",
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				s := resource.Attributes().PutEmptySlice("new")
				s.AppendEmpty()
				s.AppendEmpty()
				s.AppendEmpty().SetEmptySlice().AppendEmpty().SetStr("new")
			},
		},
		{
			name: "dropped_attributes_count",
			path: []ottl.Field{
				{
					Name: "dropped_attributes_count",
				},
			},
			orig:   int64(10),
			newVal: int64(20),
			modified: func(resource pcommon.Resource, cache pcommon.Map) {
				resource.SetDroppedAttributesCount(20)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor, err := newPathGetSetter(tt.path)
			assert.NoError(t, err)

			resource := createTelemetry()

			tCtx := NewTransformContext(resource)
			got, err := accessor.Get(context.Background(), tCtx)
			assert.Nil(t, err)
			assert.Equal(t, tt.orig, got)

			err = accessor.Set(context.Background(), tCtx, tt.newVal)
			assert.Nil(t, err)

			exRes := createTelemetry()
			exCache := pcommon.NewMap()
			tt.modified(exRes, exCache)

			assert.Equal(t, exRes, resource)
			assert.Equal(t, exCache, tCtx.getCache())
		})
	}
}

func createTelemetry() pcommon.Resource {
	resource := pcommon.NewResource()

	resource.Attributes().PutStr("str", "val")
	resource.Attributes().PutBool("bool", true)
	resource.Attributes().PutInt("int", 10)
	resource.Attributes().PutDouble("double", 1.2)
	resource.Attributes().PutEmptyBytes("bytes").FromRaw([]byte{1, 3, 2})

	arrStr := resource.Attributes().PutEmptySlice("arr_str")
	arrStr.AppendEmpty().SetStr("one")
	arrStr.AppendEmpty().SetStr("two")

	arrBool := resource.Attributes().PutEmptySlice("arr_bool")
	arrBool.AppendEmpty().SetBool(true)
	arrBool.AppendEmpty().SetBool(false)

	arrInt := resource.Attributes().PutEmptySlice("arr_int")
	arrInt.AppendEmpty().SetInt(2)
	arrInt.AppendEmpty().SetInt(3)

	arrFloat := resource.Attributes().PutEmptySlice("arr_float")
	arrFloat.AppendEmpty().SetDouble(1.0)
	arrFloat.AppendEmpty().SetDouble(2.0)

	arrBytes := resource.Attributes().PutEmptySlice("arr_bytes")
	arrBytes.AppendEmpty().SetEmptyBytes().FromRaw([]byte{1, 2, 3})
	arrBytes.AppendEmpty().SetEmptyBytes().FromRaw([]byte{2, 3, 4})

	pMap := resource.Attributes().PutEmptyMap("pMap")
	pMap.PutStr("original", "map")

	m := resource.Attributes().PutEmptyMap("map")
	m.PutStr("original", "map")

	s := resource.Attributes().PutEmptySlice("slice")
	s.AppendEmpty().SetEmptyMap().PutStr("map", "pass")

	resource.SetDroppedAttributesCount(10)

	return resource
}
