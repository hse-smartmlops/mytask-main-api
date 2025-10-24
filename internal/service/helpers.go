package service

import (
	"bytes"
	"strings"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	"emplacc-api/internal/domain"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
)

func stringPtr(value string) *string {
	return &value
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	v := strings.TrimSpace(*value)
	if v == "" {
		return nil
	}
	return &v
}

func sanitizeDescription(items []string) []string {
	clean := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	return clean
}

func checkPagination(p ports.PaginationParams, config *config.PaginationConfig) (page, pagesize int, ok bool) {
	ok = true

	if p.Page <= 0 {
		page = 1
		ok = false
	} else {
		page = p.Page
	}

	if p.PageSize <= 0 || p.PageSize > config.MaxLimit {
		pagesize = config.DefaultLimit
		ok = false
	} else {
		pagesize = p.PageSize
	}

	return page, pagesize, ok
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func int64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func prepareAvatar(data []byte) ([]byte, string, error) {
	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", domain.ErrInvalidInput
	}

	avatar := imaging.Fill(img, 256, 256, imaging.Center, imaging.Lanczos)
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, avatar, imaging.PNG); err != nil {
		return nil, "", err
	}

	return buf.Bytes(), "image/png", nil
}
