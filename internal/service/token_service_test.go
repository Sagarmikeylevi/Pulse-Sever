package service

import (
	"testing"

	"github.com/Sagarmikeylevi/Pulse-Sever/internal/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func newTestTokenService() TokenService {
	cfg := shared.JWTConfig{
		Secret:             "test-secret-key-for-testing-only",
		AccessTokenExpiry:  15,
		RefreshTokenExpiry: 43200,
	}
	return NewTokenService(cfg)
}

func TestGenerateAndValidateAccessToken(t *testing.T) {
	ts := newTestTokenService()

	// Arrange
	userID := uuid.New()
	claims := TokenClaims{
		UserID: userID,
		Email:  "sagar@test.com",
	}

	// Act: generate
	tokenString, err := ts.GenerateAccessToken(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Act: validate
	parsed, err := ts.ValidateAccessToken(tokenString)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, userID, parsed.UserID)
	assert.Equal(t, "sagar@test.com", parsed.Email)
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	ts := newTestTokenService()

	// Act
	claims, err := ts.ValidateAccessToken("garbage.token.here")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	// Generate with one secret
	ts1 := NewTokenService(shared.JWTConfig{
		Secret:            "secret-one",
		AccessTokenExpiry: 15,
	})
	// Validate with different secret
	ts2 := NewTokenService(shared.JWTConfig{
		Secret:            "secret-two",
		AccessTokenExpiry: 15,
	})

	tokenString, err := ts1.GenerateAccessToken(TokenClaims{
		UserID: uuid.New(),
		Email:  "sagar@test.com",
	})
	assert.NoError(t, err)

	// Act: validate with wrong secret
	claims, err := ts2.ValidateAccessToken(tokenString)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestGenerateRefreshToken(t *testing.T) {
	ts := newTestTokenService()

	// Act
	raw, hash, err := ts.GenerateRefreshToken()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, raw, 64)  // 32 bytes = 64 hex chars
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, raw, hash) // raw and hash must be different
}

func TestGenerateRefreshToken_Unique(t *testing.T) {
	ts := newTestTokenService()

	// Act: generate two tokens
	raw1, _, _ := ts.GenerateRefreshToken()
	raw2, _, _ := ts.GenerateRefreshToken()

	// Assert: they should never be the same
	assert.NotEqual(t, raw1, raw2)
}

func TestHashToken_Consistent(t *testing.T) {
	ts := newTestTokenService()

	// Act: hash the same input twice
	hash1 := ts.HashToken("same-input")
	hash2 := ts.HashToken("same-input")

	// Assert: same input = same hash
	assert.Equal(t, hash1, hash2)
}

func TestHashToken_DifferentInputs(t *testing.T) {
	ts := newTestTokenService()

	// Act
	hash1 := ts.HashToken("input-one")
	hash2 := ts.HashToken("input-two")

	// Assert: different inputs = different hashes
	assert.NotEqual(t, hash1, hash2)
}
