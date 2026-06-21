package service

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/ports"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

type NotificationService interface {
	ports.Notifier
	List(userID uuid.UUID, page, pageSize int, onlyUnread bool) ([]models.Notification, int64, error)
	UnreadCount(userID uuid.UUID) (int64, error)
	MarkRead(id, userID uuid.UUID) error
	MarkAllRead(userID uuid.UUID) error
}

type notificationService struct {
	repo ports.NotificationRepository
}

func NewNotificationService(repo ports.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

// Notify создаёт уведомление и шлёт realtime-сигнал. Best-effort: ошибка БД не
// должна валить вызывающую бизнес-операцию — логируем и продолжаем.
func (s *notificationService) Notify(userID uuid.UUID, typ, title, body, entityType string, entityID *uuid.UUID) {
	if userID == uuid.Nil {
		return
	}
	n := &models.Notification{
		ID:         uuid.New(),
		UserID:     userID,
		Type:       typ,
		Title:      title,
		Body:       body,
		EntityType: entityType,
		EntityID:   entityID,
		Read:       false,
		CreatedAt:  time.Now(),
	}
	if err := s.repo.Create(n); err != nil {
		log.Printf("notify: create failed (non-fatal): %v", err)
		return
	}
	// Сигнал клиенту: «у тебя что-то изменилось — дозапроси свой счётчик/список».
	publishGlobal(StreamEvent{Type: "notification", WorkItemID: userID.String()})
}

func (s *notificationService) List(userID uuid.UUID, page, pageSize int, onlyUnread bool) ([]models.Notification, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 30
	}
	return s.repo.ListByUser(userID, pageSize, (page-1)*pageSize, onlyUnread)
}

func (s *notificationService) UnreadCount(userID uuid.UUID) (int64, error) {
	return s.repo.UnreadCount(userID)
}

func (s *notificationService) MarkRead(id, userID uuid.UUID) error {
	ok, err := s.repo.MarkRead(id, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("notification not found")
	}
	return nil
}

func (s *notificationService) MarkAllRead(userID uuid.UUID) error {
	return s.repo.MarkAllRead(userID)
}
