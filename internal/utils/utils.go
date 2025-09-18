package utils

import (
	"time"

	"github.com/google/uuid"
)

func GetString(s *string) string {
    if s != nil {
        return *s
    }
    return ""
}

func GetBool(b *bool) bool {
    if b != nil {
        return *b
    }
    return false
}

func GetTime(t *time.Time) time.Time {
    if t != nil {
        return *t
    }
    return time.Time{}
}

func GetInt(t *int) int{
	if t != nil{
		return *t
	}
	return 0
}

func GetInt64(t *int64) int64{
	if t != nil{
		return *t
	}
	return 0
}

func GetInt16(t *int16) int16{
	if t != nil{
		return *t
	}
	return 0
}

func GetInt8(t *int8) int8{
	if t != nil{
		return *t
	}
	return 0
}

func GetUUIDString(u *uuid.UUID) string {
    if u != nil {
        return u.String()
    }
    return ""
}
