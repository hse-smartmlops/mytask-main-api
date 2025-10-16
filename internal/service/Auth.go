package service

import (
	"context"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/repository"
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

func (s *authService) ValidateToken(token string) error {
	ctx := context.Background()

	// Try backend client introspection
	result, err := s.keycloakClient.RetrospectToken(ctx, token, s.clientID, s.clientSecret, s.realm)
	if err != nil || !*result.Active {
		// Attempt token exchange for Flutter token
		exchanged, exErr := s.ExchangeToken(ctx, token)
		if exErr != nil {
			log.Printf("Token exchange failed: %v", exErr)
			return exErr
		}
		token = exchanged.AccessToken

		// Validate again with backend client
		result, err = s.keycloakClient.RetrospectToken(ctx, token, s.clientID, s.clientSecret, s.realm)
		if err != nil || !*result.Active {
			log.Printf("Token inactive after exchange: %v", err)
			return fmt.Errorf("token inactive")
		}
	}

	userInfo, err := s.GetUserInfo(token)
	if err != nil {
		return err
	}

	userId, err := uuid.Parse(*userInfo.Sub)
	if err != nil {
		log.Printf("Failed to parse uuid: %v", err)
		return err
	}

	// Check if user exists in database
	_, err = s.userRepo.GetUserById(userId)
	if err != nil {
		if err.Error() == "user not found" {
			// Create user in database
			temp := true
			userCreateReq := request.UserCreateRequest{
				Email:         userInfo.Email,
				FirstName:     userInfo.GivenName,
				LastName:      userInfo.FamilyName,
				IsActive:      &temp,
				EmailVerified: userInfo.EmailVerified,
			}
			err = s.userRepo.CreateUserWithID(userCreateReq, userId)
			if err != nil {
				log.Printf("Failed to create user in database: %v", err)
				return err
			}
		} else {
			log.Printf("DB error (find user by id): %v", err)
			return err
		}
	}

	return nil
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

// Добавьте в структуру authService
func (s *authService) ValidateTokenForMiddleware(token string) error {
	ctx := context.Background()

	// Try backend client introspection
	result, err := s.keycloakClient.RetrospectToken(ctx, token, s.clientID, s.clientSecret, s.realm)
	if err != nil || result == nil || !*result.Active {
		// Attempt token exchange for Flutter token
		exchanged, exErr := s.ExchangeToken(ctx, token)
		if exErr != nil {
			log.Printf("Token exchange failed: %v", exErr)
			return exErr
		}
		token = exchanged.AccessToken

		// Validate again with backend client
		result, err = s.keycloakClient.RetrospectToken(ctx, token, s.clientID, s.clientSecret, s.realm)
		if err != nil || result == nil || !*result.Active {
			log.Printf("Token inactive after exchange: %v", err)
			return fmt.Errorf("token inactive")
		}
	}

	userInfo, err := s.GetUserInfo(token)
	if err != nil {
		return err
	}

	userId, err := uuid.Parse(*userInfo.Sub)
	if err != nil {
		log.Printf("Failed to parse uuid: %v", err)
		return err
	}

	// Check if user exists in database
	_, err = s.userRepo.GetUserById(userId)
	if err != nil {
		if err.Error() == "user not found" {
			// Create user in database
			temp := true
			userCreateReq := request.UserCreateRequest{
				Email:         userInfo.Email,
				FirstName:     userInfo.GivenName,
				LastName:      userInfo.FamilyName,
				IsActive:      &temp,
				EmailVerified: userInfo.EmailVerified,
			}
			err = s.userRepo.CreateUserWithID(userCreateReq, userId)
			if err != nil {
				log.Printf("Failed to create user in database: %v", err)
				return err
			}
		} else {
			log.Printf("DB error (find user by id): %v", err)
			return err
		}
	}

	return nil
}