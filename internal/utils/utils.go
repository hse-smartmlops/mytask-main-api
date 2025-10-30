package utils

import (
	"fmt"
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

func FormatTimeForExcel(t time.Time) string {
    if t.IsZero() {
        return ""
    }
    return t.Format("02.01.2006")
}

func ConvertPriorityToText(priority int16) string {
    switch priority {
    case 5:
        return "Срочный"
    case 4:
        return "Высокий"
    case 3:
        return "Средний"
    case 2:
        return "Низкий"
    case 1:
        return "Не задан"
    default:
        return fmt.Sprintf("Неизвестный (%d)", priority)
    }
}

func BoolPtr(b bool) *bool {
	return &b
}