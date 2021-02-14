package loadgen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/swtch1/lg/store"
)

func TestLastNLatencies_SlicesCorrectAmt(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		in       []store.AggLatency
		n        int
		expected int
	}{
		{
			name:     "nil",
			in:       nil,
			n:        5,
			expected: 0,
		},
		{
			name:     "0",
			in:       []store.AggLatency{},
			n:        5,
			expected: 0,
		},
		{
			name: "2",
			in: []store.AggLatency{
				{},
				{},
			},
			n:        5,
			expected: 2,
		},
		{
			name: "n",
			in: []store.AggLatency{ // 3 total
				{},
				{},
				{},
			},
			n:        3,
			expected: 3,
		},
		{
			name: "more than n",
			in: []store.AggLatency{ // 5 total
				{},
				{},
				{},
				{},
				{},
			},
			n:        3,
			expected: 3,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out := lastNLatencies(tt.in, tt.n)
			require.Equal(t, tt.expected, len(out))
		})
	}
}

func TestIncrease(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		upper    float64
		lower    float64
		pct      float64
		expected float64
	}{
		{
			name:     "one",
			upper:    5000,
			lower:    4000,
			pct:      0.1,
			expected: 4100,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := increaseFromLower(tt.lower, tt.upper, tt.pct)
			require.Equal(t, tt.expected, f)
		})
	}
}

func TestDecrease(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		upper    float64
		lower    float64
		pct      float64
		expected float64
	}{
		{
			name:     "one",
			upper:    5000,
			lower:    4000,
			pct:      0.1,
			expected: 4900,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := decreaseFromUpper(tt.lower, tt.upper, tt.pct)
			require.Equal(t, tt.expected, f)
		})
	}
}
