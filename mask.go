package agendexpr

import "math/bits"

// Mask is a bitmask representation of a parsed cron field.
// The type parameter T determines the bit width: uint64 for seconds/minutes
// (0–59), uint32 for hours/day-of-month, uint16 for months, uint8 for day-of-week.
// Each set bit at position i means value i is included in the field.
type Mask[T MaskType] struct {
	raw T
}

// Raw returns the underlying bitmask value.
func (m Mask[T]) Raw() T {
	return m.raw
}

// Next returns the smallest value greater than from that is set in the mask.
// If no such value exists, it wraps around and returns the smallest set bit
// in the entire mask. Returns 0 if the mask is empty.
func (m Mask[T]) Next(from uint8) uint8 {
	var filtered T
	if from < 63 {
		mask := ^T(0) << (from + 1)
		filtered = m.raw & mask
	} else {
		filtered = 0
	}

	if filtered != 0 {
		return uint8(bits.TrailingZeros64(uint64(filtered)))
	}

	return uint8(bits.TrailingZeros64(uint64(m.raw)))
}

// Previous returns the largest value less than from that is set in the mask.
// If no such value exists, it wraps around and returns the largest set bit
// in the entire mask. Returns 0 if the mask is empty.
func (m Mask[T]) Previous(from uint8) uint8 {
	if below := uint64(m.raw) & (1<<from - 1); below != 0 {
		return uint8(bits.Len64(below) - 1)
	}
	if m.raw == 0 {
		return 0
	}
	return uint8(bits.Len64(uint64(m.raw)) - 1)
}

// Match reports whether val is set in the mask.
func (m Mask[T]) Match(val uint8) bool {
	return (m.raw & (T(1) << val)) != 0
}

// List returns a slice of all values set in the mask, in ascending order.
// Returns nil if no bits are set.
func (m Mask[T]) List() []uint8 {
	var result []uint8
	v := uint64(m.raw)

	for v != 0 {
		tz := uint8(bits.TrailingZeros64(v))
		result = append(result, tz)
		v &= v - 1
	}

	return result
}
