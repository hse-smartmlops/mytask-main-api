package service

import (
	"context"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/repository"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/google/uuid"
)

type AuthService interface {
	Login(email, password string) (*AuthResponse, error)
	Logout(token string) error
	GetUserInfo(token string) (*gocloak.UserInfo, error)
	RefreshToken(refreshToken string) (*TokenResponse, error)
	ValidateToken(token string) error
	ValidateTokenForMiddleware(token string) error // Новый метод для middleware
}

type authService struct {
	keycloakClient gocloak.GoCloak
	realm          string
	clientID       string
	clientSecret   string
	userRepo       repository.UserRepository
}

type AuthResponse struct {
	UserID       *string `json:"user_id"`
	Email        *string `json:"email"`
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresIn    int     `json:"expires_in"`
	RefreshExp   int     `json:"refresh_expires_in"`
	TokenType    string  `json:"token_type"`
	ExpiresAt    string  `json:"expires_at"`
}

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope,omitempty"`
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		keycloakClient: *gocloak.NewClient(os.Getenv("KEYCLOAK_URL")),
		realm:          os.Getenv("KEYCLOAK_REALM"),
		clientID:       os.Getenv("KEYCLOAK_CLIENT_ID"),
		clientSecret:   os.Getenv("KEYCLOAK_CLIENT_SECRET"),
		userRepo:       userRepo,
	}
}

func (s *authService) Login(email, password string) (*AuthResponse, error) {
	ctx := context.Background()
	token, err := s.keycloakClient.Login(ctx, s.clientID, s.clientSecret, s.realm, email, password)
	if err != nil {
		return nil, err
	}

	// получаем userInfo
	userInfo, err := s.keycloakClient.GetUserInfo(ctx, token.AccessToken, s.realm)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Format(time.RFC3339)

	return &AuthResponse{
		UserID:       userInfo.Sub,
		Email:        userInfo.Email,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
		RefreshExp:   token.RefreshExpiresIn,
		TokenType:    token.TokenType,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *authService) Logout(token string) error {
	ctx := context.Background()
	err := s.keycloakClient.Logout(ctx, s.clientID, s.clientSecret, s.realm, token)
	return err
}

func (s *authService) GetUserInfo(token string) (*gocloak.UserInfo, error) {
	ctx := context.Background()
	var userInfo *gocloak.UserInfo
	var err error

	// Try to fetch user info directly
	userInfo, err = s.keycloakClient.GetUserInfo(ctx, token, s.realm)
	if err != nil {
		// Attempt token exchange for Flutter token
		exchanged, exErr := s.ExchangeToken(ctx, token)
		if exErr != nil {
			log.Printf("Me token exchange failed: %v / original err: %v", exErr, err)
			return nil, exErr
		}
		token = exchanged.AccessToken
		userInfo, err = s.keycloakClient.GetUserInfo(ctx, token, s.realm)
		if err != nil {
			log.Printf("Me failed after exchange: %v", err)
			return nil, err
		}
	}

	return userInfo, nil
}

func (s *authService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	ctx := context.Background()
	token, err := s.keycloakClient.RefreshToken(
		ctx,
		refreshToken,
		s.clientID,
		s.clientSecret,
		s.realm,
	)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		ExpiresIn:        token.ExpiresIn,
		RefreshExpiresIn: token.RefreshExpiresIn,
		TokenType:        token.TokenType,
	}, nil
}

// parseJWTClaims разбирает payload JWT для извлечения полей.
// ВАЖНО: не проверяет подпись — вызывающий код обязан сначала вызвать verifyJWTSignature
// (как это делает ensureUserFromJWT).
func parseJWTClaims(token string) (sub, email, firstName, lastName string, exp int64, emailVerified bool, err error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		err = fmt.Errorf("malformed JWT")
		return
	}
	payload, e := base64.RawURLEncoding.DecodeString(parts[1])
	if e != nil {
		err = e
		return
	}
	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Exp           int64  `json:"exp"`
		EmailVerified bool   `json:"email_verified"`
	}
	if e := json.Unmarshal(payload, &claims); e != nil {
		err = e
		return
	}
	if claims.Sub == "" {
		err = fmt.Errorf("missing sub claim")
		return
	}
	if time.Now().Unix() > claims.Exp {
		err = fmt.Errorf("token expired")
		return
	}
	sub, email, firstName, lastName, exp, emailVerified = claims.Sub, claims.Email, claims.GivenName, claims.FamilyName, claims.Exp, claims.EmailVerified
	return
}

