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
	repo     ports.NotificationRepository
	userRepo ports.UserRepository
	mailer   ports.Mailer  // optional (nil → email disabled)
	emailSem chan struct{} // ограничивает число одновременных email-горутин
}

func NewNotificationService(repo ports.NotificationRepository, userRepo ports.UserRepository, mailer ports.Mailer) NotificationService {
	return &notificationService{
		repo:     repo,
		userRepo: userRepo,
		mailer:   mailer,
		emailSem: make(chan struct{}, 8),
	}
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

	// Email — best-effort, в фоне, чтобы не блокировать бизнес-операцию на SMTP.
	// Семафор ограничивает число одновременных отправок; при перегрузе письмо
	// тихо пропускаем (in-app/SSE уже доставлены).
	if s.mailer != nil && s.userRepo != nil {
		select {
		case s.emailSem <- struct{}{}:
			go func() {
				defer func() { <-s.emailSem }()
				s.sendEmail(userID, title, body)
			}()
		default:
			log.Printf("notify email: queue full, skipping email for %s", userID)
		}
	}
}

func (s *notificationService) sendEmail(userID uuid.UUID, title, body string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("notify email: panic recovered: %v", r)
		}
	}()
	user, err := s.userRepo.GetUserById(userID)
	if err != nil || user == nil || user.Email == "" {
		return
	}
	text := body
	if text == "" {
		text = title
	}
	if err := s.mailer.Send(user.Email, title, text); err != nil {
		log.Printf("notify email: send to %s failed (non-fatal): %v", user.Email, err)
	}
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
