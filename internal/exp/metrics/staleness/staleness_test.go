// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package staleness

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/exp/metrics/identity"
)

func TestStaleness(t *testing.T) {
	max := 1 * time.Second
	stalenessMap := NewStaleness[int](
		max,
		&RawMap[identity.Stream, int]{},
	)

	idA := generateStreamID(t, map[string]any{
		"aaa": "123",
	})
	idB := generateStreamID(t, map[string]any{
		"bbb": "456",
	})
	idC := generateStreamID(t, map[string]any{
		"ccc": "789",
	})
	idD := generateStreamID(t, map[string]any{
		"ddd": "024",
	})

	initialTime := time.Time{}
	timeA := initialTime.Add(2 * time.Second)
	timeB := initialTime.Add(1 * time.Second)
	timeC := initialTime.Add(3 * time.Second)
	timeD := initialTime.Add(4 * time.Second)

	valueA := 1
	valueB := 4
	valueC := 7
	valueD := 0

	// Add the values to the map
	NowFunc = func() time.Time { return timeA }
	stalenessMap.Store(idA, valueA)
	NowFunc = func() time.Time { return timeB }
	stalenessMap.Store(idB, valueB)
	NowFunc = func() time.Time { return timeC }
	stalenessMap.Store(idC, valueC)
	NowFunc = func() time.Time { return timeD }
	stalenessMap.Store(idD, valueD)

	// Set the time to 2.5s and run expire
	// This should remove B, but the others should remain
	// (now == 2.5s, B == 1s, max == 1s)
	// now > B + max
	NowFunc = func() time.Time { return initialTime.Add(2500 * time.Millisecond) }
	stalenessMap.ExpireOldEntries()
	validateStalenessMapEntries(t,
		map[identity.Stream]int{
			idA: valueA,
			idC: valueC,
			idD: valueD,
		},
		stalenessMap,
	)

	// Set the time to 4.5s and run expire
	// This should remove A and C, but D should remain
	// (now == 2.5s, A == 2s, C == 3s, max == 1s)
	// now > A + max AND now > C + max
	NowFunc = func() time.Time { return initialTime.Add(4500 * time.Millisecond) }
	stalenessMap.ExpireOldEntries()
	validateStalenessMapEntries(t,
		map[identity.Stream]int{
			idD: valueD,
		},
		stalenessMap,
	)
}

func validateStalenessMapEntries(t *testing.T, expected map[identity.Stream]int, sm *Staleness[int]) {
	actual := map[identity.Stream]int{}

	sm.Items()(func(key identity.Stream, value int) bool {
		actual[key] = value
		return true
	})
	require.Equal(t, expected, actual)
}
