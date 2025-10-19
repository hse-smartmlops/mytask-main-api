package presenter

import "time"

func valueOrZeroString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func valueOrZeroBool(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

func valueOrZeroInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func valueOrZeroInt16(value *int16) int16 {
	if value == nil {
		return 0
	}
	return *value
}

func valueOrZeroTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}
