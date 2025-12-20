package mappers

import (
	"time"

	"github.com/google/uuid"
	"github.com/sheltonFr/bootdev/chirspy/internal/database"
)

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserLoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	UserResponse
}

func MapUserLogin(dbUser *database.User, acessToken, refreshToken string) UserLoginResponse {
	return UserLoginResponse{
		Token:        acessToken,
		RefreshToken: refreshToken,
		UserResponse: UserResponse{ID: dbUser.ID,
			Email:     dbUser.Email,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
		},
	}
}

func MapUser(dbUser *database.User) UserResponse {
	return UserResponse{
		ID:        dbUser.ID,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
}
