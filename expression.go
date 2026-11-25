package agendexpr

import (
	"time"
)

// Expression represents a parsed cron-like schedule expression.
// It stores each field (seconds, minutes, hours, day-of-month, month,
// day-of-week) as a bitmask for O(1) match checks.
//
// Use [Parse] or [MustParse] to create an Expression.
type Expression struct {
	seconds    Mask[uint64]
	minutes    Mask[uint64]
	hours      Mask[uint32]
	dayOfMonth Mask[uint32]
	months     Mask[uint16]
	dayOfWeek  Mask[uint8]
}

func (expr Expression) GetSeconds() Mask[uint64]    { return expr.seconds }
func (expr Expression) GetMinutes() Mask[uint64]    { return expr.minutes }
func (expr Expression) GetHours() Mask[uint32]      { return expr.hours }
func (expr Expression) GetDayOfMonth() Mask[uint32] { return expr.dayOfMonth }
func (expr Expression) GetMonths() Mask[uint16]     { return expr.months }
func (expr Expression) GetDayOfWeek() Mask[uint8]   { return expr.dayOfWeek }

// Next returns the next [time.Time] after from that matches the expression.
// The search is bounded to 4 years from from; if no match is found within
// that window, a zero [time.Time] is returned. The returned time preserves
// the location of from.
func (expr Expression) Next(from time.Time) time.Time {
	loc := from.Location()
	t := from.Truncate(time.Second).Add(time.Second)

	limit := t.AddDate(4, 0, 0)

	for t.Before(limit) {
		yr := t.Year()
		mon := int(t.Month())
		dom := t.Day()
		hr := t.Hour()
		min := t.Minute()
		sec := t.Second()

		if !expr.months.Match(uint8(mon)) {
			next := expr.months.Next(uint8(mon))
			if next < uint8(mon) {
				yr++
			}
			t = time.Date(yr, time.Month(next), 1, 0, 0, 0, 0, loc)
			continue
		}

		if !expr.dayOfMonth.Match(uint8(dom)) || !expr.dayOfWeek.Match(uint8(t.Weekday())) {
			t = time.Date(yr, time.Month(mon), dom+1, 0, 0, 0, 0, loc)
			continue
		}

		if !expr.hours.Match(uint8(hr)) {
			next := expr.hours.Next(uint8(hr))
			if next < uint8(hr) {
				t = time.Date(yr, time.Month(mon), dom+1, 0, 0, 0, 0, loc)
			} else {
				t = time.Date(yr, time.Month(mon), dom, int(next), 0, 0, 0, loc)
			}
			continue
		}

		if !expr.minutes.Match(uint8(min)) {
			next := expr.minutes.Next(uint8(min))
			if next < uint8(min) {
				t = time.Date(yr, time.Month(mon), dom, hr+1, 0, 0, 0, loc)
			} else {
				t = time.Date(yr, time.Month(mon), dom, hr, int(next), 0, 0, loc)
			}
			continue
		}

		if !expr.seconds.Match(uint8(sec)) {
			next := expr.seconds.Next(uint8(sec))
			if next < uint8(sec) {
				t = time.Date(yr, time.Month(mon), dom, hr, min+1, 0, 0, loc)
			} else {
				t = time.Date(yr, time.Month(mon), dom, hr, min, int(next), 0, loc)
			}
			continue
		}

		return t
	}

	return time.Time{}
}

