package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sheltonFr/bootdev/chirspy/internal/auth"
	"github.com/sheltonFr/bootdev/chirspy/internal/database"
	"github.com/sheltonFr/bootdev/chirspy/internal/mappers"
	"github.com/sheltonFr/bootdev/chirspy/internal/utils"
)

type userHandler struct {
	db     *database.Queries
	logger *log.Logger
}

func NewUserHandler(db *database.Queries, logger *log.Logger) *userHandler {
	return &userHandler{db, logger}
}

type createUserDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *userHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userDto createUserDto
	err := json.NewDecoder(r.Body).Decode(&userDto)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Could not parse Json")
		return
	}

	passwordHash, err := auth.HashPassword(userDto.Password)
	if err != nil {
		u.logger.Printf("Hashing error: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Error hashing password")
		return
	}

	userParams := database.CreateUserParams{
		Email:          userDto.Email,
		HashedPassword: passwordHash,
	}

	user, err := u.db.CreateUser(r.Context(), userParams)

	if err != nil {
		u.logger.Printf("DB error: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}
	utils.RespondWithJSON(w, http.StatusCreated, mappers.MapUser(&user))

}
