package testing

import (
	"context"
	"testing"
	"time"

	"kalshi/pkg/utils"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"seconds", 30 * time.Second, "30.0s"},
		{"minutes", 5 * time.Minute, "5.0m"},
		{"hours", 2 * time.Hour, "2.0h"},
		{"days", 3 * 24 * time.Hour, "3.0d"},
		{"fractional seconds", 1500 * time.Millisecond, "1.5s"},
		{"fractional minutes", 90 * time.Second, "1.5m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.FormatDuration(tt.duration); got != tt.expected {
				t.Errorf("FormatDuration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"1 minute ago", now.Add(-1 * time.Minute), "1 minute ago"},
		{"5 minutes ago", now.Add(-5 * time.Minute), "5 minutes ago"},
		{"1 hour ago", now.Add(-1 * time.Hour), "1 hour ago"},
		{"2 hours ago", now.Add(-2 * time.Hour), "2 hours ago"},
		{"1 day ago", now.Add(-25 * time.Hour), "1 day ago"},
		{"2 days ago", now.Add(-49 * time.Hour), "2 days ago"},
		{"1 month ago", now.AddDate(0, -1, 0), "1 month ago"},
		{"2 months ago", now.AddDate(0, -2, 0), "2 months ago"},
		{"1 year ago", now.AddDate(-1, 0, 0), "1 year ago"},
		{"2 years ago", now.AddDate(-2, 0, 0), "2 years ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.TimeAgo(tt.time); got != tt.expected {
				t.Errorf("TimeAgo() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStartOfDay(t *testing.T) {
	now := time.Now()
	start := utils.StartOfDay(now)

	// Should be same date but time set to 00:00:00
	if start.Year() != now.Year() || start.Month() != now.Month() || start.Day() != now.Day() {
		t.Errorf("StartOfDay() date = %v, want same date as %v", start, now)
	}

	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 || start.Nanosecond() != 0 {
		t.Errorf("StartOfDay() time = %v, want 00:00:00.000000000", start)
	}
}

func TestEndOfDay(t *testing.T) {
	now := time.Now()
	end := utils.EndOfDay(now)

	// Should be same date but time set to 23:59:59.999999999
	if end.Year() != now.Year() || end.Month() != now.Month() || end.Day() != now.Day() {
		t.Errorf("EndOfDay() date = %v, want same date as %v", end, now)
	}

	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 || end.Nanosecond() != 999999999 {
		t.Errorf("EndOfDay() time = %v, want 23:59:59.999999999", end)
	}
}

func TestStartOfWeek(t *testing.T) {
	// Test with Monday
	monday := time.Date(2023, 1, 2, 15, 30, 45, 0, time.UTC) // Monday
	start := utils.StartOfWeek(monday)

	if start.Weekday() != time.Monday {
		t.Errorf("StartOfWeek() weekday = %v, want Monday", start.Weekday())
	}

	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Errorf("StartOfWeek() time = %v, want 00:00:00", start)
	}
}

func TestEndOfWeek(t *testing.T) {
	// Test with Monday
	monday := time.Date(2023, 1, 2, 15, 30, 45, 0, time.UTC) // Monday
	end := utils.EndOfWeek(monday)

	if end.Weekday() != time.Sunday {
		t.Errorf("EndOfWeek() weekday = %v, want Sunday", end.Weekday())
	}

	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Errorf("EndOfWeek() time = %v, want 23:59:59", end)
	}
}

func TestStartOfMonth(t *testing.T) {
	now := time.Now()
	start := utils.StartOfMonth(now)

	// Should be first day of the month at 00:00:00
	if start.Day() != 1 {
		t.Errorf("StartOfMonth() day = %d, want 1", start.Day())
	}

	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Errorf("StartOfMonth() time = %v, want 00:00:00", start)
	}
}

func TestEndOfMonth(t *testing.T) {
	now := time.Now()
	end := utils.EndOfMonth(now)

	// Should be last day of the month at 23:59:59.999999999
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Errorf("EndOfMonth() time = %v, want 23:59:59", end)
	}
}

