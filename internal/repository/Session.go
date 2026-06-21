package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"emplacc-api/internal/ports"

	"github.com/redis/go-redis/v9"
)

const (
	ReasonUserExit    = "user_exit"
	ReasonLongAbsence = "long_absence"
	sessionKeyPrefix  = "sess:"
)

type sessionValue struct {
	UserID            string    `json:"user_id"`
	ExpiresAt         time.Time `json:"expires_at"`
	AbsoluteExpiresAt time.Time `json:"absolute_expires_at"`
	ExpireReason      string    `json:"expire_reason"`
}

// SessionData is retained as an alias for backward compatibility;
// the canonical definition now lives in the ports package.
type SessionData = ports.SessionData

type sessionRepository struct{ rdb *redis.Client }

func NewSessionRepository(rdb *redis.Client) ports.SessionRepository {
	return &sessionRepository{rdb: rdb}
}

func rkey(hash string) string { return sessionKeyPrefix + hash }

func (r *sessionRepository) Save(hash string, data SessionData) error {
	b, err := json.Marshal(sessionValue{
		UserID:            data.UserID,
		ExpiresAt:         data.ExpiresAt,
		AbsoluteExpiresAt: data.AbsoluteExpiresAt,
		ExpireReason:      data.ExpireReason,
	})
	if err != nil {
		return err
	}
	ttl := time.Until(data.AbsoluteExpiresAt)
	if ttl <= 0 {
		return errors.New("absolute TTL already passed")
	}
	return r.rdb.Set(context.Background(), rkey(hash), b, ttl).Err()
}

func (r *sessionRepository) Find(hash string) (*SessionData, error) {
	b, err := r.rdb.Get(context.Background(), rkey(hash)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, errors.New("session not found")
	}
	if err != nil {
		return nil, err
	}
	var v sessionValue
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return &SessionData{
		UserID:            v.UserID,
		ExpiresAt:         v.ExpiresAt,
		AbsoluteExpiresAt: v.AbsoluteExpiresAt,
		ExpireReason:      v.ExpireReason,
	}, nil
}

func (r *sessionRepository) Rotate(hash string, newExpiresAt time.Time) error {
	data, err := r.Find(hash)
	if err != nil {
		return err
	}
	b, err := json.Marshal(sessionValue{
		UserID:            data.UserID,
		ExpiresAt:         newExpiresAt,
		AbsoluteExpiresAt: data.AbsoluteExpiresAt,
		ExpireReason:      data.ExpireReason,
	})
	if err != nil {
		return err
	}
	// KeepTTL — абсолютный Redis TTL не меняем
	return r.rdb.SetXX(context.Background(), rkey(hash), b, redis.KeepTTL).Err()
}

func (r *sessionRepository) Expire(hash string, reason string) error {
	data, err := r.Find(hash)
	if err != nil {
		return nil // нечего завершать
	}
	b, err := json.Marshal(sessionValue{
		UserID:            data.UserID,
		ExpiresAt:         time.Now(),
		AbsoluteExpiresAt: data.AbsoluteExpiresAt,
		ExpireReason:      reason,
	})
	if err != nil {
		return err
	}
	// 5 минут чтобы клиент прочитал причину
	return r.rdb.Set(context.Background(), rkey(hash), b, 5*time.Minute).Err()
}

func (r *sessionRepository) Delete(hash string) error {
	return r.rdb.Del(context.Background(), rkey(hash)).Err()
}
