package agendexpr

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMask_Match(t *testing.T) {
	tests := []struct {
		name  string
		token Mask[uint64]
		val   uint8
		expect  bool
	}{
		{
			name:  "matched",
			token: Mask[uint64]{raw: (1 << 5) | (1 << 10)},
			val:   5,
			expect:  true,
		},
		{
			name:  "not matched",
			token: Mask[uint64]{raw: (1 << 5) | (1 << 10)},
			val:   6,
			expect:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.Match(tt.val); got != tt.expect {
				t.Errorf("Mask.Match(%d) = %v, expect %v", tt.val, got, tt.expect)
			}
		})
	}
}

func TestMask_List(t *testing.T) {
	tests := []struct {
		name  string
		token Mask[uint64]
		expect  []uint8
	}{
		{
			name:  "empty",
			token: Mask[uint64]{raw: 0},
			expect:  nil,
		},
		{
			name:  "only one",
			token: Mask[uint64]{raw: 1 << 5},
			expect:  []uint8{5},
		},
		{
			name:  "many",
			token: Mask[uint64]{raw: (1 << 1) | (1 << 15) | (1 << 50)},
			expect:  []uint8{1, 15, 50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.token.List()
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("Mask.List() = %v, expect %v", got, tt.expect)
			}
		})
	}
}

func TestMask_Next(t *testing.T) {
	tests := []struct {
		name  string
		token Mask[uint64]
		from  uint8
		expect  uint8
	}{
		{
			name:  "next value",
			token: Mask[uint64]{raw: (1 << 5) | (1 << 10)},
			from:  5,
			expect:  10,
		},
		{
			name:  "next-of value",
			token: Mask[uint64]{raw: (1 << 2) | (1 << 50)},
			from:  10,
			expect:  50,
		},
		{
			name:  "wrap",
			token: Mask[uint64]{raw: (1 << 5) | (1 << 10)},
			from:  15,
			expect:  5,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			next := test.token.Next(test.from)

			assert.Equal(t, test.expect, next)
		})
	}
}

func TestMask_Previous(t *testing.T) {
	tests := []struct {
		name   string
		token  Mask[uint64]
		from   uint8
		expect uint8
	}{
		{
			name:   "previous value",
			token:  Mask[uint64]{raw: (1 << 5) | (1 << 10)},
			from:   10,
			expect: 5,
		},
		{
			name:   "previous-of value",
			token:  Mask[uint64]{raw: (1 << 2) | (1 << 50)},
			from:   40,
			expect: 2,
		},
		{
			name:   "wrap",
			token:  Mask[uint64]{raw: (1 << 5) | (1 << 10)},
			from:   3,
			expect: 10,
		},
		{
			name:   "from above mask width",
			token:  Mask[uint64]{raw: (1 << 5) | (1 << 63)},
			from:   200,
			expect: 63,
		},
		{
			name:   "empty",
			token:  Mask[uint64]{raw: 0},
			from:   10,
			expect: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, test.token.Previous(test.from))
		})
	}
}