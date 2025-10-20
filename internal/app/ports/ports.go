package ports

import (
	"context"
	"math"
)

type PaginationParams struct {
	Page     int
	PageSize int
}

type Page[T any] struct {
	Items      []T
	Page       int
	PageSize   int
	TotalCount int64
}

func (p Page[T]) TotalPages() int {
	if p.PageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(p.TotalCount) / float64(p.PageSize)))
}

type ObjectStorage interface {
	Upload(ctx context.Context, bucket, object string, data []byte, contentType string, metadata map[string]string) error
	Delete(ctx context.Context, bucket, object string) error
	Get(ctx context.Context, bucket, object string) ([]byte, string, error)
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) error
}
