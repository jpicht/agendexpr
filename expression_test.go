package agendexpr

import (
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/stretchr/testify/assert"
)

func TestExpression(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		expectErr error
	}{
		{
			name: "success",
			expr: "* * * * * *",
		},
		{
			name:      "invalid format",
			expr:      "* * * * * * *",
			expectErr: ErrInvalidFormat,
		},
		{
			name:      "empty",
			expr:      "",
			expectErr: ErrEmpty,
		},
		{
			name:      "invalid digits",
			expr:      "-8",
			expectErr: ErrParseDigits,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.expr)
			if test.expectErr != nil {
				assert.ErrorIs(t, err, test.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExpression_Next(t *testing.T) {
	baseTime := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		expr     string
		from     time.Time
		wantTime time.Time
	}{
		{
			name:     "any",
			expr:     "* * * * * *",
			from:     baseTime,
			wantTime: baseTime.Add(time.Second),
		},
		{
			name:     "next minute",
			expr:     "0 * * * * *",
			from:     baseTime.Add(time.Second * 30),
			wantTime: baseTime.Add(time.Second * 30).Truncate(time.Minute).Add(time.Minute),
		},
		{
			name:     "hourly step reaches last hour",
			expr:     "0 0 */1 * * *",
			from:     time.Date(2026, 7, 24, 22, 30, 0, 0, time.UTC),
			wantTime: time.Date(2026, 7, 24, 23, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := Parse(test.expr)
			assert.NoError(t, err)

			got := expr.Next(test.from)
			assert.Equal(t, test.wantTime, got)
		})
	}
}

func TestExpression_Previous(t *testing.T) {
	baseTime := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)
	utcPlus2 := time.FixedZone("UTC+2", 2*60*60)

	tests := []struct {
		name     string
		expr     string
		from     time.Time
		wantTime time.Time
	}{
		{
			name:     "any",
			expr:     "* * * * * *",
			from:     baseTime,
			wantTime: baseTime.Add(-time.Second),
		},
		{
			name:     "sub-second from includes its own second",
			expr:     "* * * * * *",
			from:     baseTime.Add(500 * time.Millisecond),
			wantTime: baseTime,
		},
		{
			name:     "from itself is excluded",
			expr:     "0 0 * * * *",
			from:     baseTime,
			wantTime: baseTime.Add(-time.Hour),
		},
		{
			name:     "previous minute",
			expr:     "0 * * * * *",
			from:     baseTime.Add(time.Second * 30),
			wantTime: baseTime,
		},
		{
			name:     "earlier second in same minute",
			expr:     "15 * * * * *",
			from:     baseTime.Add(time.Second * 30),
			wantTime: baseTime.Add(time.Second * 15),
		},
		{
			name:     "earlier minute in same hour",
			expr:     "0 20 * * * *",
			from:     baseTime.Add(time.Minute * 45),
			wantTime: baseTime.Add(time.Minute * 20),
		},
		{
			name:     "previous day",
			expr:     "0 30 11 * * *",
			from:     baseTime,
			wantTime: time.Date(2026, 7, 23, 11, 30, 0, 0, time.UTC),
		},
		{
			name:     "maximum field values",
			expr:     "59 59 23 * * *",
			from:     baseTime,
			wantTime: time.Date(2026, 7, 23, 23, 59, 59, 0, time.UTC),
		},
		{
			name:     "day of week",
			expr:     "0 0 12 * * 1",
			from:     baseTime,
			wantTime: time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "previous year",
			expr:     "0 0 0 1 12 *",
			from:     baseTime,
			wantTime: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "leap day",
			expr:     "0 0 0 29 2 *",
			from:     baseTime,
			wantTime: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "no match within 4 years",
			expr:     "0 0 0 31 2 *",
			from:     baseTime,
			wantTime: time.Time{},
		},
		{
			name:     "preserves location",
			expr:     "0 0 8 * * *",
			from:     time.Date(2026, 7, 24, 10, 0, 0, 0, utcPlus2),
			wantTime: time.Date(2026, 7, 24, 8, 0, 0, 0, utcPlus2),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := Parse(test.expr)
			assert.NoError(t, err)

			got := expr.Previous(test.from)
			assert.Equal(t, test.wantTime, got)
		})
	}
}

