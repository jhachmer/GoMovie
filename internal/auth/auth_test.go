package auth

import (
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jhachmer/gomovie/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateToken(t *testing.T) {
	os.Setenv("gomovie_JWT", "mysecret")
	config.Envs.JWT_key = "mysecret"
	defer os.Unsetenv("gomovie_JWT")

	t.Run("ValidToken", func(t *testing.T) {
		username := "testuser"
		token, err := CreateToken(username)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !parsedToken.Valid {
			t.Fatalf("expected token to be valid")
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			t.Fatalf("expected claims to be of type jwt.MapClaims")
		}

		if claims["sub"] != username {
			t.Fatalf("expected username to be %s, got %s", username, claims["sub"])
		}
	})
}

func TestVerifyToken(t *testing.T) {
	os.Setenv("gomovie_JWT", "mysecret")
	config.Envs.JWT_key = "mysecret"
	defer os.Unsetenv("gomovie_JWT")

	t.Run("ValidToken", func(t *testing.T) {
		username := "testuser"
		token, err := CreateToken(username)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		verifiedToken, err := VerifyToken(token)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !verifiedToken.Valid {
			t.Fatalf("expected token to be valid")
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		invalidToken := "invalidtoken"
		_, err := VerifyToken(invalidToken)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		expectedError := "error parsing jwt"
		if err.Error()[:len(expectedError)] != expectedError {
			t.Fatalf("expected error message to start with %s, got %s", expectedError, err.Error())
		}
	})
}

func TestHashPassword(t *testing.T) {
	t.Run("ValidPassword", func(t *testing.T) {
		password := "mypassword"
		hashedPassword, err := HashPassword(password)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