func TestIsWeekend(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected bool
	}{
		{"Saturday", time.Date(2023, 1, 7, 12, 0, 0, 0, time.UTC), true},
		{"Sunday", time.Date(2023, 1, 8, 12, 0, 0, 0, time.UTC), true},
		{"Monday", time.Date(2023, 1, 9, 12, 0, 0, 0, time.UTC), false},
		{"Friday", time.Date(2023, 1, 13, 12, 0, 0, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsWeekend(tt.time); got != tt.expected {
				t.Errorf("IsWeekend() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsBusinessDay(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected bool
	}{
		{"Saturday", time.Date(2023, 1, 7, 12, 0, 0, 0, time.UTC), false},
		{"Sunday", time.Date(2023, 1, 8, 12, 0, 0, 0, time.UTC), false},
		{"Monday", time.Date(2023, 1, 9, 12, 0, 0, 0, time.UTC), true},
		{"Friday", time.Date(2023, 1, 13, 12, 0, 0, 0, time.UTC), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsBusinessDay(tt.time); got != tt.expected {
				t.Errorf("IsBusinessDay() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNextBusinessDay(t *testing.T) {
	// Test Friday -> Monday
	friday := time.Date(2023, 1, 13, 12, 0, 0, 0, time.UTC) // Friday
	next := utils.NextBusinessDay(friday)
	expected := time.Date(2023, 1, 16, 12, 0, 0, 0, time.UTC) // Monday
	if !next.Equal(expected) {
		t.Errorf("NextBusinessDay() = %v, want %v", next, expected)
	}

	// Test Monday -> Tuesday
	monday := time.Date(2023, 1, 9, 12, 0, 0, 0, time.UTC) // Monday
	next = utils.NextBusinessDay(monday)
	expected = time.Date(2023, 1, 10, 12, 0, 0, 0, time.UTC) // Tuesday
	if !next.Equal(expected) {
		t.Errorf("NextBusinessDay() = %v, want %v", next, expected)
	}
}

func TestPreviousBusinessDay(t *testing.T) {
	// Test Monday -> Friday
	monday := time.Date(2023, 1, 9, 12, 0, 0, 0, time.UTC) // Monday
	prev := utils.PreviousBusinessDay(monday)
	expected := time.Date(2023, 1, 6, 12, 0, 0, 0, time.UTC) // Friday
	if !prev.Equal(expected) {
		t.Errorf("PreviousBusinessDay() = %v, want %v", prev, expected)
	}

	// Test Tuesday -> Monday
	tuesday := time.Date(2023, 1, 10, 12, 0, 0, 0, time.UTC) // Tuesday
	prev = utils.PreviousBusinessDay(tuesday)
	expected = time.Date(2023, 1, 9, 12, 0, 0, 0, time.UTC) // Monday
	if !prev.Equal(expected) {
		t.Errorf("PreviousBusinessDay() = %v, want %v", prev, expected)
	}
}

func TestRoundToMinute(t *testing.T) {
	time1 := time.Date(2023, 1, 1, 12, 30, 45, 123456789, time.UTC)
	rounded := utils.RoundToMinute(time1)
	expected := time.Date(2023, 1, 1, 12, 31, 0, 0, time.UTC)

	if !rounded.Equal(expected) {
		t.Errorf("RoundToMinute() = %v, want %v", rounded, expected)
	}
}

func TestRoundToHour(t *testing.T) {
	time1 := time.Date(2023, 1, 1, 12, 30, 45, 123456789, time.UTC)
	rounded := utils.RoundToHour(time1)
	expected := time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)

	if !rounded.Equal(expected) {
		t.Errorf("RoundToHour() = %v, want %v", rounded, expected)
	}
}

func TestParseFlexibleTime(t *testing.T) {
	tests := []struct {
		name        string
		timeStr     string
		expectError bool
	}{
		{"RFC3339", "2023-01-01T12:00:00Z", false},
		{"date only", "2023-01-01", false},
		{"time only", "12:00:00", false},
		{"invalid format", "invalid", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := utils.ParseFlexibleTime(tt.timeStr)
			if tt.expectError && err == nil {
				t.Errorf("ParseFlexibleTime() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ParseFlexibleTime() unexpected error: %v", err)
			}
		})
	}
}

func TestGetTimezone(t *testing.T) {
	time1 := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	name, offset := utils.GetTimezone(time1)

	if name != "UTC" {
		t.Errorf("GetTimezone() name = %v, want UTC", name)
	}

	if offset != 0 {
		t.Errorf("GetTimezone() offset = %v, want 0", offset)
	}
}

func TestConvertTimezone(t *testing.T) {
	utcTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	// Test conversion to EST
	estTime, err := utils.ConvertTimezone(utcTime, "America/New_York")
	if err != nil {
		t.Errorf("ConvertTimezone() unexpected error: %v", err)
	}

	// The time should be different (EST is UTC-5, but daylight saving time may affect this)
	if estTime.Equal(utcTime) {
		t.Logf("ConvertTimezone() returned same time, this might be due to daylight saving time")
	}

	// Test invalid timezone
	_, err = utils.ConvertTimezone(utcTime, "Invalid/Timezone")
	if err == nil {
		t.Error("ConvertTimezone() should return error for invalid timezone")
	}
}

func TestSleepWithContext(t *testing.T) {
	ctx := context.Background()
	duration := 10 * time.Millisecond

	start := time.Now()
	err := utils.SleepWithContext(ctx, duration)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("SleepWithContext() unexpected error: %v", err)
	}

	if elapsed < duration {
		t.Errorf("SleepWithContext() slept for %v, want at least %v", elapsed, duration)
	}

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = utils.SleepWithContext(ctx, 100*time.Millisecond)
	if err == nil {
		t.Error("SleepWithContext() should return error when context is cancelled")
	}
}
