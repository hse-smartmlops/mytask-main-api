package presenter

import (
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app/ports"
)

func ToAuthResponse(result *ports.AuthLoginResult) response.Auth {
	if result == nil {
		return response.Auth{}
	}

	return response.Auth{
		UserID:       result.UserID.String(),
		Email:        result.Email,
		AccessToken:  result.Tokens.AccessToken,
		RefreshToken: result.Tokens.RefreshToken,
		ExpiresIn:    result.Tokens.ExpiresIn,
		RefreshExp:   result.Tokens.RefreshExpiresIn,
		TokenType:    result.Tokens.TokenType,
		ExpiresAt:    result.Tokens.ExpiresAt,
	}
}

func ToRefreshResponse(result *ports.AuthRefreshResult) response.Refresh {
	if result == nil {
		return response.Refresh{}
	}

	return response.Refresh{
		AccessToken:  result.Tokens.AccessToken,
		RefreshToken: result.Tokens.RefreshToken,
		ExpiresIn:    result.Tokens.ExpiresIn,
		RefreshExp:   result.Tokens.RefreshExpiresIn,
		TokenType:    result.Tokens.TokenType,
		ExpiresAt:    result.Tokens.ExpiresAt,
	}
}

func ToUserInfoResponse(info *ports.AuthUserInfo) response.UserInfo {
	if info == nil {
		return response.UserInfo{}
	}

	return response.UserInfo{
		Sub:               info.Subject,
		Name:              info.Name,
		PreferredUsername: info.PreferredUsername,
		GivenName:         info.GivenName,
		FamilyName:        info.FamilyName,
		Email:             info.Email,
		EmailVerified:     info.EmailVerified,
	}
}