func TestExpression_Previous_DST(t *testing.T) {
	berlin := mustLoadLocation(t, "Europe/Berlin")
	newYork := mustLoadLocation(t, "America/New_York")
	santiago := mustLoadLocation(t, "America/Santiago")
	amman := mustLoadLocation(t, "Asia/Amman")

	utc := func(loc *time.Location, year int, month time.Month, day, hour, min, sec int) time.Time {
		return time.Date(year, month, day, hour, min, sec, 0, time.UTC).In(loc)
	}

	tests := []struct {
		name     string
		expr     string
		from     time.Time
		wantTime time.Time
	}{
		{
			name:     "skipped hour falls back to previous day",
			expr:     "0 30 2 * * *",
			from:     time.Date(2026, 3, 29, 12, 0, 0, 0, berlin),
			wantTime: time.Date(2026, 3, 28, 2, 30, 0, 0, berlin),
		},
		{
			name:     "hour after skipped hour",
			expr:     "0 30 3 * * *",
			from:     time.Date(2026, 3, 29, 12, 0, 0, 0, berlin),
			wantTime: time.Date(2026, 3, 29, 3, 30, 0, 0, berlin),
		},
		{
			name:     "repeated hour east of UTC picks later occurrence",
			expr:     "0 30 2 * * *",
			from:     utc(berlin, 2026, 10, 25, 3, 0, 0),
			wantTime: utc(berlin, 2026, 10, 25, 1, 30, 0),
		},
		{
			name:     "repeated hour east of UTC from within later occurrence",
			expr:     "0 30 2 * * *",
			from:     utc(berlin, 2026, 10, 25, 1, 10, 0),
			wantTime: utc(berlin, 2026, 10, 25, 0, 30, 0),
		},
		{
			name:     "repeated hour east of UTC from within earlier occurrence",
			expr:     "0 5 2 * * *",
			from:     utc(berlin, 2026, 10, 25, 0, 10, 0),
			wantTime: utc(berlin, 2026, 10, 25, 0, 5, 0),
		},
		{
			name:     "repeated hour west of UTC picks later occurrence",
			expr:     "0 30 1 * * *",
			from:     utc(newYork, 2026, 11, 1, 8, 0, 0),
			wantTime: utc(newYork, 2026, 11, 1, 6, 30, 0),
		},
		{
			name:     "skipped midnight",
			expr:     "59 59 23 5 9 *",
			from:     time.Date(2026, 9, 6, 12, 0, 0, 0, santiago),
			wantTime: utc(santiago, 2026, 9, 6, 3, 59, 59),
		},
		{
			name:     "repeated midnight from earlier occurrence",
			expr:     "59 59 23 28 10 *",
			from:     utc(amman, 2021, 10, 28, 21, 30, 0),
			wantTime: utc(amman, 2021, 10, 28, 20, 59, 59),
		},
		{
			name:     "repeated midnight from later occurrence",
			expr:     "59 59 23 28 10 *",
			from:     utc(amman, 2021, 10, 28, 22, 30, 0),
			wantTime: utc(amman, 2021, 10, 28, 20, 59, 59),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := Parse(test.expr)
			assert.NoError(t, err)

			got := expr.Previous(test.from)
			assert.Equal(t, test.wantTime, got)
		})
	}
}

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("load location %q: %v", name, err)
	}
	return loc
}

func TestExpression_DueMonth(t *testing.T) {
	tests := []struct {
		name string
		expr string
		of   time.Time
		want bool
	}{
		{
			name: "match",
			expr: "* * * * 7 *",
			of:   time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "mismatch",
			expr: "* * * * 8 *",
			of:   time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC),
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := Parse(test.expr)
			assert.NoError(t, err)

			got := expr.DueMonth(test.of)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestExpression_DueDay(t *testing.T) {
	tests := []struct {
		name string
		expr string
		of   time.Time
		want bool
	}{
		{
			name: "match",
			expr: "* * * 24 7 5",
			of:   time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "mismatch",
			expr: "* * * 25 7 *",
			of:   time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC),
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := Parse(test.expr)
			assert.NoError(t, err)

			got := expr.DueDay(test.of)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestExpression_DueHour(t *testing.T) {
	tests := []struct {
		name string
		expr string
		of   time.Time
		want bool
	}{
		{
			name: "match",
			expr: "* * 10 24 7 *",
			of:   time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "mismatch",
			expr: "* * 11 24 7 *",
			of:   time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC),
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := Parse(test.expr)
			assert.NoError(t, err)

			got := expr.DueHour(test.of)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestExpression_DueMinute(t *testing.T) {
	tests := []struct {
		name string
		expr string
		of   time.Time
		want bool
	}{
		{
			name: "match",
			expr: "* 30 10 24 7 *",
			of:   time.Date(2026, 7, 24, 10, 30, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "mismatch",
			expr: "* 0 10 24 7 *",
			of:   time.Date(2026, 7, 24, 10, 30, 0, 0, time.UTC),
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := Parse(test.expr)
			assert.NoError(t, err)

			got := expr.DueMinute(test.of)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestExpression_Due(t *testing.T) {
	tests := []struct {
		name string
		expr string
		of   time.Time
		want bool
	}{
		{
			name: "match",
			expr: "30 15 14 * * *",
			of:   time.Date(2026, 1, 1, 14, 15, 30, 0, time.UTC),
			want: true,
		},
		{
			name: "mismatch",
			expr: "0 15 14 * * *",
			of:   time.Date(2026, 1, 1, 14, 15, 30, 0, time.UTC),
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := Parse(test.expr)
			assert.NoError(t, err)

			got := expr.Due(test.of)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestExpression_Countdown(t *testing.T) {
	baseTime := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		expr string
		from time.Time
		want int64
	}{
		{
			name: "a second",
			expr: "* * * * * *",
			from: baseTime,
			want: 1,
		},
		{
			name: "a hour",
			expr: "0 0 * * * *",
			from: baseTime.Add(time.Minute * 30),
			want: 1800,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := Parse(test.expr)
			assert.NoError(t, err)

			got := expr.Countdown(test.from)
			assert.Equal(t, test.want, got)
		})
	}
}

func BenchmarkParse(b *testing.B) {
	for b.Loop() {
		_, _ = Parse("0 15 10 1 * *")
	}
}

func BenchmarkExpression_Next(b *testing.B) {
	expr, _ := Parse("0 0 1 1 * *")
	baseTime := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)

	for b.Loop() {
		_ = expr.Next(baseTime)
	}
}

func BenchmarkExpression_Due(b *testing.B) {
	expr, _ := Parse("30 15 14 * * *")
	checkTime := time.Date(2026, 7, 24, 14, 15, 30, 0, time.UTC)

	for b.Loop() {
		_ = expr.Due(checkTime)
	}
}
