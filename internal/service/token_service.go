package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/shared"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenClaims struct {
	UserID uuid.UUID
	Email  string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type TokenService interface {
	GenerateAccessToken(claims TokenClaims) (string, error)
	ValidateAccessToken(tokenString string) (*TokenClaims, error)
	GenerateRefreshToken() (raw string, hash string, err error)
	HashToken(raw string) string
}

type tokenService struct {
	cfg shared.JWTConfig
}

func NewTokenService(cfg shared.JWTConfig) TokenService {
	return &tokenService{cfg: cfg}
}

func (s *tokenService) GenerateAccessToken(claims TokenClaims) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   claims.UserID.String(),
		"email": claims.Email,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Duration(s.cfg.AccessTokenExpiry) * time.Minute).Unix(),
	})

	return token.SignedString([]byte(s.cfg.Secret))
}

func (s *tokenService) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	userID, err := uuid.Parse(mapClaims["sub"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	return &TokenClaims{
		UserID: userID,
		Email:  mapClaims["email"].(string),
	}, nil
}

func (s *tokenService) GenerateRefreshToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	raw := hex.EncodeToString(bytes)
	hash := s.HashToken(raw)

	return raw, hash, nil
}

func (s *tokenService) HashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
