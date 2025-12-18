package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sheltonFr/bootdev/chirspy/internal/database"
	"github.com/sheltonFr/bootdev/chirspy/internal/utils"
)

type userHandler struct {
	db *database.Queries
}

type user struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func mapUser(dbUser *database.User) user {
	return user{
		ID:        dbUser.ID,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
}

func NewUserHandler(db *database.Queries) *userHandler {
	return &userHandler{db}
}

type createUserDto struct {
	Email string `json:"email"`
}

func (u *userHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userDto createUserDto
	err := json.NewDecoder(r.Body).Decode(&userDto)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Could not parse Json")
		return
	}

	user, err := u.db.CreateUser(r.Context(), userDto.Email)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}
	utils.RespondWithJSON(w, http.StatusCreated, mapUser(&user))

}
