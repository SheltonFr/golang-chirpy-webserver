package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/sheltonFr/bootdev/chirspy/internal/auth"
	"github.com/sheltonFr/bootdev/chirspy/internal/database"
	"github.com/sheltonFr/bootdev/chirspy/internal/mappers"
	"github.com/sheltonFr/bootdev/chirspy/internal/utils"
)

type authHandler struct {
	db        *database.Queries
	logger    *log.Logger
	jwtSecret string
}

func NewAuthHandler(db *database.Queries, logger *log.Logger, jwtSecret string) *authHandler {
	return &authHandler{db, logger, jwtSecret}
}

type LoginDTO struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds uint   `json:"expires_in_seconds"`
}

func (a *authHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var loginDTO LoginDTO
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&loginDTO)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	user, err := a.db.GetUserByEmail(r.Context(), loginDTO.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		a.logger.Printf("DB error: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "An error occured")
		return
	}

	match, _ := auth.CheckPasswordHash(loginDTO.Password, user.HashedPassword)
	if !match {
		utils.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	expiry := time.Hour
	if loginDTO.ExpiresInSeconds > 0 {
		expiry = time.Duration(loginDTO.ExpiresInSeconds) * time.Second
	}
	token, err := auth.MakeJWT(user.ID, a.jwtSecret, expiry)
	if err != nil {
		a.logger.Fatalf("Jwt error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "An error occured")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, mappers.MapUserLogin(&user, token))
}
