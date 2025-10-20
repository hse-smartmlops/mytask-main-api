package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
)

type KeycloakRepository struct {
	cfg           config.KeycloakConfig
	httpClient    *http.Client
	tokenURL      string
	logoutURL     string
	userinfoURL   string
	introspectURL string
}

func NewKeycloakRepository(cfg config.KeycloakConfig) *KeycloakRepository {
	base := strings.TrimRight(cfg.URL, "/") + "/realms/" + cfg.Realm + "/protocol/openid-connect"
	client := &http.Client{Timeout: 15 * time.Second}

	return &KeycloakRepository{
		cfg:           cfg,
		httpClient:    client,
		tokenURL:      base + "/token",
		logoutURL:     base + "/logout",
		userinfoURL:   base + "/userinfo",
		introspectURL: base + "/token/introspect",
	}
}

func (r *KeycloakRepository) Login(ctx context.Context, email, password string) (*ports.AuthRepositoryTokens, error) {
	payload := url.Values{}
	payload.Set("grant_type", "password")
	payload.Set("client_id", r.cfg.ClientID)
	if r.cfg.ClientSecret != "" {
		payload.Set("client_secret", r.cfg.ClientSecret)
	}
	payload.Set("username", email)
	payload.Set("password", password)
	payload.Set("scope", "openid")

	var resp tokenResponse
	if err := r.postForm(ctx, r.tokenURL, payload, &resp); err != nil {
		return nil, err
	}

	return resp.toTokens(), nil
}

func (r *KeycloakRepository) RefreshToken(ctx context.Context, refreshToken string) (*ports.AuthRepositoryTokens, error) {
	payload := url.Values{}
	payload.Set("grant_type", "refresh_token")
	payload.Set("refresh_token", refreshToken)
	payload.Set("client_id", r.cfg.ClientID)
	if r.cfg.ClientSecret != "" {
		payload.Set("client_secret", r.cfg.ClientSecret)
	}

	var resp tokenResponse
	if err := r.postForm(ctx, r.tokenURL, payload, &resp); err != nil {
		return nil, err
	}

	return resp.toTokens(), nil
}

func (r *KeycloakRepository) Logout(ctx context.Context, refreshToken string) error {
	payload := url.Values{}
	payload.Set("client_id", r.cfg.ClientID)
	if r.cfg.ClientSecret != "" {
		payload.Set("client_secret", r.cfg.ClientSecret)
	}
	payload.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.logoutURL, strings.NewReader(payload.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("keycloak logout failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	return nil
}

func (r *KeycloakRepository) UserInfo(ctx context.Context, token string) (*ports.AuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.userinfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("keycloak userinfo failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var info userInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return info.toAuthUserInfo(), nil
}

func (r *KeycloakRepository) Introspect(ctx context.Context, token string) (bool, error) {
	payload := url.Values{}
	payload.Set("token", token)

	clientID := r.cfg.ClientID
	clientSecret := r.cfg.ClientSecret
	if r.cfg.BackendClientID != "" {
		clientID = r.cfg.BackendClientID
	}
	if r.cfg.BackendClientSecret != "" {
		clientSecret = r.cfg.BackendClientSecret
	}

	payload.Set("client_id", clientID)
	if clientSecret != "" {
		payload.Set("client_secret", clientSecret)
	}

	var resp introspectResponse
	if err := r.postForm(ctx, r.introspectURL, payload, &resp); err != nil {
		return false, err
	}

	return resp.Active, nil
}

func (r *KeycloakRepository) ExchangeToken(ctx context.Context, subjectToken string) (*ports.AuthRepositoryTokens, error) {
	payload := url.Values{}
	payload.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	payload.Set("subject_token", subjectToken)
	payload.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
	payload.Set("client_id", r.cfg.ClientID)
	if r.cfg.ClientSecret != "" {
		payload.Set("client_secret", r.cfg.ClientSecret)
	}
	payload.Set("scope", "openid")

	var resp tokenResponse
	if err := r.postForm(ctx, r.tokenURL, payload, &resp); err != nil {
		return nil, err
	}

	return resp.toTokens(), nil
}

func (r *KeycloakRepository) postForm(ctx context.Context, endpoint string, payload url.Values, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(payload.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("keycloak request failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	if out == nil {
		return nil
	}

	if err := json.Unmarshal(body, out); err != nil {
		return err
	}

	return nil
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
}

func (t tokenResponse) toTokens() *ports.AuthRepositoryTokens {
	return &ports.AuthRepositoryTokens{
		AccessToken:      t.AccessToken,
		RefreshToken:     t.RefreshToken,
		ExpiresIn:        t.ExpiresIn,
		RefreshExpiresIn: t.RefreshExpiresIn,
		TokenType:        t.TokenType,
	}
}

type userInfoResponse struct {
	Sub               string  `json:"sub"`
	Name              *string `json:"name"`
	PreferredUsername *string `json:"preferred_username"`
	GivenName         *string `json:"given_name"`
	FamilyName        *string `json:"family_name"`
	Email             *string `json:"email"`
	EmailVerified     *bool   `json:"email_verified"`
}

func (u userInfoResponse) toAuthUserInfo() *ports.AuthUserInfo {
	verified := false
	if u.EmailVerified != nil {
		verified = *u.EmailVerified
	}

	return &ports.AuthUserInfo{
		Subject:           strings.TrimSpace(u.Sub),
		Name:              trimPointer(u.Name),
		PreferredUsername: trimPointer(u.PreferredUsername),
		GivenName:         trimPointer(u.GivenName),
		FamilyName:        trimPointer(u.FamilyName),
		Email:             trimPointer(u.Email),
		EmailVerified:     verified,
	}
}

type introspectResponse struct {
	Active bool `json:"active"`
}

func trimPointer(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

var _ ports.AuthRepository = (*KeycloakRepository)(nil)