func (s *authService) ValidateToken(token string) error {
	return s.ensureUserFromJWT(token)
}

func (s *authService) ExchangeToken(ctx context.Context, subjectToken string) (*TokenResponse, error) {
    endpoint := strings.TrimRight(os.Getenv("KEYCLOAK_URL"), "/") +
        "/realms/" + os.Getenv("KEYCLOAK_REALM") + "/protocol/openid-connect/token"

    data := url.Values{}
    data.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
    data.Set("subject_token", subjectToken)
    data.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
    data.Set("client_id", s.clientID)
    data.Set("client_secret", s.clientSecret)
    data.Set("scope", "openid") // Основной scope для OpenID Connect

    log.Printf("Exchanging token with scope: openid")

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
    if err != nil {
        log.Printf("Error creating request: %v", err)
        return nil, err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Printf("Error making request: %v", err)
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading response: %v", err)
    }

    if resp.StatusCode != http.StatusOK {
        log.Printf("Token exchange failed with status %d: %s", resp.StatusCode, body)
        return nil, fmt.Errorf("token exchange failed: %s", body)
    }

    var tokenResp TokenResponse
    if err := json.Unmarshal(body, &tokenResp); err != nil {
        return nil, fmt.Errorf("error parsing token response: %v", err)
    }

    return &tokenResp, nil
}

func (s *authService) ValidateTokenForMiddleware(token string) error {
	return s.ensureUserFromJWT(token)
}

// verifyJWTSignature проверяет подпись токена по JWKS нашего Keycloak realm.
// gocloak кэширует сертификаты, так что в горячем пути это не сетевой вызов на каждый запрос.
// Без этой проверки любой может подделать payload (sub/email) и выдать себя за другого
// пользователя, включая админа — поэтому подпись обязательна.
func (s *authService) verifyJWTSignature(token string) error {
	ctx := context.Background()
	decoded, _, err := s.keycloakClient.DecodeAccessToken(ctx, token, s.realm)
	if err != nil {
		return fmt.Errorf("token signature invalid: %w", err)
	}
	if decoded == nil || !decoded.Valid {
		return fmt.Errorf("token invalid")
	}
	return nil
}

// ensureUserFromJWT проверяет подпись JWT по Keycloak realm, затем разбирает payload
// и создаёт пользователя в БД если его нет.
// Если JWT валидный — всегда возвращает nil по DB-ошибкам (не блокируем запрос из-за БД),
// но невалидная подпись/claims всегда отклоняются.
func (s *authService) ensureUserFromJWT(token string) error {
	if err := s.verifyJWTSignature(token); err != nil {
		log.Printf("JWT signature verification failed: %v", err)
		return fmt.Errorf("token invalid")
	}

	sub, email, firstName, lastName, _, emailVerified, err := parseJWTClaims(token)
	if err != nil {
		log.Printf("JWT parse failed: %v", err)
		return fmt.Errorf("token invalid")
	}

	userId, err := uuid.Parse(sub)
	if err != nil {
		return fmt.Errorf("invalid sub in token")
	}

	// Пробуем найти или создать пользователя, но не блокируем запрос при DB-ошибке
	_, dbErr := s.userRepo.GetUserById(userId)
	if dbErr != nil {
		if dbErr.Error() == "user not found" {
			isActive := true
			ev := emailVerified
			userCreateReq := request.UserCreateRequest{
				Email:         &email,
				FirstName:     &firstName,
				LastName:      &lastName,
				IsActive:      &isActive,
				EmailVerified: &ev,
			}
			if createErr := s.userRepo.CreateUserWithID(userCreateReq, userId); createErr != nil {
				log.Printf("Failed to create user (non-fatal): %v", createErr)
				// Не возвращаем ошибку — JWT валидный, запрос разрешаем
			}
		} else {
			log.Printf("DB error in ensureUserFromJWT (non-fatal): %v", dbErr)
			// Не возвращаем ошибку — JWT валидный, запрос разрешаем
		}
	}
	return nil
}