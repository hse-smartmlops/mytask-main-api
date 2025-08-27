package controller

import (
	"context"
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"strings"

	"github.com/Nerzal/gocloak/v13"
	"github.com/labstack/echo/v4"
)

func RegisterAuthRoutes(e *echo.Echo) {
	authGroup := e.Group("/auth")
	authGroup.POST("/login", Login)
	authGroup.POST("/logout", Logout)
	authGroup.POST("/register", Register)
	authGroup.GET("/me", Me)
	authGroup.POST("/oauth", OAuth)
	authGroup.POST("/totp", TOTP)
	authGroup.GET("/validate", ValidateToken)
	authGroup.POST("/refresh", RefreshToken)
}

var (
	keycloakClient = gocloak.NewClient(os.Getenv("KEYCLOAK_URL"))
	realm          = os.Getenv("KEYCLOAK_REALM")
	clientID       = os.Getenv("KEYCLOAK_CLIENT_ID")
	clientSecret   = os.Getenv("KEYCLOAK_CLIENT_SECRET")
)

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope,omitempty"`
}

func ExchangeToken(ctx context.Context, subjectToken string) (*TokenResponse, error) {
	endpoint := strings.TrimRight(os.Getenv("KEYCLOAK_URL"), "/") +
		"/realms/" + os.Getenv("KEYCLOAK_REALM") + "/protocol/openid-connect/token"

	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	data.Set("subject_token", subjectToken)
	data.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token") // <-- REQUIRED
	data.Set("client_id", clientID)       // backend client
	data.Set("client_secret", clientSecret)

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

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", body)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// Login godoc
// @Summary Аутентификация пользователя
// @Description Аутентифицирует пользователя по email и паролю через Keycloak
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body request.LoginRequest true "Данные для входа"
// @Success 200 {object} response.AuthResponse "Успешная аутентификация"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 401 {object} map[string]string "Неверные учетные данные"
// @Router /auth/login [post]
func Login(c echo.Context) error {
	var req request.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	ctx := context.Background()
	token, err := keycloakClient.Login(ctx, clientID, clientSecret, realm, req.Email, req.Password)
	if err != nil {
		log.Printf("Authorization error: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	// получаем userInfo
	userInfo, err := keycloakClient.GetUserInfo(ctx, token.AccessToken, realm)
	if err != nil {
		log.Printf("Authorization error: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "cannot fetch userinfo"})
	}

	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Format(time.RFC3339)

	resp := response.AuthResponse{
		UserID:       strPtrToVal(userInfo.Sub),
		Email:        strPtrToVal(userInfo.Email),
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
		RefreshExp:   token.RefreshExpiresIn,
		TokenType:    token.TokenType,
		ExpiresAt:    expiresAt,
	}

	return c.JSON(http.StatusOK, resp)
}

// OAuth godoc
// @Summary Аутентификация через OAuth
// @Description Выполняет аутентификацию с использованием OAuth кода авторизации
// @Tags Auth
// @Accept json
// @Produce json
// @Param oauth body request.OAuthRequest true "Данные для OAuth аутентификации"
// @Success 200 {object} response.AuthResponse "Успешная OAuth аутентификация"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 401 {object} map[string]string "Ошибка OAuth аутентификации"
// @Router /auth/oauth [post]
func OAuth(c echo.Context) error {
	var req request.OAuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	ctx := context.Background()
	token, err := keycloakClient.GetToken(ctx, realm, gocloak.TokenOptions{
		ClientID:     gocloak.StringP(clientID),
		ClientSecret: gocloak.StringP(clientSecret),
		GrantType:    gocloak.StringP("authorization_code"),
		Code:         gocloak.StringP(req.Code),
		RedirectURI:  gocloak.StringP(req.RedirectURI),
	})
	if err != nil {
		log.Printf("Authorization error: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "oauth failed"})
	}

	userInfo, err := keycloakClient.GetUserInfo(ctx, token.AccessToken, realm)
	if err != nil {
		log.Printf("Authorization error: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "cannot fetch userinfo"})
	}

	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Format(time.RFC3339)

	resp := response.AuthResponse{
		UserID:       strPtrToVal(userInfo.Sub),
		Email:        strPtrToVal(userInfo.Email),
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
		RefreshExp:   token.RefreshExpiresIn,
		TokenType:    token.TokenType,
		ExpiresAt:    expiresAt,
	}

	return c.JSON(http.StatusOK, resp)
}

// TOTP godoc
// @Summary Аутентификация с двухфакторной аутентификацией
// @Description Выполняет аутентификацию с использованием email, пароля и TOTP кода
// @Tags Auth
// @Accept json
// @Produce json
// @Param totp body request.TOTPRequest true "Данные для TOTP аутентификации"
// @Success 200 {object} response.AuthResponse "Успешная TOTP аутентификация"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 401 {object} map[string]string "Неверные учетные данные или TOTP код"
// @Router /auth/totp [post]
func TOTP(c echo.Context) error {
	var req request.TOTPRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("TOTP error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	ctx := context.Background()
	token, err := keycloakClient.GetToken(ctx, realm, gocloak.TokenOptions{
		ClientID:     gocloak.StringP(clientID),
		ClientSecret: gocloak.StringP(clientSecret),
		GrantType:    gocloak.StringP("password"),
		Username:     gocloak.StringP(req.Email),
		Password:     gocloak.StringP(req.Password),
		Totp:         gocloak.StringP(req.TOTP),
	})
	if err != nil {
		log.Printf("TOTP error: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials or otp"})
	}

	userInfo, err := keycloakClient.GetUserInfo(ctx, token.AccessToken, realm)
	if err != nil {
		log.Printf("TOTP error: %v", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "cannot fetch userinfo"})
	}

	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Format(time.RFC3339)

	resp := response.AuthResponse{
		UserID:       strPtrToVal(userInfo.Sub),
		Email:        strPtrToVal(userInfo.Email),
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.ExpiresIn,
		RefreshExp:   token.RefreshExpiresIn,
		TokenType:    token.TokenType,
		ExpiresAt:    expiresAt,
	}

	return c.JSON(http.StatusOK, resp)
}

// Logout godoc
// @Summary Выход из системы
// @Description Выполняет выход пользователя из системы, завершая сессию в Keycloak
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Refresh token"
// @Success 200 {object} map[string]string "Успешный выход из системы"
// @Failure 400 {object} map[string]string "Отсутствует токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при выходе"
// @Router /auth/logout [post]
func Logout(c echo.Context) error {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing token"})
	}
	log.Print(auth)
	ctx := context.Background()
	err := keycloakClient.Logout(ctx, clientID, clientSecret, realm, auth)
	if err != nil {
		log.Printf("Logout error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "logout failed"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "logout successful"})
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя в Keycloak с указанным email
// @Tags Auth
// @Accept json
// @Produce json
// @Param register body request.RegisterRequest true "Данные для регистрации"
// @Success 200 {object} map[string]string "Пользователь успешно зарегистрирован"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании пользователя"
// @Router /auth/register [post]
func Register(c echo.Context) error {
	var req request.RegisterRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Register error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	ctx := context.Background()
	adminToken := getAdminToken()

	// Создаём пользователя
	user := gocloak.User{
		Username: gocloak.StringP(req.Email),
		Email:    gocloak.StringP(req.Email),
		Enabled:  gocloak.BoolP(true),
		FirstName: gocloak.StringP(req.FirstName),
		LastName:  gocloak.StringP(req.LastName),
	}
	userID, err := keycloakClient.CreateUser(ctx, adminToken, realm, user)
	if err != nil {
		log.Printf("Register error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
	}

	// Устанавливаем пароль
	err = keycloakClient.SetPassword(ctx, adminToken, userID, realm, req.Password, false)
	if err != nil {
		log.Printf("Register error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to set password"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "user registered"})
}

// Me godoc
// @Summary Получение информации о текущем пользователе
// @Description Возвращает информацию о пользователе на основе переданного токена
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token, например: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
// @Success 200 {object} response.UserInfo "Информация о пользователе"
// @Failure 401 {object} map[string]string "Отсутствует или неверный токен"
// @Router /auth/me [get]
func Me(c echo.Context) error {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
	}

	token := auth
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[len("Bearer "):])
	}

	ctx := context.Background()
	var userInfo *gocloak.UserInfo
	var err error

	// Try to fetch user info directly
	userInfo, err = keycloakClient.GetUserInfo(ctx, token, realm)
	if err != nil {
		// Attempt token exchange for Flutter token
		exchanged, exErr := ExchangeToken(ctx, token)
		if exErr != nil {
			log.Printf("Me token exchange failed: %v / original err: %v", exErr, err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}
		token = exchanged.AccessToken
		userInfo, err = keycloakClient.GetUserInfo(ctx, token, realm)
		if err != nil {
			log.Printf("Me failed after exchange: %v", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}
	}

	resp := response.UserInfo{
		Sub:               strPtrToVal(userInfo.Sub),
		Name:              strPtrToVal(userInfo.Name),
		PreferredUsername: strPtrToVal(userInfo.PreferredUsername),
		GivenName:         strPtrToVal(userInfo.GivenName),
		FamilyName:        strPtrToVal(userInfo.FamilyName),
		Email:             strPtrToVal(userInfo.Email),
		EmailVerified:     userInfo.EmailVerified != nil && *userInfo.EmailVerified,
	}

	return c.JSON(http.StatusOK, resp)
}

// RefreshToken godoc
// @Summary Обновление access токена
// @Description Получение нового access_token и refresh_token на основе существующего refresh_token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body request.RefreshRequest true "Refresh token"
// @Success 200 {object} response.RefreshResponse "Новые токены и время жизни"
// @Failure 400 {object} map[string]string "Неверный формат запроса"
// @Failure 401 {object} map[string]string "Неверный или истёкший refresh_token"
// @Router /auth/refresh [post]
func RefreshToken(c echo.Context) error {
    var req request.RefreshRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }

    ctx := context.Background()
    token, err := keycloakClient.RefreshToken(
        ctx,
        req.RefreshToken,
        os.Getenv("KEYCLOAK_CLIENT_ID"),
        os.Getenv("KEYCLOAK_CLIENT_SECRET"),
        realm,
    )
    if err != nil {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
    }

	now := time.Now()

    newTokens := response.RefreshResponse{
        AccessToken:  token.AccessToken,
        RefreshToken: token.RefreshToken,
        ExpiresIn:    300,
        RefreshExp:   1800,
        TokenType:    "Bearer",
        ExpiresAt:    now.Add(time.Duration(300) * time.Second),
    }

    return c.JSON(http.StatusOK, newTokens)
}

// ValidateToken godoc
// @Summary Проверка access_token
// @Description Проверяет валидность токена через Keycloak
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token, например: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
// @Success 200 {object} response.TokenValidationResponse "Токен валиден"
// @Failure 401 {object} response.TokenValidationResponse "Невалидный или отсутствующий токен"
// @Router /auth/validate [get]
func ValidateToken(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid Authorization header"})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	ctx := context.Background()

	// Try backend client introspection
	result, err := keycloakClient.RetrospectToken(ctx, token, clientID, clientSecret, realm)
	if err != nil || !*result.Active {
		// Attempt token exchange for Flutter token
		exchanged, exErr := ExchangeToken(ctx, token)
		if exErr != nil {
			log.Printf("Token exchange failed: %v", exErr)
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token invalid"})
		}
		token = exchanged.AccessToken

		// Validate again with backend client
		result, err = keycloakClient.RetrospectToken(ctx, token, clientID, clientSecret, realm)
		if err != nil || !*result.Active {
			log.Printf("Token inactive after exchange: %v", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token inactive"})
		}
	}

	var userInfo *gocloak.UserInfo

	// Try to fetch user info directly
	userInfo, err = keycloakClient.GetUserInfo(ctx, token, realm)
	if err != nil {
		// Attempt token exchange for Flutter token
		exchanged, exErr := ExchangeToken(ctx, token)
		if exErr != nil {
			log.Printf("Me token exchange failed: %v / original err: %v", exErr, err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}
		token = exchanged.AccessToken
		userInfo, err = keycloakClient.GetUserInfo(ctx, token, realm)
		if err != nil {
			log.Printf("Me failed after exchange: %v", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}
	}

	userId, err := uuid.Parse(strPtrToVal(userInfo.Sub))
	if err != nil{
		log.Printf("Failed to parse uuid: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "invalid user_id"})
	}

	var user models.User

	dbResult := dbConn.Session(&gorm.Session{}).First(&user, "email = ?", strPtrToVal(userInfo.PreferredUsername))
	if dbResult.Error != nil {
		if errors.Is(dbResult.Error, gorm.ErrRecordNotFound) {
			temp := true
			now := time.Now()
			userCreateReq := request.UserCreateRequest{
				Email: userInfo.Email,
				FirstName: userInfo.GivenName,
				LastName: userInfo.FamilyName,
				IsActive: &temp,
				LastLogin: &now,
				CreatedAt: &now,
			}
			err = CreateUserWithIdFunc(userCreateReq, c, userId)
			if err != nil{
				log.Printf("Failed to create user in database: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "failed to create user",
				})
			}
		}else{
			log.Printf("DB error (find user by email): %v", dbResult.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении пользователя из базы данных",
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Token is valid"})
}	

func getAdminToken() string {
	ctx := context.Background()
	token, err := keycloakClient.LoginAdmin(ctx,
		os.Getenv("KEYCLOAK_ADMIN"),
		os.Getenv("KEYCLOAK_ADMIN_PASSWORD"),
		"master")
	if err != nil {
		panic(err)
	}
	return token.AccessToken
}

func strPtrToVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}