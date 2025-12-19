package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

var secret = "my-super-secret-key"
var userID = uuid.New()
var expiry = time.Hour

func TestJWTFlow(t *testing.T) {
	t.Run("Valid Token", func(t *testing.T) {
		token, err := MakeJWT(userID, secret, expiry)
		if err != nil {
			t.Fatalf("Failed to make JWT: %v\n", err)
		}

		validatedID, err := ValidateJWT(token, secret)
		if err != nil {
			t.Fatalf("Validation Failed: %v\n", err)
		}

		if validatedID != userID {
			t.Fatalf("Expected UUID %v, got %v", userID, validatedID)
		}
	})

	t.Run("Wrong Secret", func(t *testing.T) {
		token, _ := MakeJWT(userID, secret, expiry)
		_, err := ValidateJWT(token, "random-secret")
		if err == nil {
			t.Fatalf("Expected error when usgin wrong secret, but got nil")
		}
	})

	t.Run("Expired Token", func(t *testing.T) {
		token, _ := MakeJWT(userID, secret, -time.Hour)
		_, err := ValidateJWT(token, "random-secret")
		if err == nil {
			t.Fatalf("Expected error for expired token, but got nil")
		}
	})
}

func TestGetBearerToken(t *testing.T) {
	expectedToken, _ := MakeJWT(userID, secret, expiry)
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+expectedToken)

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("Did expect an error, but got one: %v", err)
	}

	if token != expectedToken {
		t.Fatalf("Expected %s, but got %s ", expectedToken, token)
	}
}
