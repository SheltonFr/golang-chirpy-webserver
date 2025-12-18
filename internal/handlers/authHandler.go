package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/sheltonFr/bootdev/chirspy/internal/auth"
	"github.com/sheltonFr/bootdev/chirspy/internal/database"
	"github.com/sheltonFr/bootdev/chirspy/internal/mappers"
	"github.com/sheltonFr/bootdev/chirspy/internal/utils"
)

type authHandler struct {
	db     *database.Queries
	logger *log.Logger
}

func NewAuthHandler(db *database.Queries, logger *log.Logger) *authHandler {
	return &authHandler{db, logger}
}

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
	utils.RespondWithJSON(w, http.StatusOK, mappers.MapUser(&user))
}