// Previous returns the latest [time.Time] before from that matches the
// expression, at whole-second precision. The search is bounded to 4 years
// before from; if no match is found within that window, a zero [time.Time]
// is returned. The returned time preserves the location of from.
func (expr Expression) Previous(from time.Time) time.Time {
	t := from.Add(-time.Nanosecond).Truncate(time.Second)

	limit := t.AddDate(-4, 0, 0)

	// Within a day, step back by elapsed seconds instead of using time.Date:
	// for wall times skipped or repeated by a DST transition, time.Date may
	// pick an instant after t, which would stall the search or overshoot.
	for t.After(limit) {
		yr, mon, dom := t.Date()
		hr, min, sec := t.Clock()

		if !expr.months.Match(uint8(mon)) {
			prev := expr.months.Previous(uint8(mon))
			if prev > uint8(mon) {
				yr--
			}
			t = endOfDayBefore(t, yr, time.Month(prev)+1, 1)
			continue
		}

		if !expr.dayOfMonth.Match(uint8(dom)) || !expr.dayOfWeek.Match(uint8(t.Weekday())) {
			t = endOfDayBefore(t, yr, mon, dom)
			continue
		}

		if !expr.hours.Match(uint8(hr)) {
			t = endOfPreviousHour(t)
			continue
		}

		if !expr.minutes.Match(uint8(min)) {
			prev := int(expr.minutes.Previous(uint8(min)))
			if prev >= min {
				t = endOfPreviousHour(t)
			} else {
				t = t.Add(-time.Duration((min-prev-1)*60+sec+1) * time.Second)
			}
			continue
		}

		if !expr.seconds.Match(uint8(sec)) {
			prev := int(expr.seconds.Previous(uint8(sec)))
			if prev >= sec {
				t = t.Add(-time.Duration(sec+1) * time.Second)
			} else {
				t = t.Add(-time.Duration(sec-prev) * time.Second)
			}
			continue
		}

		return t
	}

	return time.Time{}
}

func endOfPreviousHour(t time.Time) time.Time {
	_, min, sec := t.Clock()
	return t.Add(-time.Duration(min*60+sec+1) * time.Second)
}

// endOfDayBefore returns the last second before the given date starts. When a
// DST transition skips or repeats midnight, time.Date can resolve midnight into
// the previous day or to an instant after t; it then falls back to
// endOfPreviousHour(t), which always moves back without skipping matches.
func endOfDayBefore(t time.Time, year int, month time.Month, day int) time.Time {
	start := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	end := start.Add(-time.Second)
	if start.Day() != day || !end.Before(t) {
		return endOfPreviousHour(t)
	}
	return end
}

// DueMonth reports whether of's month matches the expression.
func (expr Expression) DueMonth(of time.Time) bool {
	return expr.months.Match(uint8(of.Month()))
}

// DueDay reports whether of's day-of-month and day-of-week match the expression.
// It also checks the month via [Expression.DueMonth].
func (expr Expression) DueDay(of time.Time) bool {
	return expr.DueMonth(of) &&
		expr.dayOfMonth.Match(uint8(of.Day())) &&
		expr.dayOfWeek.Match(uint8(of.Weekday()))
}

// DueHour reports whether of's hour matches the expression.
// It also checks day and month via [Expression.DueDay].
func (expr Expression) DueHour(of time.Time) bool {
	return expr.DueDay(of) && expr.hours.Match(uint8(of.Hour()))
}

// DueMinute reports whether of's minute matches the expression.
// It also checks hour, day, and month via [Expression.DueHour].
func (expr Expression) DueMinute(of time.Time) bool {
	return expr.DueHour(of) && expr.minutes.Match(uint8(of.Minute()))
}

// Due reports whether of fully matches the expression at second precision.
// It checks all fields: month, day-of-month, day-of-week, hour, minute, and second.
func (expr Expression) Due(of time.Time) bool {
	return expr.DueMinute(of) && expr.seconds.Match(uint8(of.Second()))
}

// Countdown returns the number of seconds from from until the next match.
// Returns -1 if no next match is found within 4 years.
func (expr Expression) Countdown(from time.Time) int64 {
	next := expr.Next(from)
	if next.IsZero() {
		return -1
	}
	return next.Unix() - from.Unix()
}

// CountdownNow returns the number of seconds from [time.Now] until the next match.
// Equivalent to calling [Expression.Countdown] with [time.Now].
func (expr Expression) CountdownNow() int64 {
	return expr.Countdown(time.Now())
}
