package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/Nerzal/gocloak/v13"
	"github.com/google/uuid"
)

type authService struct {
	client   *gocloak.GoCloak
	cfg      config.KeycloakConfig
	userRepo ports.UserRepository
}

func NewAuthService(cfg config.KeycloakConfig, userRepo ports.UserRepository) ports.AuthService {
	clientPtr := gocloak.NewClient(cfg.URL)
	return &authService{
		client:   clientPtr,
		cfg:      cfg,
		userRepo: userRepo,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (*ports.AuthLoginResult, error) {
	token, err := s.client.Login(ctx, s.cfg.ClientID, s.cfg.ClientSecret, s.cfg.Realm, email, password)
	if err != nil {
		return nil, err
	}

	userInfo, err := s.getUserInfoWithExchange(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}

	if err := s.ensureUserExists(ctx, userInfo); err != nil {
		return nil, err
	}

	userID, err := parseUUID(userInfo.Sub)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(time.Duration(token.ExpiresIn) * time.Second)

	return &ports.AuthLoginResult{
		UserID: userID,
		Email:  safeString(userInfo.Email),
		Tokens: ports.AuthTokens{
			AccessToken:      token.AccessToken,
			RefreshToken:     token.RefreshToken,
			ExpiresIn:        token.ExpiresIn,
			RefreshExpiresIn: token.RefreshExpiresIn,
			TokenType:        token.TokenType,
			ExpiresAt:        expiresAt,
		},
	}, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.client.Logout(ctx, s.cfg.ClientID, s.cfg.ClientSecret, s.cfg.Realm, refreshToken)
}

func (s *authService) GetUserInfo(ctx context.Context, token string) (*ports.AuthUserInfo, error) {
	userInfo, err := s.getUserInfoWithExchange(ctx, token)
	if err != nil {
		return nil, err
	}

	info := &ports.AuthUserInfo{
		Subject:           safeString(userInfo.Sub),
		Name:              cloneString(userInfo.Name),
		PreferredUsername: cloneString(userInfo.PreferredUsername),
		GivenName:         cloneString(userInfo.GivenName),
		FamilyName:        cloneString(userInfo.FamilyName),
		Email:             cloneString(userInfo.Email),
		EmailVerified:     userInfo.EmailVerified != nil && *userInfo.EmailVerified,
	}

	return info, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*ports.AuthRefreshResult, error) {
	token, err := s.client.RefreshToken(ctx, refreshToken, s.cfg.ClientID, s.cfg.ClientSecret, s.cfg.Realm)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(time.Duration(token.ExpiresIn) * time.Second)

	return &ports.AuthRefreshResult{
		Tokens: ports.AuthTokens{
			AccessToken:      token.AccessToken,
			RefreshToken:     token.RefreshToken,
			ExpiresIn:        token.ExpiresIn,
			RefreshExpiresIn: token.RefreshExpiresIn,
			TokenType:        token.TokenType,
			ExpiresAt:        expiresAt,
		},
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, token string) error {
	result, err := s.retrospectToken(ctx, token)
	if err != nil {
		return err
	}
	if result == nil || result.Active == nil || !*result.Active {
		return fmt.Errorf("token inactive")
	}

	userInfo, err := s.getUserInfoWithExchange(ctx, token)
	if err != nil {
		return err
	}

	return s.ensureUserExists(ctx, userInfo)
}

func (s *authService) getUserInfoWithExchange(ctx context.Context, token string) (*gocloak.UserInfo, error) {
	userInfo, err := s.client.GetUserInfo(ctx, token, s.cfg.Realm)
	if err == nil {
		return userInfo, nil
	}

	if !s.cfg.TokenExchangeEnabled {
		return nil, err
	}

	exchanged, exErr := s.exchangeToken(ctx, token)
	if exErr != nil {
		return nil, exErr
	}

	return s.client.GetUserInfo(ctx, exchanged.AccessToken, s.cfg.Realm)
}

func (s *authService) ensureUserExists(ctx context.Context, info *gocloak.UserInfo) error {
	if info == nil || info.Sub == nil {
		return fmt.Errorf("user info missing subject")
	}

	userID, err := uuid.Parse(safeString(info.Sub))
	if err != nil {
		return err
	}

	email := safeString(info.Email)
	if email == "" {
		email = userID.String()
	}

	_, err = s.userRepo.GetUserByID(ctx, userID)
	if err == nil {
		return nil
	}
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
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
		EmailVerified: info.EmailVerified != nil && *info.EmailVerified,
		FirstName:     safeString(info.GivenName),
		LastName:      safeString(info.FamilyName),
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

func (s *authService) retrospectToken(ctx context.Context, token string) (*gocloak.RetrospectResult, error) {
	clientID := s.cfg.ClientID
	clientSecret := s.cfg.ClientSecret
	if s.cfg.BackendClientID != "" {
		clientID = s.cfg.BackendClientID
	}
	if s.cfg.BackendClientSecret != "" {
		clientSecret = s.cfg.BackendClientSecret
	}

	result, err := s.client.RetrospectToken(ctx, token, clientID, clientSecret, s.cfg.Realm)
	if err == nil && result != nil && result.Active != nil && *result.Active {
		return result, nil
	}

	if !s.cfg.TokenExchangeEnabled {
		return result, err
	}

	exchanged, exErr := s.exchangeToken(ctx, token)
	if exErr != nil {
		return nil, exErr
	}

	return s.client.RetrospectToken(ctx, exchanged.AccessToken, clientID, clientSecret, s.cfg.Realm)
}

func (s *authService) exchangeToken(ctx context.Context, subjectToken string) (*tokenExchangeResponse, error) {
	endpoint := strings.TrimRight(s.cfg.URL, "/") + "/realms/" + s.cfg.Realm + "/protocol/openid-connect/token"

	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	data.Set("subject_token", subjectToken)
	data.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
	data.Set("client_id", s.cfg.ClientID)
	data.Set("client_secret", s.cfg.ClientSecret)
	data.Set("scope", "openid")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("token exchange failed: status=%d body=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("token exchange failed")
	}

	var token tokenExchangeResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func parseUUID(value *string) (uuid.UUID, error) {
	if value == nil {
		return uuid.Nil, fmt.Errorf("missing uuid value")
	}
	return uuid.Parse(strings.TrimSpace(*value))
}

func safeString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

type tokenExchangeResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint")
}

var _ ports.AuthService = (*authService)(nil)
