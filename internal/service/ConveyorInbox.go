package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	models "emplacc-api/internal/domain"

	"github.com/google/uuid"
)

type AgentInboxItemResponse struct {
	ID             uuid.UUID  `json:"id"`
	WorkItemID     *uuid.UUID `json:"work_item_id,omitempty"`
	Kind           string     `json:"kind"`
	Source         string     `json:"source"`
	Title          string     `json:"title"`
	Summary        string     `json:"summary"`
	Priority       int16      `json:"priority"`
	ActionRequired bool       `json:"action_required"`
	AckState       string     `json:"ack_state"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type conveyorInboxRepository interface {
	ListAgentInboxItems(ctx context.Context, recipientID uuid.UUID) ([]models.AgentInboxItem, error)
	GetAgentInboxItem(ctx context.Context, id uuid.UUID) (*models.AgentInboxItem, error)
	UpdateAgentInboxItemAck(ctx context.Context, itemID uuid.UUID, recipientID uuid.UUID, state string, at time.Time) error
}

func (s *conveyorService) ListAgentInbox(ctx context.Context, actor ConveyorActor) ([]AgentInboxItemResponse, error) {
	if actor.ActorID == uuid.Nil {
		return nil, ErrPermissionDenied
	}
	repo, err := s.inboxRepository()
	if err != nil {
		return nil, err
	}
	items, err := repo.ListAgentInboxItems(ctx, actor.ActorID)
	if err != nil {
		return nil, err
	}
	visible := make([]AgentInboxItemResponse, 0, len(items))
	for _, item := range items {
		if item.WorkItemID != nil {
			allowed, err := s.repo.ActorCanAccessTask(ctx, actor.ActorID, *item.WorkItemID)
			if err != nil {
				return nil, err
			}
			if !allowed {
				continue
			}
		}
		visible = append(visible, agentInboxItemResponse(item))
	}
	sort.SliceStable(visible, func(left int, right int) bool {
		if visible[left].Priority != visible[right].Priority {
			return visible[left].Priority > visible[right].Priority
		}
		return visible[left].CreatedAt.Before(visible[right].CreatedAt)
	})
	return visible, nil
}

func (s *conveyorService) AckAgentInboxItem(ctx context.Context, actor ConveyorActor, itemID uuid.UUID, state string) (*AgentInboxItemResponse, error) {
	if actor.ActorID == uuid.Nil {
		return nil, ErrPermissionDenied
	}
	if itemID == uuid.Nil || !validAgentInboxAckState(state) {
		return nil, fmt.Errorf("%w: invalid inbox ack", ErrValidation)
	}
	repo, err := s.inboxRepository()
	if err != nil {
		return nil, err
	}
	item, err := repo.GetAgentInboxItem(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item.RecipientID != actor.ActorID {
		return nil, ErrPermissionDenied
	}
	if item.WorkItemID != nil {
		allowed, err := s.repo.ActorCanAccessTask(ctx, actor.ActorID, *item.WorkItemID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, ErrPermissionDenied
		}
	}
	now := time.Now()
	if err := repo.UpdateAgentInboxItemAck(ctx, item.ID, actor.ActorID, state, now); err != nil {
		return nil, err
	}
	item.AckState = state
	item.UpdatedAt = now
	response := agentInboxItemResponse(*item)
	return &response, nil
}

func (s *conveyorService) inboxRepository() (conveyorInboxRepository, error) {
	repo, ok := s.repo.(conveyorInboxRepository)
	if !ok {
		return nil, fmt.Errorf("%w: inbox repository is not configured", ErrValidation)
	}
	return repo, nil
}

func validAgentInboxAckState(state string) bool {
	switch state {
	case models.AgentInboxAckStateDelivered, models.AgentInboxAckStateRead, models.AgentInboxAckStateHandled:
		return true
	default:
		return false
	}
}

func agentInboxItemResponse(item models.AgentInboxItem) AgentInboxItemResponse {
	return AgentInboxItemResponse{
		ID:             item.ID,
		WorkItemID:     item.WorkItemID,
		Kind:           item.Kind,
		Source:         item.Source,
		Title:          item.Title,
		Summary:        item.Summary,
		Priority:       item.Priority,
		ActionRequired: item.ActionRequired,
		AckState:       item.AckState,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}
