package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type authService struct {
	repo     ports.AuthRepository
	userRepo ports.UserRepository
	cfg      config.KeycloakConfig
	config   *config.PaginationConfig
}

func NewAuthService(
	cfg config.KeycloakConfig,
	repository ports.AuthRepository,
	userRepo ports.UserRepository,
	config *config.PaginationConfig,
) ports.AuthService {
	return &authService{
		repo:     repository,
		userRepo: userRepo,
		cfg:      cfg,
		config:   config,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (*ports.AuthLoginResult, error) {
	tokens, err := s.repo.Login(ctx, strings.TrimSpace(email), password)
	if err != nil {
		return nil, err
	}

	info, err := s.fetchUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		return nil, err
	}

	if err := s.ensureUserExists(ctx, info); err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(strings.TrimSpace(info.Subject))
	if err != nil {
		return nil, err
	}

	authTokens := buildAuthTokens(tokens)

	return &ports.AuthLoginResult{
		UserID: userID,
		Email:  safePointer(info.Email),
		Tokens: authTokens,
	}, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.Logout(ctx, refreshToken)
}

func (s *authService) GetUserInfo(ctx context.Context, token string) (*ports.AuthUserInfo, error) {
	info, err := s.fetchUserInfo(ctx, token)
	if err != nil {
		return nil, err
	}

	if err := s.ensureUserExists(ctx, info); err != nil {
		return nil, err
	}

	return info, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*ports.AuthRefreshResult, error) {
	tokens, err := s.repo.RefreshToken(ctx, strings.TrimSpace(refreshToken))
	if err != nil {
		return nil, err
	}

	return &ports.AuthRefreshResult{Tokens: buildAuthTokens(tokens)}, nil
}

func (s *authService) ValidateToken(ctx context.Context, token string) error {
	active, err := s.repo.Introspect(ctx, token)
	if err != nil || !active {
		if !s.cfg.TokenExchangeEnabled {
			if err != nil {
				return err
			}
			return fmt.Errorf("token inactive")
		}

		exchanged, exErr := s.repo.ExchangeToken(ctx, token)
		if exErr != nil {
			return exErr
		}

		active, err = s.repo.Introspect(ctx, exchanged.AccessToken)
		if err != nil {
			return err
		}
		if !active {
			return fmt.Errorf("token inactive")
		}

		token = exchanged.AccessToken
	}

	info, err := s.fetchUserInfo(ctx, token)
	if err != nil {
		return err
	}

	return s.ensureUserExists(ctx, info)
}

func (s *authService) fetchUserInfo(ctx context.Context, token string) (*ports.AuthUserInfo, error) {
	info, err := s.repo.UserInfo(ctx, token)
	if err == nil {
		return info, nil
	}

	if !s.cfg.TokenExchangeEnabled {
		return nil, err
	}

	exchanged, exErr := s.repo.ExchangeToken(ctx, token)
	if exErr != nil {
		return nil, exErr
	}

	info, err = s.repo.UserInfo(ctx, exchanged.AccessToken)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (s *authService) ensureUserExists(ctx context.Context, info *ports.AuthUserInfo) error {
	if info == nil || strings.TrimSpace(info.Subject) == "" {
		return fmt.Errorf("user info missing subject")
	}

	userID, err := uuid.Parse(strings.TrimSpace(info.Subject))
	if err != nil {
		return err
	}

	email := safePointer(info.Email)
	if email == "" {
		email = userID.String()
	}

	_, err = s.userRepo.GetUserByID(ctx, userID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	now := time.Now().UTC()
	user := &models.User{
		ID:            userID,
		Email:         email,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
		TgID:          "",
		TgUserID:      0,
		Profession:    "",
		EmailVerified: info.EmailVerified,
		FirstName:     safePointer(info.GivenName),
		LastName:      safePointer(info.FamilyName),
		LastLogin:     now,
		Deleted:       false,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		if !isUniqueViolation(err) {
			return err
		}
	}

	return nil
}

func buildAuthTokens(tokens *ports.AuthRepositoryTokens) ports.AuthTokens {
	if tokens == nil {
		return ports.AuthTokens{}
	}

	now := time.Now().UTC()

	return ports.AuthTokens{
		AccessToken:      tokens.AccessToken,
		RefreshToken:     tokens.RefreshToken,
		ExpiresIn:        tokens.ExpiresIn,
		RefreshExpiresIn: tokens.RefreshExpiresIn,
		TokenType:        tokens.TokenType,
		ExpiresAt:        now.Add(time.Duration(tokens.ExpiresIn) * time.Second),
	}
}

func safePointer(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint")
}

var _ ports.AuthService = (*authService)(nil)
