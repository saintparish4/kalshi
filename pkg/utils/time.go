package utils

import (
	"context"
	"fmt"
	"time"
)

// Common time constants
const (
	DayInHours  = 24
	WeekInDays  = 7
	MonthInDays = 30 // Average days per month
	YearInDays  = 365
	YearInHours = DayInHours * YearInDays
)

// FormatDuration formats duration in human-readable format.
// Returns duration as "Xs", "Xm", "Xh", or "Xd" where X is the duration value.
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	} else {
		days := d.Hours() / 24
		return fmt.Sprintf("%.1fd", days)
	}
}

// TimeAgo returns human-readable time difference.
// Returns strings like "just now", "5 minutes ago", "2 hours ago", etc.
func TimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if diff < DayInHours*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < MonthInDays*DayInHours*time.Hour {
		days := int(diff.Hours() / DayInHours)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if diff < YearInHours*time.Hour {
		// More accurate month calculation
		months := calculateMonths(t, now)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	} else {
		years := int(diff.Hours() / YearInHours)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// calculateMonths calculates the number of months between two dates more accurately
func calculateMonths(start, end time.Time) int {
	yearDiff := end.Year() - start.Year()
	monthDiff := int(end.Month()) - int(start.Month())

	months := yearDiff*12 + monthDiff

	// Adjust if the day of the month hasn't been reached yet
	if end.Day() < start.Day() {
		months--
	}

	if months < 0 {
		return 0
	}
	return months
}

// StartOfDay returns the start of the day for given time.
// Returns time set to 00:00:00.000000000 of the same day.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day for given time.
// Returns time set to 23:59:59.999999999 of the same day.
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek returns the start of the week (Monday).
// Uses Monday as the first day of the week.
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	return StartOfDay(t.AddDate(0, 0, -weekday+1))
}

// EndOfWeek returns the end of the week (Sunday).
// Returns the end of the week starting from Monday.
func EndOfWeek(t time.Time) time.Time {
	return EndOfDay(StartOfWeek(t).AddDate(0, 0, 6))
}

// StartOfMonth returns the start of the month.
// Returns time set to 00:00:00.000000000 of the first day of the month.
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the end of the month.
// Returns time set to 23:59:59.999999999 of the last day of the month.
func EndOfMonth(t time.Time) time.Time {
	return EndOfDay(StartOfMonth(t).AddDate(0, 1, -1))
}

// IsWeekend checks if the given time is weekend.
// Returns true for Saturday and Sunday, false otherwise.
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsBusinessDay checks if the given time is a business day.
// Returns true for Monday through Friday, false for weekends.
func IsBusinessDay(t time.Time) bool {
	return !IsWeekend(t)
}

// NextBusinessDay returns the next business day.
// Skips weekends and returns the next Monday if current day is Friday.
func NextBusinessDay(t time.Time) time.Time {
	next := t.AddDate(0, 0, 1)
	for IsWeekend(next) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// PreviousBusinessDay returns the previous business day.
// Skips weekends and returns the previous Friday if current day is Monday.
func PreviousBusinessDay(t time.Time) time.Time {
	prev := t.AddDate(0, 0, -1)
	for IsWeekend(prev) {
		prev = prev.AddDate(0, 0, -1)
	}
	return prev
}

// RoundToMinute rounds time to the nearest minute.
// Truncates seconds and nanoseconds.
func RoundToMinute(t time.Time) time.Time {
	return t.Round(time.Minute)
}

// RoundToHour rounds time to the nearest hour.
// Truncates minutes, seconds and nanoseconds.
func RoundToHour(t time.Time) time.Time {
	return t.Round(time.Hour)
}

// ParseFlexibleTime parses time with multiple possible formats.
// Supports RFC3339, ISO formats, date-only, time-only, and various separators.
// Returns error if the time string cannot be parsed with any supported format.
func ParseFlexibleTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("time string cannot be empty")
	}

	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"15:04:05",
		"15:04",
		"01/02/2006",
		"01/02/2006 15:04:05",
		"2006/01/02",
		"2006/01/02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// GetTimezone returns timezone information.
// Returns the timezone name and offset in seconds from UTC.
func GetTimezone(t time.Time) (string, int) {
	zone, offset := t.Zone()
	return zone, offset
}

// ConvertTimezone converts time to different timezone.
// Returns error if the timezone string is invalid or not supported.
func ConvertTimezone(t time.Time, timezone string) (time.Time, error) {
	if timezone == "" {
		return time.Time{}, fmt.Errorf("timezone cannot be empty")
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}
	return t.In(loc), nil
}

// SleepWithContext sleeps for the specified duration with context cancellation support.
// Returns ctx.Err() if the context is cancelled before the duration completes.
// Returns nil if the duration completes successfully.
func SleepWithContext(ctx context.Context, duration time.Duration) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	if duration < 0 {
		return fmt.Errorf("duration cannot be negative")
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
